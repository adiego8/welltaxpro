package types

import (
	"time"

	"github.com/google/uuid"
)

// Affiliate represents a sales partner/affiliate from a tenant's database
// This is the universal type that ALL adapters must return
//
// Field Mapping (MyWellTax adapter):
//   taxes.affiliates.id → ID
//   taxes.affiliates.first_name → FirstName
//   taxes.affiliates.last_name → LastName
//   taxes.affiliates.email → Email
//   taxes.affiliates.phone → Phone
//   taxes.affiliates.default_commission_rate → DefaultCommissionRate
//   taxes.affiliates.stripe_connect_account_id → StripeConnectAccountID
//   taxes.affiliates.payout_method → PayoutMethod
//   taxes.affiliates.payout_threshold → PayoutThreshold
//   taxes.affiliates.is_active → IsActive
//   taxes.affiliates.created_at → CreatedAt
//   taxes.affiliates.updated_at → UpdatedAt
type Affiliate struct {
	ID                     uuid.UUID  `json:"id"`
	FirstName              string     `json:"firstName"`
	LastName               string     `json:"lastName"`
	Email                  string     `json:"email"`
	Phone                  *string    `json:"phone,omitempty"`
	DefaultCommissionRate  float64    `json:"defaultCommissionRate"` // Percentage (0-100)
	StripeConnectAccountID *string    `json:"stripeConnectAccountId,omitempty"`
	PayoutMethod           string     `json:"payoutMethod"` // MANUAL, STRIPE, PAYPAL
	PayoutThreshold        float64    `json:"payoutThreshold"`
	IsActive               bool       `json:"isActive"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              *time.Time `json:"updatedAt,omitempty"`
}

// Commission represents a commission earned by an affiliate
// Field Mapping (MyWellTax adapter):
//   taxes.commissions.* → Commission fields
type Commission struct {
	ID               uuid.UUID  `json:"id"`
	AffiliateID      uuid.UUID  `json:"affiliateId"`
	FilingID         uuid.UUID  `json:"filingId"`
	UserID           uuid.UUID  `json:"userId"`
	DiscountCodeID   uuid.UUID  `json:"discountCodeId"`
	PaymentID        *uuid.UUID `json:"paymentId,omitempty"`
	OrderAmount      float64    `json:"orderAmount"`      // Original order amount
	DiscountAmount   float64    `json:"discountAmount"`   // Discount applied
	NetAmount        float64    `json:"netAmount"`        // Amount after discount
	CommissionRate   float64    `json:"commissionRate"`   // Rate applied (0-100)
	CommissionAmount float64    `json:"commissionAmount"` // Affiliate's earning
	Status           string     `json:"status"`           // PENDING, APPROVED, PAID, CANCELLED
	ApprovedAt       *time.Time `json:"approvedAt,omitempty"`
	PaidAt           *time.Time `json:"paidAt,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt,omitempty"`

	// Related entities (optional, populated based on query)
	Affiliate *Affiliate     `json:"affiliate,omitempty"`
	Customer  *CustomerInfo  `json:"customer,omitempty"`
	Filing    *FilingSummary `json:"filing,omitempty"`
}

// CustomerInfo holds basic customer information for commission display
type CustomerInfo struct {
	ID        uuid.UUID `json:"id"`
	FirstName *string   `json:"firstName,omitempty"`
	LastName  *string   `json:"lastName,omitempty"`
	Email     string    `json:"email"`
}

// FilingSummary holds basic filing information for commission display
type FilingSummary struct {
	ID     uuid.UUID `json:"id"`
	Year   int       `json:"year"`
	Status string    `json:"status"`
}

// AffiliateToken represents a secure access token for an affiliate
type AffiliateToken struct {
	ID         uuid.UUID  `json:"id"`
	AffiliateID uuid.UUID `json:"affiliateId"`
	TokenHash  string     `json:"-"` // Never send to client
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	IsActive   bool       `json:"isActive"`
	Notes      *string    `json:"notes,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
}

// AffiliateStats represents aggregate statistics for an affiliate
type AffiliateStats struct {
	AffiliateID             uuid.UUID `json:"affiliateId"`
	TotalClicks             int       `json:"totalClicks"`
	TotalConversions        int       `json:"totalConversions"`
	ConversionRate          float64   `json:"conversionRate"` // Percentage
	TotalCommissionsEarned  float64   `json:"totalCommissionsEarned"`
	PendingCommissions      float64   `json:"pendingCommissions"`
	ApprovedCommissions     float64   `json:"approvedCommissions"`
	PaidCommissions         float64   `json:"paidCommissions"`
	CancelledCommissions    float64   `json:"cancelledCommissions"`
	TotalOrders             int       `json:"totalOrders"`
	TotalRevenue            float64   `json:"totalRevenue"` // Total order amounts
}

// DiscountCode represents a discount code in the system
// Field Mapping (MyWellTax adapter):
//   taxes.discount_codes.* → DiscountCode fields
type DiscountCode struct {
	ID              uuid.UUID  `json:"id"`
	Code            string     `json:"code"`
	Description     *string    `json:"description,omitempty"`
	DiscountType    string     `json:"discountType"`    // PERCENTAGE or FIXED_AMOUNT
	DiscountValue   float64    `json:"discountValue"`
	MaxUses         *int       `json:"maxUses,omitempty"`        // NULL means unlimited
	CurrentUses     int        `json:"currentUses"`
	ValidFrom       *string    `json:"validFrom,omitempty"`
	ValidUntil      *string    `json:"validUntil,omitempty"`
	IsActive        bool       `json:"isActive"`
	IsAffiliateCode bool       `json:"isAffiliateCode"`         // True if affiliate code
	AffiliateID     *uuid.UUID `json:"affiliateId,omitempty"`   // References affiliate
	CommissionRate  *float64   `json:"commissionRate,omitempty"` // Commission rate for this code
	CreatedAt       string     `json:"createdAt"`
	UpdatedAt       *string    `json:"updatedAt,omitempty"`
}

// IsValid checks if the discount code is valid for use
func (dc *DiscountCode) IsValid() bool {
	if !dc.IsActive {
		return false
	}

	now := time.Now()

	// Check validity dates
	if dc.ValidFrom != nil {
		validFrom, err := time.Parse("2006-01-02 15:04:05", *dc.ValidFrom)
		if err == nil && now.Before(validFrom) {
			return false
		}
	}

	if dc.ValidUntil != nil {
		validUntil, err := time.Parse("2006-01-02 15:04:05", *dc.ValidUntil)
		if err == nil && now.After(validUntil) {
			return false
		}
	}

	// Check usage limits
	if dc.MaxUses != nil && dc.CurrentUses >= *dc.MaxUses {
		return false
	}

	return true
}

// Commission status constants
const (
	CommissionStatusPending   = "PENDING"
	CommissionStatusApproved  = "APPROVED"
	CommissionStatusPaid      = "PAID"
	CommissionStatusCancelled = "CANCELLED"
)

// Payout method constants
const (
	PayoutMethodManual = "MANUAL"
	PayoutMethodStripe = "STRIPE"
	PayoutMethodPayPal = "PAYPAL"
)

// Discount type constants
const (
	DiscountTypePercentage  = "PERCENTAGE"
	DiscountTypeFixedAmount = "FIXED_AMOUNT"
)
