package adapter

import (
	"database/sql"
	"fmt"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

// GetClientsByFilings retrieves all clients with their filings (with pagination)
func (a *MyWellTaxAdapter) GetClientsByFilings(db *sql.DB, schemaPrefix string, limit int, offset int) ([]*types.ClientComprehensive, error) {
	// Build query to find distinct client IDs with filings (ordered by most recent filing)
	query := fmt.Sprintf(`
		SELECT DISTINCT ON (f.user_id) f.user_id
		FROM %s.filing f
		ORDER BY f.user_id, f.created_at DESC
		LIMIT $1 OFFSET $2
	`, schemaPrefix)

	logger.Infof("Querying filings with pagination - limit: %d, offset: %d", limit, offset)

	// Query for client IDs
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query client IDs: %w", err)
	}
	defer rows.Close()

	var clientIDs []string
	for rows.Next() {
		var clientID string
		if err := rows.Scan(&clientID); err != nil {
			return nil, fmt.Errorf("failed to scan client ID: %w", err)
		}
		clientIDs = append(clientIDs, clientID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client IDs: %w", err)
	}

	logger.Infof("Found %d clients with filings", len(clientIDs))

	// For each client, get comprehensive data (includes all their filings)
	result := make([]*types.ClientComprehensive, 0, len(clientIDs))
	for _, clientID := range clientIDs {
		comprehensive, err := a.GetClientComprehensive(db, schemaPrefix, clientID)
		if err != nil {
			logger.Warningf("Failed to get comprehensive data for client %s: %v", clientID, err)
			continue
		}
		result = append(result, comprehensive)
	}

	logger.Infof("Returning %d clients with all their filings", len(result))
	return result, nil
}
