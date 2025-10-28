package types

import "github.com/google/uuid"

// ClientComprehensive contains all data related to a client
// This is the complete view of a client including all relationships
//
// ALL adapters returning comprehensive data must populate:
//   - Client: Basic client information (required)
//
// OPTIONAL (populate if available in tenant's database):
//   - Spouse: Spouse information
//   - Dependents: List of dependents
//   - Filings: Tax filings with all related data
type ClientComprehensive struct {
	Client     *Client      `json:"client"`               // Basic client info (REQUIRED)
	Spouse     *Spouse      `json:"spouse,omitempty"`     // Spouse info (optional)
	Dependents []*Dependent `json:"dependents,omitempty"` // Dependents (optional)
	Filings    []*Filing    `json:"filings,omitempty"`    // Tax filings (optional)
}

// Spouse information
type Spouse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"userId"`
	FirstName  string    `json:"firstName"`
	MiddleName *string   `json:"middleName"`
	LastName   string    `json:"lastName"`
	Email      *string   `json:"email"`
	Phone      *string   `json:"phone"`
	Dob        string    `json:"dob"`
	Ssn        string    `json:"ssn"`
	IsDeath    bool      `json:"isDeath"`
	DeathDate  *string   `json:"deathDate"`
	CreatedAt  string    `json:"createdAt"`
}

// Dependent information
type Dependent struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"userId"`
	FirstName          string    `json:"firstName"`
	MiddleName         *string   `json:"middleName"`
	LastName           string    `json:"lastName"`
	Dob                string    `json:"dob"`
	Ssn                string    `json:"ssn"`
	Relationship       string    `json:"relationship"`
	TimeWithApplicant  string    `json:"timeWithApplicant"`
	ExclusiveClaim     bool      `json:"exclusiveClaim"`
	Documents          []string  `json:"documents,omitempty"` // Required documentation types
	CreatedAt          string    `json:"createdAt"`
	UpdatedAt          *string   `json:"updatedAt"`
}

// Filing represents a tax filing for a specific year
type Filing struct {
	ID                    uuid.UUID  `json:"id"`
	Year                  int        `json:"year"`
	UserID                uuid.UUID  `json:"userId"`
	MaritalStatus         *string    `json:"maritalStatus"`
	SpouseID              *uuid.UUID `json:"spouseId"`
	SourceOfIncome        []string   `json:"sourceOfIncome"`
	Deductions            []string   `json:"deductions"`
	Income                *int64     `json:"income"`
	MarketplaceInsurance  *bool      `json:"marketplaceInsurance"`
	CreatedAt             string     `json:"createdAt"`
	UpdatedAt             *string    `json:"updatedAt"`

	// Related data
	Status            *FilingStatus       `json:"status,omitempty"`
	Documents         []*Document         `json:"documents,omitempty"`
	Properties        []*Property         `json:"properties,omitempty"`
	IRAContributions  []*IRAContribution  `json:"iraContributions,omitempty"`
	Charities         []*Charity          `json:"charities,omitempty"`
	Childcares        []*Childcare        `json:"childcares,omitempty"`
	Payments          []*Payment          `json:"payments,omitempty"`
	Discounts         []*FilingDiscount   `json:"discounts,omitempty"`
}

// FilingStatus tracks the progress of a filing
type FilingStatus struct {
	ID          uuid.UUID `json:"id"`
	FilingID    uuid.UUID `json:"filingId"`
	LatestStep  int       `json:"latestStep"`
	IsCompleted bool      `json:"isCompleted"`
	Status      string    `json:"status"`
}

// Document represents an uploaded document
type Document struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"userId"`
	FilingID  *uuid.UUID `json:"filingId"`
	Name      string     `json:"name"`
	FilePath  string     `json:"filePath"`
	Type      string     `json:"type"`
	CreatedAt string     `json:"createdAt"`
	UpdatedAt *string    `json:"updatedAt"`
}

// Property represents rental property
type Property struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"userId"`
	Address1      string     `json:"address1"`
	Address2      *string    `json:"address2"`
	State         string     `json:"state"`
	City          string     `json:"city"`
	Zipcode       string     `json:"zipcode"`
	PurchasePrice float64    `json:"purchasePrice"`
	ClosingCost   float64    `json:"closingCost"`
	PurchaseDate  string     `json:"purchaseDate"`
	Rents         *float64   `json:"rents"`
	Royalties     *float64   `json:"royalties"`
	UpdatedAt     *string    `json:"updatedAt"`
	CreatedAt     string     `json:"createdAt"`
	Expenses      []*Expense `json:"expenses,omitempty"`
}

// Expense represents property expense
type Expense struct {
	ID         uuid.UUID `json:"id"`
	PropertyID uuid.UUID `json:"propertyId"`
	Name       string    `json:"name"`
	Amount     float64   `json:"amount"`
	CreatedAt  string    `json:"createdAt"`
}

// IRAContribution represents IRA contribution
type IRAContribution struct {
	ID          uuid.UUID `json:"id"`
	FilingID    uuid.UUID `json:"filingId"`
	AccountType string    `json:"accountType"`
	Amount      float64   `json:"amount"`
}

// Charity represents charitable contribution
type Charity struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"userId"`
	FilingID     *uuid.UUID `json:"filingId"`
	Name         string     `json:"name"`
	Contribution float64    `json:"contribution"`
}

// Childcare represents childcare expense
type Childcare struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"userId"`
	Name     string    `json:"name"`
	Amount   float64   `json:"amount"`
	TaxID    string    `json:"taxId"`
	Address1 string    `json:"address1"`
	Address2 *string   `json:"address2"`
	City     string    `json:"city"`
	State    string    `json:"state"`
	Zipcode  string    `json:"zipcode"`
}

// Payment represents a payment transaction
type Payment struct {
	ID               uuid.UUID      `json:"id"`
	FilingID         uuid.UUID      `json:"filingId"`
	StripeSessionID  string         `json:"stripeSessionId"`
	Amount           float64        `json:"amount"`
	OriginalAmount   *float64       `json:"originalAmount"`
	DiscountAmount   *float64       `json:"discountAmount"`
	DiscountCode     *string        `json:"discountCode"`
	Status           string         `json:"status"`
	CreatedAt        string         `json:"createdAt"`
	UpdatedAt        *string        `json:"updatedAt"`
	Items            []*PaymentItem `json:"items,omitempty"`
}

// PaymentItem represents line item in a payment
type PaymentItem struct {
	ID         uuid.UUID `json:"id"`
	PaymentID  uuid.UUID `json:"paymentId"`
	PriceID    string    `json:"priceId"`
	Name       string    `json:"name"`
	Quantity   int       `json:"quantity"`
	UnitAmount float64   `json:"unitAmount"`
}

// FilingDiscount represents discount applied to a filing
type FilingDiscount struct {
	ID             uuid.UUID `json:"id"`
	FilingID       uuid.UUID `json:"filingId"`
	DiscountCodeID uuid.UUID `json:"discountCodeId"`
	OriginalAmount float64   `json:"originalAmount"`
	DiscountAmount float64   `json:"discountAmount"`
	FinalAmount    float64   `json:"finalAmount"`
	AppliedAt      string    `json:"appliedAt"`
	Code           *string   `json:"code,omitempty"` // Joined from discount_codes
}
