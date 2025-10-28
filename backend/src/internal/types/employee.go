package types

import (
	"time"

	"github.com/google/uuid"
)

// Employee represents a WellTaxPro staff member
type Employee struct {
	ID          uuid.UUID `json:"id"`
	FirebaseUID string    `json:"firebaseUid"`
	Email       string    `json:"email"`
	FirstName   *string   `json:"firstName,omitempty"`
	LastName    *string   `json:"lastName,omitempty"`
	Role        string    `json:"role"` // 'admin', 'accountant', 'support'
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// FullName returns the employee's full name
func (e *Employee) FullName() string {
	if e.FirstName != nil && e.LastName != nil {
		return *e.FirstName + " " + *e.LastName
	}
	if e.FirstName != nil {
		return *e.FirstName
	}
	if e.LastName != nil {
		return *e.LastName
	}
	return e.Email
}

// IsAdmin checks if employee has admin role
func (e *Employee) IsAdmin() bool {
	return e.Role == "admin"
}

// CanAccessTenant checks if employee can access a specific tenant
// In the future, this could check against employee-tenant associations
func (e *Employee) CanAccessTenant(tenantID string) bool {
	// For now, all active employees can access all tenants
	// TODO: Implement tenant-specific permissions
	return e.IsActive
}
