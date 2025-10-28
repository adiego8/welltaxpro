package store

import (
	"database/sql"
	"fmt"
	"time"
	"welltaxpro/src/internal/crypto"
	"welltaxpro/src/internal/types"

	"github.com/Masterminds/squirrel"
	"github.com/google/logger"
)

// GetTenantConnection retrieves tenant connection details from welltaxpro database
func (s *Store) getTenantConnection(tenantID string) (*types.TenantConnection, error) {
	// query := `
	// 	SELECT id, tenant_id, tenant_name, db_host, db_port, db_user,
	// 	       db_password, db_name, db_sslmode, schema_prefix, adapter_type,
	// 	       COALESCE(storage_provider, 'gcs'), COALESCE(storage_bucket, ''),
	// 	       COALESCE(storage_credentials_secret, ''), COALESCE(storage_credentials_path, ''),
	// 	       COALESCE(docusign_integration_key, ''), COALESCE(docusign_client_id, ''),
	// 	       COALESCE(docusign_private_key_secret, ''), COALESCE(docusign_api_url, ''),
	// 	       is_active, created_at, updated_at, created_by, notes
	// 	FROM tenant_connections
	// 	WHERE tenant_id = $1 AND is_active = true
	// `
	query, args, err := squirrel.Select(
		"id",
		"tenant_id",
		"tenant_name",
		"db_host",
		"db_port",
		"db_user",
		"db_password",
		"db_name",
		"db_sslmode",
		"schema_prefix",
		"adapter_type",
		"COALESCE(storage_provider, 'gcs')",
		"COALESCE(storage_bucket, '')",
		"COALESCE(storage_credentials_secret, '')",
		"COALESCE(storage_credentials_path, '')",
		"COALESCE(docusign_integration_key, '')",
		"COALESCE(docusign_client_id, '')",
		"COALESCE(docusign_private_key_secret, '')",
		"COALESCE(docusign_api_url, '')",
		"is_active",
		"created_at",
		"updated_at",
		"created_by",
		"notes",
	).From("tenant_connections").
		Where(squirrel.Eq{"tenant_id": tenantID}).
		Where(squirrel.Eq{"is_active": true}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build SQL query for tenant %s: %v", tenantID, err)
		return nil, err
	}

	row := s.DB.QueryRow(query, args...)

	tc := &types.TenantConnection{}
	err = row.Scan(
		&tc.ID,
		&tc.TenantID,
		&tc.TenantName,
		&tc.DBHost,
		&tc.DBPort,
		&tc.DBUser,
		&tc.DBPassword,
		&tc.DBName,
		&tc.DBSslMode,
		&tc.SchemaPrefix,
		&tc.AdapterType,
		&tc.StorageProvider,
		&tc.StorageBucket,
		&tc.StorageCredentialsSecret,
		&tc.StorageCredentialsPath,
		&tc.DocuSignIntegrationKey,
		&tc.DocuSignClientID,
		&tc.DocuSignPrivateKeySecret,
		&tc.DocuSignAPIURL,
		&tc.IsActive,
		&tc.CreatedAt,
		&tc.UpdatedAt,
		&tc.CreatedBy,
		&tc.Notes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		logger.Errorf("Failed to get tenant connection for %s: %v", tenantID, err)
		return nil, err
	}

	// Decrypt password if it's encrypted
	if crypto.IsEncryptedPassword(tc.DBPassword) {
		decrypted, err := crypto.DecryptPassword(tc.DBPassword)
		if err != nil {
			logger.Errorf("Failed to decrypt password for tenant %s: %v", tenantID, err)
			return nil, fmt.Errorf("failed to decrypt tenant password: %w", err)
		}
		tc.DBPassword = decrypted
	}

	return tc, nil
}

// GetTenantConfig is an alias for GetTenantConnection for clarity
func (s *Store) GetTenantConfig(tenantID string) (*types.TenantConnection, error) {
	return s.getTenantConnection(tenantID)
}

// GetTenantDB gets or creates a database connection for a tenant
func (s *Store) GetTenantDB(tenantID string) (*sql.DB, *types.TenantConnection, error) {
	// Check if connection already exists
	s.tenantConnsMutex.RLock()
	if conn, exists := s.tenantConns[tenantID]; exists {
		s.tenantConnsMutex.RUnlock()

		// Update last access time
		s.tenantConnsMutex.Lock()
		conn.lastAccess = time.Now()
		s.tenantConnsMutex.Unlock()

		// Get tenant config for schema info
		tc, err := s.getTenantConnection(tenantID)
		if err != nil {
			return nil, nil, err
		}
		return conn.db, tc, nil
	}
	s.tenantConnsMutex.RUnlock()

	// Get tenant connection details
	tc, err := s.getTenantConnection(tenantID)
	if err != nil {
		return nil, nil, err
	}

	// Create new connection
	s.tenantConnsMutex.Lock()
	defer s.tenantConnsMutex.Unlock()

	// Double-check if connection was created while waiting for lock
	if conn, exists := s.tenantConns[tenantID]; exists {
		conn.lastAccess = time.Now()
		return conn.db, tc, nil
	}

	// Open database connection
	connStr := tc.GetConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Errorf("Failed to open connection to tenant %s: %v", tenantID, err)
		return nil, nil, fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		logger.Errorf("Failed to ping tenant %s database: %v", tenantID, err)
		return nil, nil, fmt.Errorf("failed to ping tenant database: %w", err)
	}

	// Store connection with current timestamp
	s.tenantConns[tenantID] = &tenantConnection{
		db:         db,
		lastAccess: time.Now(),
	}
	logger.Infof("Successfully connected to tenant %s database", tenantID)

	return db, tc, nil
}
