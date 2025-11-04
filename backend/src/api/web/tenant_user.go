package webapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"welltaxpro/src/internal/middleware"
	"welltaxpro/src/internal/storage"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// NewClientUUID is the placeholder UUID for users who don't have a client record yet
var NewClientUUID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

// autoRegisterTenantUser handles automatic tenant user registration on first sign-in
// This endpoint is called after Firebase authentication to create or retrieve tenant_user record
func (api *API) autoRegisterTenantUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	// Get Firebase UID from context (set by TenantUserAuthMiddleware)
	firebaseUID, err := middleware.GetFirebaseUIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body to get email
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	logger.Infof("Auto-registering tenant user: tenant=%s, firebaseUID=%s, email=%s", tenantID, firebaseUID, req.Email)

	// Check if tenant user already exists
	existingUser, err := api.store.GetTenantUserByFirebaseUID(firebaseUID)
	if err == nil {
		logger.Infof("Tenant user already exists: %s", existingUser.ID.String())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(existingUser)
		return
	}

	// Try to find existing client in tenant database by email
	clientID := NewClientUUID // Default to "new client"

	tenantDB, tc, err := api.store.GetTenantDB(tenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant database: %v", err)
		// Continue with NewClientUUID
	} else {
		// Query for existing client by email
		query := fmt.Sprintf(`
			SELECT id FROM %s.user
			WHERE email = $1
			LIMIT 1
		`, tc.SchemaPrefix)

		var foundClientID string
		err = tenantDB.QueryRow(query, req.Email).Scan(&foundClientID)
		if err == nil {
			// Client exists, use their ID
			parsedClientID, parseErr := uuid.Parse(foundClientID)
			if parseErr == nil {
				clientID = parsedClientID
				logger.Infof("Found existing client: %s for email: %s", clientID.String(), req.Email)
			}
		} else if err != sql.ErrNoRows {
			logger.Errorf("Error querying for client: %v", err)
		} else {
			logger.Infof("No existing client found for email: %s, using NewClientUUID", req.Email)
		}
	}

	// Create new tenant user
	tenantUser := &types.TenantUser{
		TenantID:    tenantID,
		ClientID:    clientID,
		FirebaseUID: firebaseUID,
		Email:       req.Email,
		IsActive:    true,
	}

	if err := api.store.CreateTenantUser(tenantUser); err != nil {
		logger.Errorf("Failed to create tenant user: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	logger.Infof("Successfully auto-registered tenant user: %s (client_id: %s)", tenantUser.ID.String(), clientID.String())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenantUser)
}

// registerTenantUser handles tenant user registration (requires Firebase auth)
// Admin creates the link between Firebase UID and client record
func (api *API) registerTenantUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	var req struct {
		ClientID    string `json:"clientId"`    // UUID of client in tenant database
		FirebaseUID string `json:"firebaseUid"` // Firebase UID from user's auth
		Email       string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.ClientID == "" || req.FirebaseUID == "" || req.Email == "" {
		http.Error(w, "clientId, firebaseUid, and email are required", http.StatusBadRequest)
		return
	}

	clientUUID, err := uuid.Parse(req.ClientID)
	if err != nil {
		http.Error(w, "Invalid clientId format", http.StatusBadRequest)
		return
	}

	// Create tenant user
	tenantUser := &types.TenantUser{
		TenantID:    tenantID,
		ClientID:    clientUUID,
		FirebaseUID: req.FirebaseUID,
		Email:       req.Email,
		IsActive:    true,
	}

	if err := api.store.CreateTenantUser(tenantUser); err != nil {
		logger.Errorf("Failed to create tenant user: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	logger.Infof("Registered tenant user %s for tenant %s", tenantUser.ID.String(), tenantID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenantUser)
}

// getTenantUserProfile returns the authenticated tenant user's profile and comprehensive data
func (api *API) getTenantUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get Firebase UID from context (set by TenantUserAuthMiddleware)
	firebaseUID, err := middleware.GetFirebaseUIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get tenant user record
	tenantUser, err := api.store.GetTenantUserByFirebaseUID(firebaseUID)
	if err != nil {
		logger.Errorf("Tenant user not found for firebase uid %s: %v", firebaseUID, err)
		http.Error(w, "User not registered for portal access", http.StatusNotFound)
		return
	}

	// Verify tenant ID matches URL parameter
	vars := mux.Vars(r)
	requestedTenantID := vars["tenantId"]
	if tenantUser.TenantID != requestedTenantID {
		logger.Warningf("Tenant mismatch: user belongs to %s but requested %s", tenantUser.TenantID, requestedTenantID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if this is a new user (client_id = 00000000-0000-0000-0000-000000000000)
	if tenantUser.ClientID == NewClientUUID {
		logger.Infof("Tenant user %s is a new user with no client record yet", firebaseUID)

		// Return empty client data for new users
		emptyData := map[string]interface{}{
			"client": map[string]interface{}{
				"id":        NewClientUUID.String(),
				"email":     tenantUser.Email,
				"firstName": "",
				"lastName":  "",
			},
			"filings":   []interface{}{},
			"documents": []interface{}{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emptyData)
		return
	}

	// Get comprehensive client data from tenant database
	clientData, err := api.store.GetClientComprehensive(tenantUser.TenantID, tenantUser.ClientID.String())
	if err != nil {
		logger.Errorf("Failed to get client data: %v", err)
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}

	logger.Infof("Tenant user %s accessed their profile (client: %s, tenant: %s)",
		firebaseUID, tenantUser.ClientID.String(), tenantUser.TenantID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientData)
}

// downloadTenantUserDocument allows authenticated tenant users to download their own documents
func (api *API) downloadTenantUserDocument(w http.ResponseWriter, r *http.Request) {
	// Get Firebase UID from context (set by TenantUserAuthMiddleware)
	firebaseUID, err := middleware.GetFirebaseUIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get tenant user record
	tenantUser, err := api.store.GetTenantUserByFirebaseUID(firebaseUID)
	if err != nil {
		logger.Errorf("Tenant user not found for firebase uid %s: %v", firebaseUID, err)
		http.Error(w, "User not registered for portal access", http.StatusNotFound)
		return
	}

	vars := mux.Vars(r)
	documentID := vars["documentId"]
	requestedTenantID := vars["tenantId"]

	// Verify tenant ID matches
	if tenantUser.TenantID != requestedTenantID {
		logger.Warningf("Tenant mismatch for document download: user belongs to %s but requested %s", tenantUser.TenantID, requestedTenantID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	logger.Infof("Tenant user %s downloading document %s", firebaseUID, documentID)

	// Get tenant database connection
	tenantDB, tc, err := api.store.GetTenantDB(tenantUser.TenantID)
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
	if ownerID != tenantUser.ClientID.String() {
		logger.Warningf("Client %s attempted to download document %s owned by %s",
			tenantUser.ClientID.String(), documentID, ownerID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Stream the file directly from storage
	logger.Infof("Streaming document %s to tenant user %s", documentID, tenantUser.ClientID.String())

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
