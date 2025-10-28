package adapter

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
)

// GetDiscountCodes retrieves discount codes from MyWellTax database
func (a *MyWellTaxAdapter) GetDiscountCodes(db *sql.DB, schemaPrefix string, affiliateID *string, activeOnly bool) ([]*types.DiscountCode, error) {
	var conditions []string
	var args []interface{}
	argCount := 0

	if affiliateID != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("affiliate_id = $%d", argCount))
		args = append(args, *affiliateID)
	}

	if activeOnly {
		conditions = append(conditions, "is_active = true")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, code, description, discount_type, discount_value,
		       max_uses, current_uses, valid_from, valid_until, is_active,
		       is_affiliate_code, affiliate_id, commission_rate, created_at, updated_at
		FROM %s.discount_codes
		%s
		ORDER BY created_at DESC
	`, schemaPrefix, whereClause)

	logger.Infof("MyWellTax adapter fetching discount codes (affiliateID=%v, activeOnly=%v)", affiliateID, activeOnly)

	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Errorf("MyWellTax adapter failed to query discount codes: %v", err)
		return nil, fmt.Errorf("failed to query discount codes: %w", err)
	}
	defer rows.Close()

	var codes []*types.DiscountCode
	for rows.Next() {
		code := &types.DiscountCode{}
		var description, validFrom, validUntil, updatedAt sql.NullString
		var maxUses sql.NullInt32
		var affiliateIDScan sql.NullString
		var commissionRate sql.NullFloat64

		err := rows.Scan(
			&code.ID,
			&code.Code,
			&description,
			&code.DiscountType,
			&code.DiscountValue,
			&maxUses,
			&code.CurrentUses,
			&validFrom,
			&validUntil,
			&code.IsActive,
			&code.IsAffiliateCode,
			&affiliateIDScan,
			&commissionRate,
			&code.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			logger.Errorf("MyWellTax adapter failed to scan discount code row: %v", err)
			return nil, fmt.Errorf("failed to scan discount code: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			code.Description = &description.String
		}
		if maxUses.Valid {
			maxUsesInt := int(maxUses.Int32)
			code.MaxUses = &maxUsesInt
		}
		if validFrom.Valid {
			code.ValidFrom = &validFrom.String
		}
		if validUntil.Valid {
			code.ValidUntil = &validUntil.String
		}
		if affiliateIDScan.Valid {
			aID, err := uuid.Parse(affiliateIDScan.String)
			if err == nil {
				code.AffiliateID = &aID
			}
		}
		if commissionRate.Valid {
			code.CommissionRate = &commissionRate.Float64
		}
		if updatedAt.Valid {
			code.UpdatedAt = &updatedAt.String
		}

		codes = append(codes, code)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("MyWellTax adapter error iterating discount code rows: %v", err)
		return nil, fmt.Errorf("error iterating discount codes: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully fetched %d discount codes", len(codes))
	return codes, nil
}

// GetDiscountCodeByID retrieves a specific discount code by ID
func (a *MyWellTaxAdapter) GetDiscountCodeByID(db *sql.DB, schemaPrefix string, codeID string) (*types.DiscountCode, error) {
	query := fmt.Sprintf(`
		SELECT id, code, description, discount_type, discount_value,
		       max_uses, current_uses, valid_from, valid_until, is_active,
		       is_affiliate_code, affiliate_id, commission_rate, created_at, updated_at
		FROM %s.discount_codes
		WHERE id = $1
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter fetching discount code %s", codeID)

	row := db.QueryRow(query, codeID)

	code := &types.DiscountCode{}
	var description, validFrom, validUntil, updatedAt sql.NullString
	var maxUses sql.NullInt32
	var affiliateID sql.NullString
	var commissionRate sql.NullFloat64

	err := row.Scan(
		&code.ID,
		&code.Code,
		&description,
		&code.DiscountType,
		&code.DiscountValue,
		&maxUses,
		&code.CurrentUses,
		&validFrom,
		&validUntil,
		&code.IsActive,
		&code.IsAffiliateCode,
		&affiliateID,
		&commissionRate,
		&code.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warningf("MyWellTax adapter discount code %s not found", codeID)
			return nil, fmt.Errorf("discount code not found")
		}
		logger.Errorf("MyWellTax adapter failed to scan discount code: %v", err)
		return nil, fmt.Errorf("failed to scan discount code: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		code.Description = &description.String
	}
	if maxUses.Valid {
		maxUsesInt := int(maxUses.Int32)
		code.MaxUses = &maxUsesInt
	}
	if validFrom.Valid {
		code.ValidFrom = &validFrom.String
	}
	if validUntil.Valid {
		code.ValidUntil = &validUntil.String
	}
	if affiliateID.Valid {
		aID, err := uuid.Parse(affiliateID.String)
		if err == nil {
			code.AffiliateID = &aID
		}
	}
	if commissionRate.Valid {
		code.CommissionRate = &commissionRate.Float64
	}
	if updatedAt.Valid {
		code.UpdatedAt = &updatedAt.String
	}

	logger.Infof("MyWellTax adapter successfully fetched discount code %s", code.Code)
	return code, nil
}

