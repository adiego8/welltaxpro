package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
)

// GenerateAffiliateToken creates a new access token for an affiliate
// Returns the plain token (to be shared with affiliate) and stores the hash
func GenerateAffiliateToken(db *sql.DB, schemaPrefix string, affiliateID uuid.UUID, expiresAt *time.Time, notes *string) (string, *types.AffiliateToken, error) {
	// Generate a secure random token (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	plainToken := hex.EncodeToString(tokenBytes)

	// Hash the token before storing (SHA256)
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	query := fmt.Sprintf(`
		INSERT INTO %s.affiliate_tokens (
			affiliate_id, token_hash, expires_at, notes, is_active
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, affiliate_id, token_hash, expires_at, last_used_at, is_active, notes, created_at, updated_at
	`, schemaPrefix)

	logger.Infof("Generating affiliate token for affiliate %s", affiliateID)

	token := &types.AffiliateToken{}
	err := db.QueryRow(
		query,
		affiliateID,
		tokenHash,
		expiresAt,
		notes,
		true,
	).Scan(
		&token.ID,
		&token.AffiliateID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.LastUsedAt,
		&token.IsActive,
		&token.Notes,
		&token.CreatedAt,
		&token.UpdatedAt,
	)

	if err != nil {
		logger.Errorf("Failed to generate affiliate token: %v", err)
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	logger.Infof("Successfully generated token %s for affiliate %s", token.ID, affiliateID)
	return plainToken, token, nil
}

// ValidateAffiliateToken validates a token and returns the affiliate ID
// Also updates the last_used_at timestamp
func ValidateAffiliateToken(db *sql.DB, schemaPrefix string, plainToken string) (uuid.UUID, error) {
	// Hash the provided token
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	query := fmt.Sprintf(`
		UPDATE %s.affiliate_tokens
		SET last_used_at = NOW()
		WHERE token_hash = $1
		  AND is_active = true
		  AND (expires_at IS NULL OR expires_at > NOW())
		RETURNING affiliate_id
	`, schemaPrefix)

	logger.Infof("Validating affiliate token")

	var affiliateID uuid.UUID
	err := db.QueryRow(query, tokenHash).Scan(&affiliateID)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warning("Invalid or expired affiliate token")
			return uuid.Nil, fmt.Errorf("invalid or expired token")
		}
		logger.Errorf("Failed to validate affiliate token: %v", err)
		return uuid.Nil, fmt.Errorf("failed to validate token: %w", err)
	}

	logger.Infof("Successfully validated token for affiliate %s", affiliateID)
	return affiliateID, nil
}

// GetAffiliateTokens retrieves all tokens for a specific affiliate
func GetAffiliateTokens(db *sql.DB, schemaPrefix string, affiliateID uuid.UUID, activeOnly bool) ([]*types.AffiliateToken, error) {
	whereClause := "WHERE affiliate_id = $1"
	if activeOnly {
		whereClause += " AND is_active = true"
	}

	query := fmt.Sprintf(`
		SELECT id, affiliate_id, token_hash, expires_at, last_used_at, is_active, notes, created_at, updated_at
		FROM %s.affiliate_tokens
		%s
		ORDER BY created_at DESC
	`, schemaPrefix, whereClause)

	logger.Infof("Fetching tokens for affiliate %s (activeOnly=%v)", affiliateID, activeOnly)

	rows, err := db.Query(query, affiliateID)
	if err != nil {
		logger.Errorf("Failed to query affiliate tokens: %v", err)
		return nil, fmt.Errorf("failed to query tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*types.AffiliateToken
	for rows.Next() {
		token := &types.AffiliateToken{}
		err := rows.Scan(
			&token.ID,
			&token.AffiliateID,
			&token.TokenHash,
			&token.ExpiresAt,
			&token.LastUsedAt,
			&token.IsActive,
			&token.Notes,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			logger.Errorf("Failed to scan token row: %v", err)
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("Error iterating token rows: %v", err)
		return nil, fmt.Errorf("error iterating tokens: %w", err)
	}

	logger.Infof("Successfully fetched %d tokens for affiliate %s", len(tokens), affiliateID)
	return tokens, nil
}

// RevokeAffiliateToken revokes (deactivates) a token
func RevokeAffiliateToken(db *sql.DB, schemaPrefix string, tokenID uuid.UUID) error {
	query := fmt.Sprintf(`
		UPDATE %s.affiliate_tokens
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`, schemaPrefix)

	logger.Infof("Revoking affiliate token %s", tokenID)

	result, err := db.Exec(query, tokenID)
	if err != nil {
		logger.Errorf("Failed to revoke token: %v", err)
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	logger.Infof("Successfully revoked token %s", tokenID)
	return nil
}

// DeleteExpiredTokens removes expired tokens from the database
// This is a maintenance function that should be run periodically
func DeleteExpiredTokens(db *sql.DB, schemaPrefix string) (int64, error) {
	query := fmt.Sprintf(`
		DELETE FROM %s.affiliate_tokens
		WHERE expires_at IS NOT NULL AND expires_at < NOW()
	`, schemaPrefix)

	logger.Info("Deleting expired affiliate tokens")

	result, err := db.Exec(query)
	if err != nil {
		logger.Errorf("Failed to delete expired tokens: %v", err)
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Infof("Successfully deleted %d expired tokens", rowsAffected)
	return rowsAffected, nil
}
