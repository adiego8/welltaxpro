package store

import (
	"fmt"
	"welltaxpro/src/internal/adapter"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

// GetClients retrieves all clients for a specific tenant using the appropriate adapter
func (s *Store) GetClients(tenantID string) ([]*types.Client, error) {
	logger.Infof("[Store.GetClients] Step 1: Getting tenant DB connection - TenantID: %s", tenantID)

	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		logger.Errorf("[Store.GetClients] FAILED at Step 1 - TenantID: %s, Error: %v", tenantID, err)
		return nil, err
	}

	logger.Infof("[Store.GetClients] Step 2: Creating adapter - TenantID: %s, AdapterType: %s, SchemaPrefix: %s, DBHost: %s",
		tenantID, tc.AdapterType, tc.SchemaPrefix, tc.DBHost)

	// Get the appropriate adapter for this tenant
	clientAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("[Store.GetClients] FAILED at Step 2 - TenantID: %s, AdapterType: %s, Error: %v",
			tenantID, tc.AdapterType, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("[Store.GetClients] Step 3: Fetching clients from adapter - TenantID: %s", tenantID)

	// Use adapter to fetch clients
	clients, err := clientAdapter.GetClients(db, tc.SchemaPrefix)
	if err != nil {
		logger.Errorf("[Store.GetClients] FAILED at Step 3 - TenantID: %s, Error: %v", tenantID, err)
		return nil, err
	}

	logger.Infof("[Store.GetClients] SUCCESS - TenantID: %s, ClientsFetched: %d", tenantID, len(clients))
	return clients, nil
}

// GetClientByID retrieves a specific client by ID for a tenant using the appropriate adapter
func (s *Store) GetClientByID(tenantID string, clientID string) (*types.Client, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	clientAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch client
	return clientAdapter.GetClientByID(db, tc.SchemaPrefix, clientID)
}

// GetClientComprehensive retrieves all data for a client including filings, dependents, etc.
func (s *Store) GetClientComprehensive(tenantID string, clientID string) (*types.ClientComprehensive, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	clientAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter to fetch comprehensive data for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch comprehensive client data
	return clientAdapter.GetClientComprehensive(db, tc.SchemaPrefix, clientID)
}

// GetClientsByFilings retrieves clients with their filings (paginated)
func (s *Store) GetClientsByFilings(tenantID string, limit int, offset int) ([]*types.ClientComprehensive, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	clientAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter to fetch clients by filings for tenant %s (limit: %d, offset: %d)", tc.AdapterType, tenantID, limit, offset)

	// Use adapter to fetch clients with filings (paginated)
	return clientAdapter.GetClientsByFilings(db, tc.SchemaPrefix, limit, offset)
}
