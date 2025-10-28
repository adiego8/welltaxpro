package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an access record for compliance
type AuditLog struct {
	ID           uuid.UUID       `json:"id"`
	EmployeeID   uuid.UUID       `json:"employeeId"`
	TenantID     string          `json:"tenantId"`
	ClientID     *uuid.UUID      `json:"clientId,omitempty"`
	Action       string          `json:"action"` // VIEW, EDIT, DELETE, DOWNLOAD, CREATE, EXPORT
	ResourceType string          `json:"resourceType"` // CLIENT, FILING, DOCUMENT, SSN, SPOUSE, DEPENDENT
	ResourceID   *uuid.UUID      `json:"resourceId,omitempty"`
	Details      json.RawMessage `json:"details,omitempty"`
	IPAddress    *string         `json:"ipAddress,omitempty"`
	UserAgent    *string         `json:"userAgent,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
}

// Audit action constants
const (
	AuditActionView     = "VIEW"
	AuditActionEdit     = "EDIT"
	AuditActionDelete   = "DELETE"
	AuditActionDownload = "DOWNLOAD"
	AuditActionUpload   = "UPLOAD"
	AuditActionCreate   = "CREATE"
	AuditActionExport   = "EXPORT"
)

// Audit resource type constants
const (
	AuditResourceClient    = "CLIENT"
	AuditResourceFiling    = "FILING"
	AuditResourceDocument  = "DOCUMENT"
	AuditResourceSSN       = "SSN"
	AuditResourceSpouse    = "SPOUSE"
	AuditResourceDependent = "DEPENDENT"
)
