package types

import (
	"time"

	"github.com/google/uuid"
)

// EmployeeTenantAssociation represents the relationship between an employee and a tenant
type EmployeeTenantAssociation struct {
	ID         uuid.UUID `json:"id"`
	EmployeeID uuid.UUID `json:"employeeId"`
	TenantID   string    `json:"tenantId"`
	Role       string    `json:"role"`     // Role within this tenant: 'admin', 'accountant', 'viewer'
	IsActive   bool      `json:"isActive"` // Can this employee access this tenant?
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	CreatedBy  uuid.UUID `json:"createdBy"` // Who granted this access
}

// EmployeeWithTenants represents an employee with their tenant associations
type EmployeeWithTenants struct {
	Employee
	Tenants []EmployeeTenantAssociation `json:"tenants"`
}

// TenantAccess represents simplified tenant access info for an employee
type TenantAccess struct {
	TenantID   string `json:"tenantId"`
	TenantName string `json:"tenantName"`
	Role       string `json:"role"`
	IsActive   bool   `json:"isActive"`
}
