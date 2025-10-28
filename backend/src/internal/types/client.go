package types

import "github.com/google/uuid"

// Client represents a tax client from a tenant's database
// This is the universal type that ALL adapters must return
//
// REQUIRED FIELDS (all adapters must provide):
//   - ID: Unique identifier
//   - Email: Client's email address
//   - Role: Client role (user, admin, etc.)
//   - CreatedAt: Account creation timestamp
//
// OPTIONAL FIELDS (may be nil if not available):
//   - FirstName, LastName: Client name
//   - Phone: Contact phone number
//   - Address1, City, State, Zipcode: Physical address
//
// Field Mapping (MyWellTax adapter):
//   taxes.user.id → ID
//   taxes.user.email → Email
//   taxes.user.first_name → FirstName
//   taxes.user.middle_name → MiddleName
//   taxes.user.last_name → LastName
//   taxes.user.phone → Phone
//   taxes.user.dob → Dob
//   taxes.user.ssn → Ssn (masked)
//   taxes.user.address1 → Address1
//   taxes.user.address2 → Address2
//   taxes.user.city → City
//   taxes.user.state → State
//   taxes.user.zipcode → Zipcode
//   taxes.user.role → Role
//   taxes.user.created_at → CreatedAt
type Client struct {
	// REQUIRED FIELDS
	ID        uuid.UUID `json:"id"`        // Unique client identifier
	Email     string    `json:"email"`     // Client email (required)
	Role      string    `json:"role"`      // Client role: user, admin, guest
	CreatedAt string    `json:"createdAt"` // Account creation timestamp

	// OPTIONAL FIELDS
	FirstName  *string `json:"firstName,omitempty"`  // Client first name
	MiddleName *string `json:"middleName,omitempty"` // Client middle name
	LastName   *string `json:"lastName,omitempty"`   // Client last name
	Phone      *string `json:"phone,omitempty"`      // Contact phone
	Dob        *string `json:"dob,omitempty"`        // Date of birth
	Ssn        *string `json:"ssn,omitempty"`        // SSN (masked to show last 4 only)
	Address1   *string `json:"address1,omitempty"`   // Street address line 1
	Address2   *string `json:"address2,omitempty"`   // Street address line 2
	City       *string `json:"city,omitempty"`       // City
	State      *string `json:"state,omitempty"`      // State/province
	Zipcode    *int32  `json:"zipcode,omitempty"`    // Postal code
}
