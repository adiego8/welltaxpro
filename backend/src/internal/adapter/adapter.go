package adapter

import (
	"database/sql"
	"welltaxpro/src/internal/types"
)

// ClientAdapter defines the interface for tenant-specific client data access
// Each tax platform (MyWellTax, Drake, Lacerte, etc.) implements this interface
type ClientAdapter interface {
	// GetClients retrieves all clients from the tenant's database
	GetClients(db *sql.DB, schemaPrefix string) ([]*types.Client, error)

	// GetClientByID retrieves a specific client by ID from the tenant's database
	GetClientByID(db *sql.DB, schemaPrefix string, clientID string) (*types.Client, error)

	// GetClientComprehensive retrieves all data related to a client (filings, dependents, etc.)
	GetClientComprehensive(db *sql.DB, schemaPrefix string, clientID string) (*types.ClientComprehensive, error)

	// GetClientsByFilings retrieves clients with their filings (paginated)
	// Returns ClientComprehensive for each client with all their filings
	// Filtering should be done on the frontend
	GetClientsByFilings(db *sql.DB, schemaPrefix string, limit int, offset int) ([]*types.ClientComprehensive, error)

	// GetAffiliates retrieves all affiliates from the tenant's database
	GetAffiliates(db *sql.DB, schemaPrefix string, activeOnly bool) ([]*types.Affiliate, error)

	// GetAffiliateByID retrieves a specific affiliate by ID from the tenant's database
	GetAffiliateByID(db *sql.DB, schemaPrefix string, affiliateID string) (*types.Affiliate, error)

	// CreateAffiliate creates a new affiliate in the tenant's database
	CreateAffiliate(db *sql.DB, schemaPrefix string, affiliate *types.Affiliate) (*types.Affiliate, error)

	// UpdateAffiliate updates an existing affiliate in the tenant's database
	UpdateAffiliate(db *sql.DB, schemaPrefix string, affiliateID string, affiliate *types.Affiliate) (*types.Affiliate, error)

	// GetCommissionsByAffiliate retrieves commissions for a specific affiliate (or all if affiliateID is nil)
	GetCommissionsByAffiliate(db *sql.DB, schemaPrefix string, affiliateID *string, status *string, limit int) ([]*types.Commission, error)

	// GetAffiliateStats calculates aggregate statistics for an affiliate
	GetAffiliateStats(db *sql.DB, schemaPrefix string, affiliateID string) (*types.AffiliateStats, error)

	// ApproveCommission approves a pending commission
	ApproveCommission(db *sql.DB, schemaPrefix string, commissionID string) (*types.Commission, error)

	// MarkCommissionPaid marks an approved commission as paid
	MarkCommissionPaid(db *sql.DB, schemaPrefix string, commissionID string) (*types.Commission, error)

	// CancelCommission cancels a commission with a reason
	CancelCommission(db *sql.DB, schemaPrefix string, commissionID string, reason string) (*types.Commission, error)

	// GetDiscountCodes retrieves discount codes for a tenant, optionally filtered by affiliate
	GetDiscountCodes(db *sql.DB, schemaPrefix string, affiliateID *string, activeOnly bool) ([]*types.DiscountCode, error)

	// GetDiscountCodeByID retrieves a specific discount code by ID
	GetDiscountCodeByID(db *sql.DB, schemaPrefix string, codeID string) (*types.DiscountCode, error)

	// GetDiscountCodeByCode retrieves a discount code by its code string
	GetDiscountCodeByCode(db *sql.DB, schemaPrefix string, code string) (*types.DiscountCode, error)

	// CreateDiscountCode creates a new discount code for an affiliate
	CreateDiscountCode(db *sql.DB, schemaPrefix string, discountCode *types.DiscountCode) (*types.DiscountCode, error)

	// UpdateDiscountCode updates an existing discount code
	UpdateDiscountCode(db *sql.DB, schemaPrefix string, codeID string, discountCode *types.DiscountCode) (*types.DiscountCode, error)

	// DeactivateDiscountCode deactivates a discount code
	DeactivateDiscountCode(db *sql.DB, schemaPrefix string, codeID string) error

	// CreateDocument creates a new document record in the tenant's database
	CreateDocument(db *sql.DB, schemaPrefix string, document *types.Document) (*types.Document, error)

	// GetDocumentByID retrieves a specific document by ID
	GetDocumentByID(db *sql.DB, schemaPrefix string, documentID string) (*types.Document, error)

	// GetDocumentsByFilingID retrieves all documents associated with a filing
	GetDocumentsByFilingID(db *sql.DB, schemaPrefix string, filingID string) ([]*types.Document, error)

	// DeleteDocument removes a document record from the tenant's database
	DeleteDocument(db *sql.DB, schemaPrefix string, documentID string) error

	// GetAdapterType returns the unique identifier for this adapter
	GetAdapterType() string
}

// AdapterFactory creates the appropriate adapter based on adapter type
func NewAdapter(adapterType string) (ClientAdapter, error) {
	switch adapterType {
	case "mywelltax":
		return &MyWellTaxAdapter{}, nil
	default:
		// Default to MyWellTax for now
		return &MyWellTaxAdapter{}, nil
	}
}
