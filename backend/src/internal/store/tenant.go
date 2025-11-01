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
	logger.Infof("[GetTenantDB] Starting - TenantID: %s", tenantID)

	// Check if connection already exists
	s.tenantConnsMutex.RLock()
	if conn, exists := s.tenantConns[tenantID]; exists {
		s.tenantConnsMutex.RUnlock()
		logger.Infof("[GetTenantDB] Reusing existing connection - TenantID: %s", tenantID)

		// Update last access time
		s.tenantConnsMutex.Lock()
		conn.lastAccess = time.Now()
		s.tenantConnsMutex.Unlock()

		// Get tenant config for schema info
		tc, err := s.getTenantConnection(tenantID)
		if err != nil {
			logger.Errorf("[GetTenantDB] Failed to get tenant config - TenantID: %s, Error: %v", tenantID, err)
			return nil, nil, err
		}
		return conn.db, tc, nil
	}
	s.tenantConnsMutex.RUnlock()

	logger.Infof("[GetTenantDB] No existing connection, fetching config - TenantID: %s", tenantID)

	// Get tenant connection details
	tc, err := s.getTenantConnection(tenantID)
	if err != nil {
		logger.Errorf("[GetTenantDB] Failed to get tenant connection - TenantID: %s, Error: %v", tenantID, err)
		return nil, nil, err
	}

	logger.Infof("[GetTenantDB] Config fetched - TenantID: %s, DBHost: %s, DBPort: %d, DBName: %s, SSLMode: %s",
		tenantID, tc.DBHost, tc.DBPort, tc.DBName, tc.DBSslMode)

	// Create new connection
	s.tenantConnsMutex.Lock()
	defer s.tenantConnsMutex.Unlock()

	// Double-check if connection was created while waiting for lock
	if conn, exists := s.tenantConns[tenantID]; exists {
		logger.Infof("[GetTenantDB] Connection created while waiting for lock - TenantID: %s", tenantID)
		conn.lastAccess = time.Now()
		return conn.db, tc, nil
	}

	logger.Infof("[GetTenantDB] Opening new database connection - TenantID: %s", tenantID)

	// Open database connection (DO NOT log connection string - contains password)
	connStr := tc.GetConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Errorf("[GetTenantDB] Failed to open connection - TenantID: %s, DBHost: %s, Error: %v",
			tenantID, tc.DBHost, err)
		return nil, nil, fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Second)

	logger.Infof("[GetTenantDB] Testing connection with ping - TenantID: %s", tenantID)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		logger.Errorf("[GetTenantDB] FAILED - Ping failed - TenantID: %s, DBHost: %s, DBPort: %d, Error: %v",
			tenantID, tc.DBHost, tc.DBPort, err)
		return nil, nil, fmt.Errorf("failed to ping tenant database: %w", err)
	}

	// Store connection with current timestamp
	s.tenantConns[tenantID] = &tenantConnection{
		db:         db,
		lastAccess: time.Now(),
	}
	logger.Infof("[GetTenantDB] SUCCESS - Connection established - TenantID: %s, DBHost: %s", tenantID, tc.DBHost)

	return db, tc, nil
}
