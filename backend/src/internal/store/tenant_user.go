package store

import (
	"database/sql"
	"fmt"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
)

// GetTenantUserByFirebaseUID retrieves a tenant user by their Firebase UID
func (s *Store) GetTenantUserByFirebaseUID(firebaseUID string) (*types.TenantUser, error) {
	query := `
		SELECT id, tenant_id, client_id, firebase_uid, email, is_active, created_at, updated_at
		FROM tenant_users
		WHERE firebase_uid = $1 AND is_active = true
	`

	var tu types.TenantUser
	err := s.DB.QueryRow(query, firebaseUID).Scan(
		&tu.ID,
		&tu.TenantID,
		&tu.ClientID,
		&tu.FirebaseUID,
		&tu.Email,
		&tu.IsActive,
		&tu.CreatedAt,
		&tu.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant user not found for firebase uid: %s", firebaseUID)
		}
		logger.Errorf("Failed to get tenant user by firebase uid %s: %v", firebaseUID, err)
		return nil, err
	}

	logger.Infof("Found tenant user %s for firebase uid %s (tenant: %s, client: %s)",
		tu.ID.String(), firebaseUID, tu.TenantID, tu.ClientID.String())
	return &tu, nil
}

// GetTenantUser retrieves a tenant user by ID
func (s *Store) GetTenantUser(id uuid.UUID) (*types.TenantUser, error) {
	query := `
		SELECT id, tenant_id, client_id, firebase_uid, email, is_active, created_at, updated_at
		FROM tenant_users
		WHERE id = $1
	`

	var tu types.TenantUser
	err := s.DB.QueryRow(query, id).Scan(
		&tu.ID,
		&tu.TenantID,
		&tu.ClientID,
		&tu.FirebaseUID,
		&tu.Email,
		&tu.IsActive,
		&tu.CreatedAt,
		&tu.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant user not found: %s", id.String())
		}
		logger.Errorf("Failed to get tenant user %s: %v", id.String(), err)
		return nil, err
	}

	return &tu, nil
}

// CreateTenantUser creates a new tenant user
func (s *Store) CreateTenantUser(tu *types.TenantUser) error {
	query := `
		INSERT INTO tenant_users (id, tenant_id, client_id, firebase_uid, email, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	// Generate UUID if not provided
	if tu.ID == uuid.Nil {
		tu.ID = uuid.New()
	}

	err := s.DB.QueryRow(
		query,
		tu.ID,
		tu.TenantID,
		tu.ClientID,
		tu.FirebaseUID,
		tu.Email,
		tu.IsActive,
	).Scan(&tu.CreatedAt, &tu.UpdatedAt)

	if err != nil {
		logger.Errorf("Failed to create tenant user: %v", err)
		return err
	}

	logger.Infof("Created tenant user %s (firebase_uid: %s, tenant: %s, client: %s)",
		tu.ID.String(), tu.FirebaseUID, tu.TenantID, tu.ClientID.String())
	return nil
}

// GetTenantUsersByTenant retrieves all tenant users for a specific tenant
func (s *Store) GetTenantUsersByTenant(tenantID string) ([]*types.TenantUser, error) {
	query := `
		SELECT id, tenant_id, client_id, firebase_uid, email, is_active, created_at, updated_at
		FROM tenant_users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.DB.Query(query, tenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant users for tenant %s: %v", tenantID, err)
		return nil, err
	}
	defer rows.Close()

	var users []*types.TenantUser
	for rows.Next() {
		var tu types.TenantUser
		err := rows.Scan(
			&tu.ID,
			&tu.TenantID,
			&tu.ClientID,
			&tu.FirebaseUID,
			&tu.Email,
			&tu.IsActive,
			&tu.CreatedAt,
			&tu.UpdatedAt,
		)
		if err != nil {
			logger.Errorf("Failed to scan tenant user: %v", err)
			continue
		}
		users = append(users, &tu)
	}

	return users, nil
}

// DeactivateTenantUser deactivates a tenant user
func (s *Store) DeactivateTenantUser(id uuid.UUID) error {
	query := `
		UPDATE tenant_users
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.DB.Exec(query, id)
	if err != nil {
		logger.Errorf("Failed to deactivate tenant user %s: %v", id.String(), err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tenant user not found: %s", id.String())
	}

	logger.Infof("Deactivated tenant user %s", id.String())
	return nil
}
