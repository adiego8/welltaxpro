package adapter

import (
	"database/sql"
	"fmt"
	"welltaxpro/src/internal/crypto"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

// MyWellTaxAdapter implements the ClientAdapter interface for MyWellTax schema
type MyWellTaxAdapter struct{}

// GetAdapterType returns the unique identifier for this adapter
func (a *MyWellTaxAdapter) GetAdapterType() string {
	return "mywelltax"
}

// GetClients retrieves all clients from MyWellTax database
// MyWellTax schema: taxes.user table with role='user' for clients
func (a *MyWellTaxAdapter) GetClients(db *sql.DB, schemaPrefix string) ([]*types.Client, error) {
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, email, phone, address1, city, state, zipcode, role, created_at
		FROM %s.user
		WHERE role = 'user'
		ORDER BY created_at DESC
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter executing query: %s", query)

	rows, err := db.Query(query)
	if err != nil {
		logger.Errorf("MyWellTax adapter failed to query clients: %v", err)
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []*types.Client
	for rows.Next() {
		client := &types.Client{}
		err := rows.Scan(
			&client.ID,
			&client.FirstName,
			&client.LastName,
			&client.Email,
			&client.Phone,
			&client.Address1,
			&client.City,
			&client.State,
			&client.Zipcode,
			&client.Role,
			&client.CreatedAt,
		)
		if err != nil {
			logger.Errorf("MyWellTax adapter failed to scan client row: %v", err)
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("MyWellTax adapter error iterating client rows: %v", err)
		return nil, fmt.Errorf("error iterating clients: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully fetched %d clients", len(clients))
	return clients, nil
}

// GetClientByID retrieves a specific client by ID from MyWellTax database
func (a *MyWellTaxAdapter) GetClientByID(db *sql.DB, schemaPrefix string, clientID string) (*types.Client, error) {
	query := fmt.Sprintf(`
		SELECT id, first_name, middle_name, last_name, email, phone, dob, ssn, address1, address2, city, state, zipcode, role, created_at
		FROM %s.user
		WHERE id = $1
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter fetching client %s", clientID)

	row := db.QueryRow(query, clientID)

	client := &types.Client{}
	var ssnEncrypted sql.NullString
	err := row.Scan(
		&client.ID,
		&client.FirstName,
		&client.MiddleName,
		&client.LastName,
		&client.Email,
		&client.Phone,
		&client.Dob,
		&ssnEncrypted,
		&client.Address1,
		&client.Address2,
		&client.City,
		&client.State,
		&client.Zipcode,
		&client.Role,
		&client.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("client not found")
		}
		logger.Errorf("MyWellTax adapter failed to get client %s: %v", clientID, err)
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Mask SSN for API response
	if ssnEncrypted.Valid && ssnEncrypted.String != "" {
		maskedSSN := crypto.MaskSSN(ssnEncrypted.String)
		client.Ssn = &maskedSSN
	}

	return client, nil
}
