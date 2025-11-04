package types

import (
	"time"

	"github.com/google/uuid"
)

// TenantUser represents a user who can access their own data in the tenant portal
// These are clients who have registered to view their filings, documents, and profile (read-only)
type TenantUser struct {
	ID          uuid.UUID `json:"id"`
	TenantID    string    `json:"tenantId"`    // Reference to tenant_connections.tenant_id
	ClientID    uuid.UUID `json:"clientId"`    // Reference to the client record in tenant's database
	FirebaseUID string    `json:"firebaseUid"` // Firebase UID for authentication (Google/Phone)
	Email       string    `json:"email"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CanAccess checks if this tenant user can access specific data
func (tu *TenantUser) CanAccess(tenantID string, clientID string) bool {
	return tu.IsActive && tu.TenantID == tenantID && tu.ClientID.String() == clientID
}
