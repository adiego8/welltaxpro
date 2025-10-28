package store

import (
	"fmt"
	"time"
	"welltaxpro/src/internal/adapter"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
)

// GetAffiliates retrieves all affiliates for a specific tenant using the appropriate adapter
func (s *Store) GetAffiliates(tenantID string, activeOnly bool) ([]*types.Affiliate, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch affiliates
	return affiliateAdapter.GetAffiliates(db, tc.SchemaPrefix, activeOnly)
}

// GetAffiliateByID retrieves a specific affiliate by ID for a tenant using the appropriate adapter
func (s *Store) GetAffiliateByID(tenantID string, affiliateID string) (*types.Affiliate, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch affiliate
	return affiliateAdapter.GetAffiliateByID(db, tc.SchemaPrefix, affiliateID)
}

// CreateAffiliate creates a new affiliate for a tenant using the appropriate adapter
func (s *Store) CreateAffiliate(tenantID string, affiliate *types.Affiliate) (*types.Affiliate, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to create affiliate
	return affiliateAdapter.CreateAffiliate(db, tc.SchemaPrefix, affiliate)
}

// UpdateAffiliate updates an existing affiliate for a tenant using the appropriate adapter
func (s *Store) UpdateAffiliate(tenantID string, affiliateID string, affiliate *types.Affiliate) (*types.Affiliate, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to update affiliate
	return affiliateAdapter.UpdateAffiliate(db, tc.SchemaPrefix, affiliateID, affiliate)
}

// GetCommissionsByAffiliate retrieves commissions for a specific affiliate (or all if affiliateID is nil)
func (s *Store) GetCommissionsByAffiliate(tenantID string, affiliateID *string, status *string, limit int) ([]*types.Commission, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch commissions
	return affiliateAdapter.GetCommissionsByAffiliate(db, tc.SchemaPrefix, affiliateID, status, limit)
}

// GetAffiliateStats retrieves aggregate statistics for an affiliate
func (s *Store) GetAffiliateStats(tenantID string, affiliateID string) (*types.AffiliateStats, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to fetch stats
	return affiliateAdapter.GetAffiliateStats(db, tc.SchemaPrefix, affiliateID)
}

// ApproveCommission approves a pending commission
func (s *Store) ApproveCommission(tenantID string, commissionID string) (*types.Commission, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to approve commission
	return affiliateAdapter.ApproveCommission(db, tc.SchemaPrefix, commissionID)
}

// MarkCommissionPaid marks an approved commission as paid
func (s *Store) MarkCommissionPaid(tenantID string, commissionID string) (*types.Commission, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to mark commission as paid
	return affiliateAdapter.MarkCommissionPaid(db, tc.SchemaPrefix, commissionID)
}

// CancelCommission cancels a commission with a reason
func (s *Store) CancelCommission(tenantID string, commissionID string, reason string) (*types.Commission, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate adapter for this tenant
	affiliateAdapter, err := adapter.NewAdapter(tc.AdapterType)
	if err != nil {
		logger.Errorf("Failed to create adapter for tenant %s: %v", tenantID, err)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	logger.Infof("Using %s adapter for tenant %s", tc.AdapterType, tenantID)

	// Use adapter to cancel commission
	return affiliateAdapter.CancelCommission(db, tc.SchemaPrefix, commissionID, reason)
}

// GenerateAffiliateToken generates a new access token for an affiliate
func (s *Store) GenerateAffiliateToken(tenantID string, affiliateID uuid.UUID, expiresAt *time.Time, notes *string) (string, *types.AffiliateToken, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return "", nil, err
	}

	logger.Infof("Generating token for affiliate %s in tenant %s", affiliateID, tenantID)

	// Call the store function directly (not adapter-specific)
	return GenerateAffiliateToken(db, tc.SchemaPrefix, affiliateID, expiresAt, notes)
}

// GetAffiliateTokens retrieves all tokens for a specific affiliate
func (s *Store) GetAffiliateTokens(tenantID string, affiliateID uuid.UUID, activeOnly bool) ([]*types.AffiliateToken, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	logger.Infof("Fetching tokens for affiliate %s in tenant %s (activeOnly=%v)", affiliateID, tenantID, activeOnly)

	// Call the store function directly (not adapter-specific)
	return GetAffiliateTokens(db, tc.SchemaPrefix, affiliateID, activeOnly)
}

// RevokeAffiliateToken revokes (deactivates) a token
func (s *Store) RevokeAffiliateToken(tenantID string, tokenID uuid.UUID) error {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	logger.Infof("Revoking token %s in tenant %s", tokenID, tenantID)

	// Call the store function directly (not adapter-specific)
	return RevokeAffiliateToken(db, tc.SchemaPrefix, tokenID)
}

// ValidateAffiliateToken validates a token and returns the affiliate ID
func (s *Store) ValidateAffiliateToken(tenantID string, plainToken string) (uuid.UUID, error) {
	// Get tenant database connection and config
	db, tc, err := s.GetTenantDB(tenantID)
	if err != nil {
		return uuid.Nil, err
	}

	logger.Infof("Validating affiliate token for tenant %s", tenantID)

	// Call the store function directly (not adapter-specific)
	return ValidateAffiliateToken(db, tc.SchemaPrefix, plainToken)
}
