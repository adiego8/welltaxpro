package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"welltaxpro/src/internal/crypto"
	"welltaxpro/src/internal/middleware"
	"welltaxpro/src/internal/notification"
	"welltaxpro/src/internal/portal"
	"welltaxpro/src/internal/storage"

	"github.com/google/logger"
	"github.com/gorilla/mux"
)

// sendPortalLink sends a magic link email to a client for portal access (admin only)
func (api *API) sendPortalLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	clientID := vars["clientId"]

	logger.Infof("Sending portal link for client %s in tenant %s", clientID, tenantID)

	// Get tenant database connection
	tenantDB, tc, err := api.store.GetTenantDB(tenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant database: %v", err)
		http.Error(w, "Failed to connect to tenant database", http.StatusInternalServerError)
		return
	}

	// Get client information
	var clientEmail, clientFirstName, clientLastName string
	clientQuery := `
		SELECT email, COALESCE(first_name, ''), COALESCE(last_name, '')
		FROM ` + tc.SchemaPrefix + `.user
		WHERE id = $1
	`

	err = tenantDB.QueryRow(clientQuery, clientID).Scan(&clientEmail, &clientFirstName, &clientLastName)
	if err != nil {
		logger.Errorf("Failed to get client info: %v", err)
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Generate magic link token (15 min expiry, one-time use)
	token, tokenID, err := portal.GenerateMagicLinkToken(clientID, tenantID, clientEmail, api.portalJWTSecret)
	if err != nil {
		logger.Errorf("Failed to generate magic link token: %v", err)
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	// Store token in database for one-time use validation
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = api.store.DB.Exec(`
		INSERT INTO portal_magic_tokens (id, client_id, tenant_id, email, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, tokenID, clientID, tenantID, clientEmail, expiresAt)
	if err != nil {
		logger.Errorf("Failed to store magic link token: %v", err)
		http.Error(w, "Failed to create access token", http.StatusInternalServerError)
		return
	}

	// Build portal URL
	portalURL := fmt.Sprintf("%s/%s/portal?token=%s", api.portalBaseURL, tenantID, token)

	// Prepare client name
	clientName := clientFirstName
	if clientLastName != "" {
		clientName = fmt.Sprintf("%s %s", clientFirstName, clientLastName)
	}
	if clientName == "" {
		clientName = "Valued Client"
	}

	// Generate email content
	subject, htmlBody, textBody := notification.GeneratePortalAccessEmail(notification.PortalAccessEmail{
		ClientName: clientName,
		TenantName: tc.TenantName,
		PortalURL:  portalURL,
	})

	// Send email
	err = api.emailService.SendEmail(clientEmail, clientName, subject, htmlBody, textBody)
	if err != nil {
		logger.Errorf("Failed to send portal link email to %s: %v", clientEmail, err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	logger.Infof("Portal link sent successfully to %s", clientEmail)

	// Return success response
	response := map[string]interface{}{
		"message": "Portal access link sent successfully",
		"email":   clientEmail,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}

// validatePortalToken validates a portal access token and returns client info
func (api *API) validatePortalToken(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter", http.StatusBadRequest)
		return
	}

	// Validate token
	claims, err := portal.ValidatePortalToken(token, api.portalJWTSecret)
	if err != nil {
		logger.Errorf("Invalid portal token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	logger.Infof("Valid portal token for client %s in tenant %s", claims.ClientID, claims.TenantID)

	// Return client info
	response := map[string]interface{}{
		"clientId": claims.ClientID,
		"tenantId": claims.TenantID,
		"email":    claims.Email,
		"valid":    true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}

// exchangeMagicToken exchanges a magic link token for a session token after SSN verification
func (api *API) exchangeMagicToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MagicToken string `json:"magicToken"`
		LastFourSSN string `json:"lastFourSSN"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate magic token
	claims, err := portal.ValidatePortalToken(req.MagicToken, api.portalJWTSecret)
	if err != nil {
		logger.Errorf("Invalid magic token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Verify token type is magic_link
	if claims.TokenType != "magic_link" {
		logger.Warningf("Attempted to exchange non-magic-link token: %s", claims.TokenType)
		http.Error(w, "Invalid token type", http.StatusUnauthorized)
		return
	}

	// Check if token has already been used
	var used bool
	var usedAt *time.Time
	err = api.store.DB.QueryRow(`
		SELECT used, used_at FROM portal_magic_tokens WHERE id = $1
	`, claims.ID).Scan(&used, &usedAt)

	if err != nil {
		logger.Errorf("Token not found in database: %s", claims.ID)
		http.Error(w, "Token not found", http.StatusUnauthorized)
		return
	}

	if used {
		logger.Warningf("Attempted to reuse magic token (ID: %s, used at: %v)", claims.ID, usedAt)
		http.Error(w, "This link has already been used. Please request a new access link.", http.StatusUnauthorized)
		return
	}

	// Get tenant database to verify SSN
	tenantDB, tc, err := api.store.GetTenantDB(claims.TenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant database: %v", err)
		http.Error(w, "Failed to connect to tenant database", http.StatusInternalServerError)
		return
	}

	// Verify last 4 SSN against client record
	var clientSSN string
	query := `
		SELECT COALESCE(ssn, '')
		FROM ` + tc.SchemaPrefix + `.user
		WHERE id = $1
	`
	err = tenantDB.QueryRow(query, claims.ClientID).Scan(&clientSSN)
	if err != nil {
		logger.Errorf("Failed to get client SSN: %v", err)
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Decrypt SSN if encrypted
	decryptedSSN := clientSSN
	if crypto.IsEncryptedSSN(clientSSN) {
		decryptedSSN, err = crypto.DecryptSSN(clientSSN)
		if err != nil {
			logger.Errorf("Failed to decrypt SSN for client %s: %v", claims.ClientID, err)
			http.Error(w, "Identity verification not available. Please contact support.", http.StatusInternalServerError)
			return
		}
	}

	// Strip non-numeric characters from decrypted SSN (handles formats like "123-45-6789")
	numericSSN := ""
	for _, char := range decryptedSSN {
		if char >= '0' && char <= '9' {
			numericSSN += string(char)
		}
	}

	// Verify last 4 digits of SSN
	if len(numericSSN) < 4 {
		logger.Warningf("Client SSN too short or empty for client %s", claims.ClientID)
		http.Error(w, "Identity verification not available. Please contact support.", http.StatusBadRequest)
		return
	}

	storedLast4 := numericSSN[len(numericSSN)-4:]
	if storedLast4 != req.LastFourSSN {
		logger.Warningf("SSN verification failed for client %s (tenant: %s) - expected: %s, got: %s",
			claims.ClientID, claims.TenantID, storedLast4, req.LastFourSSN)
		http.Error(w, "Identity verification failed. Please check your information and try again.", http.StatusUnauthorized)
		return
	}

	logger.Infof("SSN verification successful for client %s", claims.ClientID)

	// Mark magic token as used
	_, err = api.store.DB.Exec(`
		UPDATE portal_magic_tokens
		SET used = TRUE, used_at = NOW(), ip_address = $2, user_agent = $3
		WHERE id = $1
	`, claims.ID, r.RemoteAddr, r.UserAgent())

	if err != nil {
		logger.Errorf("Failed to mark token as used: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Generate session token (2 hours)
	sessionToken, err := portal.GenerateSessionToken(claims.ClientID, claims.TenantID, claims.Email, api.portalJWTSecret)
	if err != nil {
		logger.Errorf("Failed to generate session token: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	logger.Infof("Magic token exchanged for session: client=%s tenant=%s", claims.ClientID, claims.TenantID)

	// Return session token
	response := map[string]interface{}{
		"sessionToken": sessionToken,
		"expiresIn":    7200, // 2 hours in seconds
		"clientId":     claims.ClientID,
		"tenantId":     claims.TenantID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}

// getPortalClient returns comprehensive client data for the authenticated portal user
func (api *API) getPortalClient(w http.ResponseWriter, r *http.Request) {
	// Get claims from context (set by PortalAuthMiddleware)
	claims, ok := middleware.GetPortalClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger.Infof("Getting comprehensive data for client %s in tenant %s via portal", claims.ClientID, claims.TenantID)

	// Call the same comprehensive endpoint but use clientId from JWT claims
	vars := mux.Vars(r)
	vars["clientId"] = claims.ClientID
	vars["tenantId"] = claims.TenantID

	// Set the updated vars back to the request
	r = mux.SetURLVars(r, vars)

	// Call the existing comprehensive endpoint
	api.getClientComprehensive(w, r)
}

// downloadPortalDocument allows authenticated portal users to download their own documents
func (api *API) downloadPortalDocument(w http.ResponseWriter, r *http.Request) {
	// Get claims from context (set by PortalAuthMiddleware)
	claims, ok := middleware.GetPortalClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	documentID := vars["documentId"]

	logger.Infof("Portal user %s downloading document %s", claims.ClientID, documentID)

	// Get tenant database connection
	tenantDB, tc, err := api.store.GetTenantDB(claims.TenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant database: %v", err)
		http.Error(w, "Failed to connect to tenant database", http.StatusInternalServerError)
		return
	}

	// Verify document belongs to this client
	var filePath, fileName, ownerID string
	query := `
		SELECT d.file_path, d.name, d.user_id
		FROM ` + tc.SchemaPrefix + `.document d
		WHERE d.id = $1
	`
	err = tenantDB.QueryRow(query, documentID).Scan(&filePath, &fileName, &ownerID)
	if err != nil {
		logger.Errorf("Failed to get document: %v", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Check ownership
	if ownerID != claims.ClientID {
		logger.Warningf("Client %s attempted to download document %s owned by %s", claims.ClientID, documentID, ownerID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Stream the file directly from storage
	logger.Infof("Streaming document %s to portal user %s", documentID, claims.ClientID)

	// Create storage provider
	storageProvider, err := storage.NewStorageProviderForTenant(context.Background(), tc)
	if err != nil {
		logger.Errorf("Failed to create storage provider: %v", err)
		http.Error(w, "Failed to initialize storage", http.StatusInternalServerError)
		return
	}

	// Download file from storage
	reader, err := storageProvider.Download(context.Background(), tc.StorageBucket, filePath)
	if err != nil {
		logger.Errorf("Failed to download document from storage: %v", err)
		http.Error(w, "Failed to download document", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Set response headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Stream the file to the response
	if _, err := io.Copy(w, reader); err != nil {
		logger.Errorf("Failed to stream document: %v", err)
		return
	}

	logger.Infof("Successfully streamed document %s", documentID)
}
