package store

import (
	"fmt"
	"welltaxpro/src/internal/adapter"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

// GetDiscountCodes retrieves discount codes for a tenant, optionally filtered by affiliate
func (s *Store) GetDiscountCodes(tenantID string, affiliateID *string, activeOnly bool) ([]*types.DiscountCode, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	adpt, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch discount codes
	return adpt.GetDiscountCodes(db, tc.SchemaPrefix, affiliateID, activeOnly)
}

// GetDiscountCodeByID retrieves a specific discount code by ID
func (s *Store) GetDiscountCodeByID(tenantID string, codeID string) (*types.DiscountCode, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	adpt, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch discount code
	return adpt.GetDiscountCodeByID(db, tc.SchemaPrefix, codeID)
}

// GetDiscountCodeByCode retrieves a discount code by its code string
func (s *Store) GetDiscountCodeByCode(tenantID string, code string) (*types.DiscountCode, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	adpt, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch discount code
	return adpt.GetDiscountCodeByCode(db, tc.SchemaPrefix, code)
}

// CreateDiscountCode creates a new discount code for an affiliate
func (s *Store) CreateDiscountCode(tenantID string, discountCode *types.DiscountCode) (*types.DiscountCode, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	adpt, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to create discount code
	return adpt.CreateDiscountCode(db, tc.SchemaPrefix, discountCode)
}

// UpdateDiscountCode updates an existing discount code
func (s *Store) UpdateDiscountCode(tenantID string, codeID string, discountCode *types.DiscountCode) (*types.DiscountCode, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	adpt, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to update discount code
	return adpt.UpdateDiscountCode(db, tc.SchemaPrefix, codeID, discountCode)
}

// DeactivateDiscountCode deactivates a discount code
func (s *Store) DeactivateDiscountCode(tenantID string, codeID string) error {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	// Get the appropriate adapter for this tenant
	adpt, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to deactivate discount code
	return adpt.DeactivateDiscountCode(db, tc.SchemaPrefix, codeID)
}
