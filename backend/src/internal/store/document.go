package store

import (
	"fmt"
	"welltaxpro/src/internal/adapter"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

// CreateDocument creates a new document record in the tenant's database
func (s *Store) CreateDocument(tenantID string, document *types.Document) (*types.Document, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	documentAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to create document
	return documentAdapter.CreateDocument(db, tc.SchemaPrefix, document)
}

// GetDocumentByID retrieves a specific document by ID
func (s *Store) GetDocumentByID(tenantID string, documentID string) (*types.Document, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	documentAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch document
	return documentAdapter.GetDocumentByID(db, tc.SchemaPrefix, documentID)
}

// GetDocumentsByFilingID retrieves all documents associated with a filing
func (s *Store) GetDocumentsByFilingID(tenantID string, filingID string) ([]*types.Document, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	documentAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch documents
	return documentAdapter.GetDocumentsByFilingID(db, tc.SchemaPrefix, filingID)
}

// DeleteDocument removes a document record from the tenant's database
func (s *Store) DeleteDocument(tenantID string, documentID string) error {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	// Get the appropriate adapter for this tenant
	documentAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to delete document
	return documentAdapter.DeleteDocument(db, tc.SchemaPrefix, documentID)
}
