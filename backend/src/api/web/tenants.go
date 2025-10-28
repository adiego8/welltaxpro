package webapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"welltaxpro/src/internal/crypto"
	"welltaxpro/src/internal/middleware"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// getAllTenants returns all tenant connections (admin only)
func (api *API) getAllTenants(w http.ResponseWriter, r *http.Request) {
	logger.Info("Getting all tenants")

	query := `
		SELECT id, tenant_id, tenant_name, db_host, db_port, db_user,
		       db_name, db_sslmode, schema_prefix, adapter_type,
		       COALESCE(storage_provider, ''), COALESCE(storage_bucket, ''),
		       COALESCE(docusign_integration_key, ''), COALESCE(docusign_client_id, ''),
		       COALESCE(docusign_api_url, ''),
		       is_active, created_at, updated_at, created_by, notes
		FROM tenant_connections
		ORDER BY created_at DESC
	`

	rows, err := api.store.DB.Query(query)
	if err != nil {
		logger.Errorf("Failed to query tenants: %v", err)
		http.Error(w, "Failed to fetch tenants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tenants []types.TenantConnection
	for rows.Next() {
		var tc types.TenantConnection
		err := rows.Scan(
			&tc.ID,
			&tc.TenantID,
			&tc.TenantName,
			&tc.DBHost,
			&tc.DBPort,
			&tc.DBUser,
			&tc.DBName,
			&tc.DBSslMode,
			&tc.SchemaPrefix,
			&tc.AdapterType,
			&tc.StorageProvider,
			&tc.StorageBucket,
			&tc.DocuSignIntegrationKey,
			&tc.DocuSignClientID,
			&tc.DocuSignAPIURL,
			&tc.IsActive,
			&tc.CreatedAt,
			&tc.UpdatedAt,
			&tc.CreatedBy,
			&tc.Notes,
		)
		if err != nil {
			logger.Errorf("Failed to scan tenant: %v", err)
			continue
		}
		tenants = append(tenants, tc)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tenants); err != nil {
		logger.Errorf("Failed to encode tenants: %v", err)
	}
}

// getTenant returns a single tenant by ID (admin only)
func (api *API) getTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	logger.Infof("Getting tenant: %s", tenantID)

	tc, err := api.store.GetTenantConfig(tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Tenant not found", http.StatusNotFound)
		} else {
			logger.Errorf("Failed to get tenant: %v", err)
			http.Error(w, "Failed to fetch tenant", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tc); err != nil {
		logger.Errorf("Failed to encode tenant: %v", err)
	}
}

// createTenant creates a new tenant connection (admin only)
func (api *API) createTenant(w http.ResponseWriter, r *http.Request) {
	// Get employee from context
	employee, ok := middleware.GetEmployeeFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		TenantID                 string  `json:"tenantId"`
		TenantName               string  `json:"tenantName"`
		DBHost                   string  `json:"dbHost"`
		DBPort                   int     `json:"dbPort"`
		DBUser                   string  `json:"dbUser"`
		DBPassword               string  `json:"dbPassword"`
		DBName                   string  `json:"dbName"`
		DBSslMode                string  `json:"dbSslMode"`
		SchemaPrefix             string  `json:"schemaPrefix"`
		AdapterType              string  `json:"adapterType"`
		StorageProvider          string  `json:"storageProvider"`
		StorageBucket            string  `json:"storageBucket"`
		StorageCredentialsSecret string  `json:"storageCredentialsSecret"`
		StorageCredentialsPath   string  `json:"storageCredentialsPath"`
		DocuSignIntegrationKey   string  `json:"docusignIntegrationKey"`
		DocuSignClientID         string  `json:"docusignClientId"`
		DocuSignPrivateKeySecret string  `json:"docusignPrivateKeySecret"`
		DocuSignAPIURL           string  `json:"docusignApiUrl"`
		Notes                    *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.TenantID == "" || req.TenantName == "" || req.DBHost == "" ||
		req.DBUser == "" || req.DBPassword == "" || req.DBName == "" ||
		req.SchemaPrefix == "" || req.AdapterType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.DBPort == 0 {
		req.DBPort = 5432
	}
	if req.DBSslMode == "" {
		req.DBSslMode = "require"
	}
	if req.DocuSignAPIURL == "" {
		req.DocuSignAPIURL = "https://demo.docusign.net/restapi"
	}

	// Encrypt password before storing
	encryptedPassword, err := crypto.EncryptPassword(req.DBPassword)
	if err != nil {
		logger.Errorf("Failed to encrypt password: %v", err)
		http.Error(w, "Failed to encrypt credentials", http.StatusInternalServerError)
		return
	}

	// Insert tenant connection
	query := `
		INSERT INTO tenant_connections (
			tenant_id, tenant_name, db_host, db_port, db_user, db_password,
			db_name, db_sslmode, schema_prefix, adapter_type,
			storage_provider, storage_bucket, storage_credentials_secret, storage_credentials_path,
			docusign_integration_key, docusign_client_id, docusign_private_key_secret, docusign_api_url,
			created_by, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		) RETURNING id, created_at, updated_at
	`

	var tenantID uuid.UUID
	var createdAt, updatedAt string
	err = api.store.DB.QueryRow(
		query,
		req.TenantID,
		req.TenantName,
		req.DBHost,
		req.DBPort,
		req.DBUser,
		encryptedPassword,
		req.DBName,
		req.DBSslMode,
		req.SchemaPrefix,
		req.AdapterType,
		nullIfEmpty(req.StorageProvider),
		nullIfEmpty(req.StorageBucket),
		nullIfEmpty(req.StorageCredentialsSecret),
		nullIfEmpty(req.StorageCredentialsPath),
		nullIfEmpty(req.DocuSignIntegrationKey),
		nullIfEmpty(req.DocuSignClientID),
		nullIfEmpty(req.DocuSignPrivateKeySecret),
		req.DocuSignAPIURL,
		employee.Email,
		req.Notes,
	).Scan(&tenantID, &createdAt, &updatedAt)

	if err != nil {
		logger.Errorf("Failed to create tenant: %v", err)
		http.Error(w, "Failed to create tenant", http.StatusInternalServerError)
		return
	}

	logger.Infof("Created tenant %s (ID: %s) by %s", req.TenantID, tenantID, employee.Email)

	// Automatically grant the creating employee admin access to this tenant
	_, err = api.store.DB.Exec(`
		INSERT INTO employee_tenant_access (employee_id, tenant_id, role, created_by)
		VALUES ($1, $2, 'admin', $3)
	`, employee.ID, req.TenantID, employee.ID)

	if err != nil {
		logger.Errorf("Failed to grant tenant access to employee: %v", err)
		// Don't fail the whole operation, just log the error
		// The admin can manually assign access later
	} else {
		logger.Infof("Granted admin access to tenant %s for employee %s", req.TenantID, employee.Email)
	}

	response := map[string]interface{}{
		"id":        tenantID,
		"tenantId":  req.TenantID,
		"createdAt": createdAt,
		"updatedAt": updatedAt,
		"message":   "Tenant created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}

// updateTenant updates an existing tenant connection (admin only)
func (api *API) updateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	var req struct {
		TenantName               string  `json:"tenantName"`
		DBHost                   string  `json:"dbHost"`
		DBPort                   int     `json:"dbPort"`
		DBUser                   string  `json:"dbUser"`
		DBPassword               *string `json:"dbPassword"` // Optional - only update if provided
		DBName                   string  `json:"dbName"`
		DBSslMode                string  `json:"dbSslMode"`
		SchemaPrefix             string  `json:"schemaPrefix"`
		AdapterType              string  `json:"adapterType"`
		StorageProvider          string  `json:"storageProvider"`
		StorageBucket            string  `json:"storageBucket"`
		StorageCredentialsSecret string  `json:"storageCredentialsSecret"`
		StorageCredentialsPath   string  `json:"storageCredentialsPath"`
		DocuSignIntegrationKey   string  `json:"docusignIntegrationKey"`
		DocuSignClientID         string  `json:"docusignClientId"`
		DocuSignPrivateKeySecret string  `json:"docusignPrivateKeySecret"`
		DocuSignAPIURL           string  `json:"docusignApiUrl"`
		IsActive                 *bool   `json:"isActive"`
		Notes                    *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build update query dynamically based on provided fields
	query := `UPDATE tenant_connections SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.TenantName != "" {
		query += `, tenant_name = $` + formatArgIdx(argIdx)
		args = append(args, req.TenantName)
		argIdx++
	}
	if req.DBHost != "" {
		query += `, db_host = $` + formatArgIdx(argIdx)
		args = append(args, req.DBHost)
		argIdx++
	}
	if req.DBPort != 0 {
		query += `, db_port = $` + formatArgIdx(argIdx)
		args = append(args, req.DBPort)
		argIdx++
	}
	if req.DBUser != "" {
		query += `, db_user = $` + formatArgIdx(argIdx)
		args = append(args, req.DBUser)
		argIdx++
	}
	if req.DBPassword != nil && *req.DBPassword != "" {
		// Encrypt new password
		encryptedPassword, err := crypto.EncryptPassword(*req.DBPassword)
		if err != nil {
			logger.Errorf("Failed to encrypt password: %v", err)
			http.Error(w, "Failed to encrypt credentials", http.StatusInternalServerError)
			return
		}
		query += `, db_password = $` + formatArgIdx(argIdx)
		args = append(args, encryptedPassword)
		argIdx++
	}
	if req.DBName != "" {
		query += `, db_name = $` + formatArgIdx(argIdx)
		args = append(args, req.DBName)
		argIdx++
	}
	if req.DBSslMode != "" {
		query += `, db_sslmode = $` + formatArgIdx(argIdx)
		args = append(args, req.DBSslMode)
		argIdx++
	}
	if req.SchemaPrefix != "" {
		query += `, schema_prefix = $` + formatArgIdx(argIdx)
		args = append(args, req.SchemaPrefix)
		argIdx++
	}
	if req.AdapterType != "" {
		query += `, adapter_type = $` + formatArgIdx(argIdx)
		args = append(args, req.AdapterType)
		argIdx++
	}
	if req.StorageProvider != "" {
		query += `, storage_provider = $` + formatArgIdx(argIdx)
		args = append(args, nullIfEmpty(req.StorageProvider))
		argIdx++
	}
	if req.StorageBucket != "" {
		query += `, storage_bucket = $` + formatArgIdx(argIdx)
		args = append(args, nullIfEmpty(req.StorageBucket))
		argIdx++
	}
	if req.StorageCredentialsSecret != "" {
		query += `, storage_credentials_secret = $` + formatArgIdx(argIdx)
		args = append(args, nullIfEmpty(req.StorageCredentialsSecret))
		argIdx++
	}
	if req.StorageCredentialsPath != "" {
		query += `, storage_credentials_path = $` + formatArgIdx(argIdx)
		args = append(args, nullIfEmpty(req.StorageCredentialsPath))
		argIdx++
	}
	if req.DocuSignIntegrationKey != "" {
		query += `, docusign_integration_key = $` + formatArgIdx(argIdx)
		args = append(args, nullIfEmpty(req.DocuSignIntegrationKey))
		argIdx++
	}
	if req.DocuSignClientID != "" {
		query += `, docusign_client_id = $` + formatArgIdx(argIdx)
		args = append(args, nullIfEmpty(req.DocuSignClientID))
		argIdx++
	}
	if req.DocuSignPrivateKeySecret != "" {
		query += `, docusign_private_key_secret = $` + formatArgIdx(argIdx)
		args = append(args, nullIfEmpty(req.DocuSignPrivateKeySecret))
		argIdx++
	}
	if req.DocuSignAPIURL != "" {
		query += `, docusign_api_url = $` + formatArgIdx(argIdx)
		args = append(args, req.DocuSignAPIURL)
		argIdx++
	}
	if req.IsActive != nil {
		query += `, is_active = $` + formatArgIdx(argIdx)
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.Notes != nil {
		query += `, notes = $` + formatArgIdx(argIdx)
		args = append(args, req.Notes)
		argIdx++
	}

	query += ` WHERE tenant_id = $` + formatArgIdx(argIdx)
	args = append(args, tenantID)

	result, err := api.store.DB.Exec(query, args...)
	if err != nil {
		logger.Errorf("Failed to update tenant: %v", err)
		http.Error(w, "Failed to update tenant", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	logger.Infof("Updated tenant: %s", tenantID)

	response := map[string]interface{}{
		"message":  "Tenant updated successfully",
		"tenantId": tenantID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}

// deleteTenant deactivates a tenant (admin only)
func (api *API) deleteTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	logger.Infof("Deactivating tenant: %s", tenantID)

	query := `UPDATE tenant_connections SET is_active = false, updated_at = NOW() WHERE tenant_id = $1`
	result, err := api.store.DB.Exec(query, tenantID)
	if err != nil {
		logger.Errorf("Failed to deactivate tenant: %v", err)
		http.Error(w, "Failed to deactivate tenant", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	logger.Infof("Deactivated tenant: %s", tenantID)

	response := map[string]interface{}{
		"message":  "Tenant deactivated successfully",
		"tenantId": tenantID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}

// Helper functions

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func formatArgIdx(idx int) string {
	return fmt.Sprintf("%d", idx)
}
