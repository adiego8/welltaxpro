package store

import (
	"fmt"
	"welltaxpro/src/internal/adapter"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

// GetClients retrieves all clients for a specific tenant using the appropriate adapter
func (s *Store) GetClients(tenantID string) ([]*types.Client, error) {
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

	// Use adapter to fetch clients
	return clientAdapter.GetClients(db, tc.SchemaPrefix)
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