// GetDiscountCodeByCode retrieves a discount code by its code string
func (a *MyWellTaxAdapter) GetDiscountCodeByCode(db *sql.DB, schemaPrefix string, code string) (*types.DiscountCode, error) {
	query := fmt.Sprintf(`
		SELECT id, code, description, discount_type, discount_value,
		       max_uses, current_uses, valid_from, valid_until, is_active,
		       is_affiliate_code, affiliate_id, commission_rate, created_at, updated_at
		FROM %s.discount_codes
		WHERE UPPER(code) = UPPER($1)
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter fetching discount code by code: %s", code)

	row := db.QueryRow(query, code)

	discountCode := &types.DiscountCode{}
	var description, validFrom, validUntil, updatedAt sql.NullString
	var maxUses sql.NullInt32
	var affiliateID sql.NullString
	var commissionRate sql.NullFloat64

	err := row.Scan(
		&discountCode.ID,
		&discountCode.Code,
		&description,
		&discountCode.DiscountType,
		&discountCode.DiscountValue,
		&maxUses,
		&discountCode.CurrentUses,
		&validFrom,
		&validUntil,
		&discountCode.IsActive,
		&discountCode.IsAffiliateCode,
		&affiliateID,
		&commissionRate,
		&discountCode.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warningf("MyWellTax adapter discount code %s not found", code)
			return nil, fmt.Errorf("discount code not found")
		}
		logger.Errorf("MyWellTax adapter failed to scan discount code: %v", err)
		return nil, fmt.Errorf("failed to scan discount code: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		discountCode.Description = &description.String
	}
	if maxUses.Valid {
		maxUsesInt := int(maxUses.Int32)
		discountCode.MaxUses = &maxUsesInt
	}
	if validFrom.Valid {
		discountCode.ValidFrom = &validFrom.String
	}
	if validUntil.Valid {
		discountCode.ValidUntil = &validUntil.String
	}
	if affiliateID.Valid {
		aID, err := uuid.Parse(affiliateID.String)
		if err == nil {
			discountCode.AffiliateID = &aID
		}
	}
	if commissionRate.Valid {
		discountCode.CommissionRate = &commissionRate.Float64
	}
	if updatedAt.Valid {
		discountCode.UpdatedAt = &updatedAt.String
	}

	logger.Infof("MyWellTax adapter successfully fetched discount code %s", discountCode.Code)
	return discountCode, nil
}

// CreateDiscountCode creates a new discount code for an affiliate
func (a *MyWellTaxAdapter) CreateDiscountCode(db *sql.DB, schemaPrefix string, discountCode *types.DiscountCode) (*types.DiscountCode, error) {
	// Generate UUID if not provided
	if discountCode.ID == uuid.Nil {
		discountCode.ID = uuid.New()
	}

	// Set created timestamp
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	discountCode.CreatedAt = now

	// Uppercase the code
	discountCode.Code = strings.ToUpper(discountCode.Code)

	query := fmt.Sprintf(`
		INSERT INTO %s.discount_codes
		(id, code, description, discount_type, discount_value, max_uses, current_uses,
		 valid_from, valid_until, is_active, is_affiliate_code, affiliate_id, commission_rate, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, code, description, discount_type, discount_value, max_uses, current_uses,
		          valid_from, valid_until, is_active, is_affiliate_code, affiliate_id, commission_rate, created_at
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter creating discount code: %s", discountCode.Code)

	var description, validFrom, validUntil sql.NullString
	var maxUses sql.NullInt32
	var affiliateID sql.NullString
	var commissionRate sql.NullFloat64

	// Prepare nullable values for insert
	if discountCode.Description != nil {
		description.String = *discountCode.Description
		description.Valid = true
	}
	if discountCode.MaxUses != nil {
		maxUses.Int32 = int32(*discountCode.MaxUses)
		maxUses.Valid = true
	}
	if discountCode.ValidFrom != nil {
		validFrom.String = *discountCode.ValidFrom
		validFrom.Valid = true
	}
	if discountCode.ValidUntil != nil {
		validUntil.String = *discountCode.ValidUntil
		validUntil.Valid = true
	}
	if discountCode.AffiliateID != nil {
		affiliateID.String = discountCode.AffiliateID.String()
		affiliateID.Valid = true
	}
	if discountCode.CommissionRate != nil {
		commissionRate.Float64 = *discountCode.CommissionRate
		commissionRate.Valid = true
	}

	row := db.QueryRow(query,
		discountCode.ID,
		discountCode.Code,
		description,
		discountCode.DiscountType,
		discountCode.DiscountValue,
		maxUses,
		0, // current_uses starts at 0
		validFrom,
		validUntil,
		discountCode.IsActive,
		discountCode.IsAffiliateCode,
		affiliateID,
		commissionRate,
		now,
	)

	created := &types.DiscountCode{}
	err := row.Scan(
		&created.ID,
		&created.Code,
		&description,
		&created.DiscountType,
		&created.DiscountValue,
		&maxUses,
		&created.CurrentUses,
		&validFrom,
		&validUntil,
		&created.IsActive,
		&created.IsAffiliateCode,
		&affiliateID,
		&commissionRate,
		&created.CreatedAt,
	)
	if err != nil {
		logger.Errorf("MyWellTax adapter failed to create discount code: %v", err)
		return nil, fmt.Errorf("failed to create discount code: %w", err)
	}

	// Handle nullable fields in response
	if description.Valid {
		created.Description = &description.String
	}
	if maxUses.Valid {
		maxUsesInt := int(maxUses.Int32)
		created.MaxUses = &maxUsesInt
	}
	if validFrom.Valid {
		created.ValidFrom = &validFrom.String
	}
	if validUntil.Valid {
		created.ValidUntil = &validUntil.String
	}
	if affiliateID.Valid {
		aID, err := uuid.Parse(affiliateID.String)
		if err == nil {
			created.AffiliateID = &aID
		}
	}
	if commissionRate.Valid {
		created.CommissionRate = &commissionRate.Float64
	}

	logger.Infof("MyWellTax adapter successfully created discount code %s", created.Code)
	return created, nil
}

// UpdateDiscountCode updates an existing discount code
func (a *MyWellTaxAdapter) UpdateDiscountCode(db *sql.DB, schemaPrefix string, codeID string, discountCode *types.DiscountCode) (*types.DiscountCode, error) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	updatedAt := now

	// Uppercase the code
	if discountCode.Code != "" {
		discountCode.Code = strings.ToUpper(discountCode.Code)
	}

	query := fmt.Sprintf(`
		UPDATE %s.discount_codes
		SET code = $1, description = $2, discount_type = $3, discount_value = $4,
		    max_uses = $5, valid_from = $6, valid_until = $7, is_active = $8,
		    commission_rate = $9, updated_at = $10
		WHERE id = $11
		RETURNING id, code, description, discount_type, discount_value, max_uses, current_uses,
		          valid_from, valid_until, is_active, is_affiliate_code, affiliate_id, commission_rate, created_at, updated_at
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter updating discount code %s", codeID)

	var description, validFrom, validUntil sql.NullString
	var maxUses sql.NullInt32
	var commissionRate sql.NullFloat64

	// Prepare nullable values
	if discountCode.Description != nil {
		description.String = *discountCode.Description
		description.Valid = true
	}
	if discountCode.MaxUses != nil {
		maxUses.Int32 = int32(*discountCode.MaxUses)
		maxUses.Valid = true
	}
	if discountCode.ValidFrom != nil {
		validFrom.String = *discountCode.ValidFrom
		validFrom.Valid = true
	}
	if discountCode.ValidUntil != nil {
		validUntil.String = *discountCode.ValidUntil
		validUntil.Valid = true
	}
	if discountCode.CommissionRate != nil {
		commissionRate.Float64 = *discountCode.CommissionRate
		commissionRate.Valid = true
	}

	row := db.QueryRow(query,
		discountCode.Code,
		description,
		discountCode.DiscountType,
		discountCode.DiscountValue,
		maxUses,
		validFrom,
		validUntil,
		discountCode.IsActive,
		commissionRate,
		updatedAt,
		codeID,
	)

	updated := &types.DiscountCode{}
	var affiliateID sql.NullString
	var updatedAtScan sql.NullString

	err := row.Scan(
		&updated.ID,
		&updated.Code,
		&description,
		&updated.DiscountType,
		&updated.DiscountValue,
		&maxUses,
		&updated.CurrentUses,
		&validFrom,
		&validUntil,
		&updated.IsActive,
		&updated.IsAffiliateCode,
		&affiliateID,
		&commissionRate,
		&updated.CreatedAt,
		&updatedAtScan,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warningf("MyWellTax adapter discount code %s not found for update", codeID)
			return nil, fmt.Errorf("discount code not found")
		}
		logger.Errorf("MyWellTax adapter failed to update discount code: %v", err)
		return nil, fmt.Errorf("failed to update discount code: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		updated.Description = &description.String
	}
	if maxUses.Valid {
		maxUsesInt := int(maxUses.Int32)
		updated.MaxUses = &maxUsesInt
	}
	if validFrom.Valid {
		updated.ValidFrom = &validFrom.String
	}
	if validUntil.Valid {
		updated.ValidUntil = &validUntil.String
	}
	if affiliateID.Valid {
		aID, err := uuid.Parse(affiliateID.String)
		if err == nil {
			updated.AffiliateID = &aID
		}
	}
	if commissionRate.Valid {
		updated.CommissionRate = &commissionRate.Float64
	}
	if updatedAtScan.Valid {
		updated.UpdatedAt = &updatedAtScan.String
	}

	logger.Infof("MyWellTax adapter successfully updated discount code %s", updated.Code)
	return updated, nil
}

// DeactivateDiscountCode deactivates a discount code
func (a *MyWellTaxAdapter) DeactivateDiscountCode(db *sql.DB, schemaPrefix string, codeID string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	query := fmt.Sprintf(`
		UPDATE %s.discount_codes
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter deactivating discount code %s", codeID)

	result, err := db.Exec(query, now, codeID)
	if err != nil {
		logger.Errorf("MyWellTax adapter failed to deactivate discount code: %v", err)
		return fmt.Errorf("failed to deactivate discount code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warningf("MyWellTax adapter discount code %s not found for deactivation", codeID)
		return fmt.Errorf("discount code not found")
	}

	logger.Infof("MyWellTax adapter successfully deactivated discount code %s", codeID)
	return nil
}
