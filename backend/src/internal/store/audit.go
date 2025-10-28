package store

import (
	"encoding/json"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
)

// LogAudit creates an audit log entry
func (s *Store) LogAudit(log *types.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			employee_id, tenant_id, client_id, action, resource_type,
			resource_id, details, ip_address, user_agent
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	// Convert json.RawMessage to string for lib/pq JSONB compatibility
	// lib/pq expects JSONB to be passed as string, not []byte
	var detailsValue interface{}
	if len(log.Details) > 0 {
		detailsValue = string(log.Details)
	}

	err := s.DB.QueryRow(
		query,
		log.EmployeeID,
		log.TenantID,
		log.ClientID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		detailsValue,
		log.IPAddress,
		log.UserAgent,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		logger.Errorf("Failed to create audit log: %v", err)
		return err
	}

	return nil
}

// CreateAuditLog is a helper to create an audit log with common parameters
func (s *Store) CreateAuditLog(
	employeeID uuid.UUID,
	tenantID string,
	clientID *uuid.UUID,
	action string,
	resourceType string,
	resourceID *uuid.UUID,
	details interface{},
	ipAddress *string,
	userAgent *string,
) error {
	var detailsJSON json.RawMessage
	if details != nil {
		jsonData, err := json.Marshal(details)
		if err != nil {
			logger.Errorf("Failed to marshal audit details: %v", err)
			return err
		}
		detailsJSON = jsonData
	}

	auditLog := &types.AuditLog{
		EmployeeID:   employeeID,
		TenantID:     tenantID,
		ClientID:     clientID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      detailsJSON,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return s.LogAudit(auditLog)
}

// GetAuditLogsByEmployee retrieves audit logs for a specific employee
func (s *Store) GetAuditLogsByEmployee(employeeID uuid.UUID, limit int) ([]*types.AuditLog, error) {
	query := `
		SELECT id, employee_id, tenant_id, client_id, action, resource_type,
		       resource_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE employee_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	return s.queryAuditLogs(query, employeeID, limit)
}

// GetAuditLogsByClient retrieves audit logs for a specific client
func (s *Store) GetAuditLogsByClient(tenantID string, clientID uuid.UUID, limit int) ([]*types.AuditLog, error) {
	query := `
		SELECT id, employee_id, tenant_id, client_id, action, resource_type,
		       resource_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE tenant_id = $1 AND client_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	return s.queryAuditLogs(query, tenantID, clientID, limit)
}

// GetAuditLogsByTenant retrieves audit logs for a specific tenant
func (s *Store) GetAuditLogsByTenant(tenantID string, limit int) ([]*types.AuditLog, error) {
	query := `
		SELECT id, employee_id, tenant_id, client_id, action, resource_type,
		       resource_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	return s.queryAuditLogs(query, tenantID, limit)
}

// queryAuditLogs is a helper function to query audit logs
func (s *Store) queryAuditLogs(query string, args ...interface{}) ([]*types.AuditLog, error) {
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		logger.Errorf("Failed to query audit logs: %v", err)
		return nil, err
	}
	defer rows.Close()

	var logs []*types.AuditLog
	for rows.Next() {
		log := &types.AuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.EmployeeID,
			&log.TenantID,
			&log.ClientID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.Details,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			logger.Errorf("Failed to scan audit log: %v", err)
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
