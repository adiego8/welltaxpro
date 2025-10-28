package adapter

import (
	"database/sql"
	"fmt"
	"strings"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
)

// GetAffiliates retrieves all affiliates from MyWellTax database
func (a *MyWellTaxAdapter) GetAffiliates(db *sql.DB, schemaPrefix string, activeOnly bool) ([]*types.Affiliate, error) {
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, email, phone, default_commission_rate,
		       stripe_connect_account_id, payout_method, payout_threshold,
		       is_active, created_at, updated_at
		FROM %s.affiliates
		%s
		ORDER BY created_at DESC
	`, schemaPrefix, func() string {
		if activeOnly {
			return "WHERE is_active = true"
		}
		return ""
	}())

	logger.Infof("MyWellTax adapter fetching affiliates (activeOnly=%v)", activeOnly)

	rows, err := db.Query(query)
	if err != nil {
		logger.Errorf("MyWellTax adapter failed to query affiliates: %v", err)
		return nil, fmt.Errorf("failed to query affiliates: %w", err)
	}
	defer rows.Close()

	var affiliates []*types.Affiliate
	for rows.Next() {
		affiliate := &types.Affiliate{}
		err := rows.Scan(
			&affiliate.ID,
			&affiliate.FirstName,
			&affiliate.LastName,
			&affiliate.Email,
			&affiliate.Phone,
			&affiliate.DefaultCommissionRate,
			&affiliate.StripeConnectAccountID,
			&affiliate.PayoutMethod,
			&affiliate.PayoutThreshold,
			&affiliate.IsActive,
			&affiliate.CreatedAt,
			&affiliate.UpdatedAt,
		)
		if err != nil {
			logger.Errorf("MyWellTax adapter failed to scan affiliate row: %v", err)
			return nil, fmt.Errorf("failed to scan affiliate: %w", err)
		}
		affiliates = append(affiliates, affiliate)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("MyWellTax adapter error iterating affiliate rows: %v", err)
		return nil, fmt.Errorf("error iterating affiliates: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully fetched %d affiliates", len(affiliates))
	return affiliates, nil
}

// GetAffiliateByID retrieves a specific affiliate by ID
func (a *MyWellTaxAdapter) GetAffiliateByID(db *sql.DB, schemaPrefix string, affiliateID string) (*types.Affiliate, error) {
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, email, phone, default_commission_rate,
		       stripe_connect_account_id, payout_method, payout_threshold,
		       is_active, created_at, updated_at
		FROM %s.affiliates
		WHERE id = $1
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter fetching affiliate %s", affiliateID)

	row := db.QueryRow(query, affiliateID)

	affiliate := &types.Affiliate{}
	err := row.Scan(
		&affiliate.ID,
		&affiliate.FirstName,
		&affiliate.LastName,
		&affiliate.Email,
		&affiliate.Phone,
		&affiliate.DefaultCommissionRate,
		&affiliate.StripeConnectAccountID,
		&affiliate.PayoutMethod,
		&affiliate.PayoutThreshold,
		&affiliate.IsActive,
		&affiliate.CreatedAt,
		&affiliate.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("affiliate not found")
		}
		logger.Errorf("MyWellTax adapter failed to get affiliate %s: %v", affiliateID, err)
		return nil, fmt.Errorf("failed to get affiliate: %w", err)
	}

	return affiliate, nil
}

// CreateAffiliate creates a new affiliate
func (a *MyWellTaxAdapter) CreateAffiliate(db *sql.DB, schemaPrefix string, affiliate *types.Affiliate) (*types.Affiliate, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s.affiliates (
			first_name, last_name, email, phone, default_commission_rate,
			payout_method, payout_threshold, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter creating affiliate: %s %s (%s)", affiliate.FirstName, affiliate.LastName, affiliate.Email)

	err := db.QueryRow(
		query,
		affiliate.FirstName,
		affiliate.LastName,
		affiliate.Email,
		affiliate.Phone,
		affiliate.DefaultCommissionRate,
		affiliate.PayoutMethod,
		affiliate.PayoutThreshold,
		affiliate.IsActive,
	).Scan(&affiliate.ID, &affiliate.CreatedAt, &affiliate.UpdatedAt)

	if err != nil {
		logger.Errorf("MyWellTax adapter failed to create affiliate: %v", err)
		return nil, fmt.Errorf("failed to create affiliate: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully created affiliate %s", affiliate.ID)
	return affiliate, nil
}

// UpdateAffiliate updates an existing affiliate
func (a *MyWellTaxAdapter) UpdateAffiliate(db *sql.DB, schemaPrefix string, affiliateID string, affiliate *types.Affiliate) (*types.Affiliate, error) {
	query := fmt.Sprintf(`
		UPDATE %s.affiliates
		SET first_name = $1, last_name = $2, email = $3, phone = $4,
		    default_commission_rate = $5, payout_method = $6,
		    payout_threshold = $7, is_active = $8,
		    updated_at = NOW()
		WHERE id = $9
		RETURNING id, first_name, last_name, email, phone, default_commission_rate,
		          stripe_connect_account_id, payout_method, payout_threshold,
		          is_active, created_at, updated_at
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter updating affiliate %s", affiliateID)

	row := db.QueryRow(
		query,
		affiliate.FirstName,
		affiliate.LastName,
		affiliate.Email,
		affiliate.Phone,
		affiliate.DefaultCommissionRate,
		affiliate.PayoutMethod,
		affiliate.PayoutThreshold,
		affiliate.IsActive,
		affiliateID,
	)

	updated := &types.Affiliate{}
	err := row.Scan(
		&updated.ID,
		&updated.FirstName,
		&updated.LastName,
		&updated.Email,
		&updated.Phone,
		&updated.DefaultCommissionRate,
		&updated.StripeConnectAccountID,
		&updated.PayoutMethod,
		&updated.PayoutThreshold,
		&updated.IsActive,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("affiliate not found")
		}
		logger.Errorf("MyWellTax adapter failed to update affiliate %s: %v", affiliateID, err)
		return nil, fmt.Errorf("failed to update affiliate: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully updated affiliate %s", affiliateID)
	return updated, nil
}

// GetCommissionsByAffiliate retrieves commissions for a specific affiliate (or all if affiliateID is nil)
func (a *MyWellTaxAdapter) GetCommissionsByAffiliate(db *sql.DB, schemaPrefix string, affiliateID *string, status *string, limit int) ([]*types.Commission, error) {
	var whereClause string
	args := []interface{}{}

	// Build WHERE clause dynamically
	conditions := []string{}

	if affiliateID != nil {
		conditions = append(conditions, fmt.Sprintf("c.affiliate_id = $%d", len(args)+1))
		args = append(args, *affiliateID)
	}

	if status != nil {
		conditions = append(conditions, fmt.Sprintf("c.status = $%d", len(args)+1))
		args = append(args, *status)
	}

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.affiliate_id, c.filing_id, c.user_id, c.discount_code_id,
		       c.payment_id, c.order_amount, c.discount_amount, c.net_amount,
		       c.commission_rate, c.commission_amount, c.status,
		       c.approved_at, c.paid_at, c.notes, c.created_at, c.updated_at,
		       u.id, u.first_name, u.last_name, u.email
		FROM %s.commissions c
		JOIN %s.user u ON c.user_id = u.id
		%s
		ORDER BY c.created_at DESC
		LIMIT $%d
	`, schemaPrefix, schemaPrefix, whereClause, len(args)+1)

	args = append(args, limit)

	if affiliateID != nil {
		logger.Infof("MyWellTax adapter fetching commissions for affiliate %s (status=%v, limit=%d)", *affiliateID, status, limit)
	} else {
		logger.Infof("MyWellTax adapter fetching all commissions (status=%v, limit=%d)", status, limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Errorf("MyWellTax adapter failed to query commissions: %v", err)
		return nil, fmt.Errorf("failed to query commissions: %w", err)
	}
	defer rows.Close()

	var commissions []*types.Commission
	for rows.Next() {
		commission := &types.Commission{
			Customer: &types.CustomerInfo{},
		}
		err := rows.Scan(
			&commission.ID,
			&commission.AffiliateID,
			&commission.FilingID,
			&commission.UserID,
			&commission.DiscountCodeID,
			&commission.PaymentID,
			&commission.OrderAmount,
			&commission.DiscountAmount,
			&commission.NetAmount,
			&commission.CommissionRate,
			&commission.CommissionAmount,
			&commission.Status,
			&commission.ApprovedAt,
			&commission.PaidAt,
			&commission.Notes,
			&commission.CreatedAt,
			&commission.UpdatedAt,
			&commission.Customer.ID,
			&commission.Customer.FirstName,
			&commission.Customer.LastName,
			&commission.Customer.Email,
		)
		if err != nil {
			logger.Errorf("MyWellTax adapter failed to scan commission row: %v", err)
			return nil, fmt.Errorf("failed to scan commission: %w", err)
		}
		commissions = append(commissions, commission)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("MyWellTax adapter error iterating commission rows: %v", err)
		return nil, fmt.Errorf("error iterating commissions: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully fetched %d commissions", len(commissions))
	return commissions, nil
}

// GetAffiliateStats calculates aggregate statistics for an affiliate
func (a *MyWellTaxAdapter) GetAffiliateStats(db *sql.DB, schemaPrefix string, affiliateID string) (*types.AffiliateStats, error) {
	query := fmt.Sprintf(`
		SELECT
			-- Clicks
			COALESCE((SELECT COUNT(*) FROM %s.affiliate_clicks WHERE affiliate_id = $1), 0) as total_clicks,

			-- Conversions (commissions)
			COALESCE(COUNT(c.id), 0) as total_conversions,

			-- Commission totals by status
			COALESCE(SUM(CASE WHEN c.status = 'PENDING' THEN c.commission_amount ELSE 0 END), 0) as pending_commissions,
			COALESCE(SUM(CASE WHEN c.status = 'APPROVED' THEN c.commission_amount ELSE 0 END), 0) as approved_commissions,
			COALESCE(SUM(CASE WHEN c.status = 'PAID' THEN c.commission_amount ELSE 0 END), 0) as paid_commissions,
			COALESCE(SUM(CASE WHEN c.status = 'CANCELLED' THEN c.commission_amount ELSE 0 END), 0) as cancelled_commissions,

			-- Total earned (all except cancelled)
			COALESCE(SUM(CASE WHEN c.status != 'CANCELLED' THEN c.commission_amount ELSE 0 END), 0) as total_earned,

			-- Revenue metrics
			COALESCE(SUM(c.order_amount), 0) as total_revenue
		FROM %s.commissions c
		WHERE c.affiliate_id = $1
	`, schemaPrefix, schemaPrefix)

	logger.Infof("MyWellTax adapter calculating stats for affiliate %s", affiliateID)

	stats := &types.AffiliateStats{
		AffiliateID: uuid.MustParse(affiliateID),
	}

	err := db.QueryRow(query, affiliateID).Scan(
		&stats.TotalClicks,
		&stats.TotalConversions,
		&stats.PendingCommissions,
		&stats.ApprovedCommissions,
		&stats.PaidCommissions,
		&stats.CancelledCommissions,
		&stats.TotalCommissionsEarned,
		&stats.TotalRevenue,
	)

	if err != nil {
		logger.Errorf("MyWellTax adapter failed to calculate stats: %v", err)
		return nil, fmt.Errorf("failed to calculate stats: %w", err)
	}

	// Calculate conversion rate
	if stats.TotalClicks > 0 {
		stats.ConversionRate = (float64(stats.TotalConversions) / float64(stats.TotalClicks)) * 100
	}

	stats.TotalOrders = stats.TotalConversions

	logger.Infof("MyWellTax adapter successfully calculated stats for affiliate %s", affiliateID)
	return stats, nil
}

// ApproveCommission approves a pending commission
func (a *MyWellTaxAdapter) ApproveCommission(db *sql.DB, schemaPrefix string, commissionID string) (*types.Commission, error) {
	query := fmt.Sprintf(`
		UPDATE %s.commissions
		SET status = 'APPROVED', approved_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING id, affiliate_id, filing_id, user_id, discount_code_id, payment_id,
		          order_amount, discount_amount, net_amount, commission_rate,
		          commission_amount, status, approved_at, paid_at, notes,
		          created_at, updated_at
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter approving commission %s", commissionID)

	commission := &types.Commission{}
	err := db.QueryRow(query, commissionID).Scan(
		&commission.ID,
		&commission.AffiliateID,
		&commission.FilingID,
		&commission.UserID,
		&commission.DiscountCodeID,
		&commission.PaymentID,
		&commission.OrderAmount,
		&commission.DiscountAmount,
		&commission.NetAmount,
		&commission.CommissionRate,
		&commission.CommissionAmount,
		&commission.Status,
		&commission.ApprovedAt,
		&commission.PaidAt,
		&commission.Notes,
		&commission.CreatedAt,
		&commission.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("commission not found or not pending")
		}
		logger.Errorf("MyWellTax adapter failed to approve commission %s: %v", commissionID, err)
		return nil, fmt.Errorf("failed to approve commission: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully approved commission %s", commissionID)
	return commission, nil
}

// MarkCommissionPaid marks an approved commission as paid
func (a *MyWellTaxAdapter) MarkCommissionPaid(db *sql.DB, schemaPrefix string, commissionID string) (*types.Commission, error) {
	query := fmt.Sprintf(`
		UPDATE %s.commissions
		SET status = 'PAID', paid_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'APPROVED'
		RETURNING id, affiliate_id, filing_id, user_id, discount_code_id, payment_id,
		          order_amount, discount_amount, net_amount, commission_rate,
		          commission_amount, status, approved_at, paid_at, notes,
		          created_at, updated_at
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter marking commission %s as paid", commissionID)

	commission := &types.Commission{}
	err := db.QueryRow(query, commissionID).Scan(
		&commission.ID,
		&commission.AffiliateID,
		&commission.FilingID,
		&commission.UserID,
		&commission.DiscountCodeID,
		&commission.PaymentID,
		&commission.OrderAmount,
		&commission.DiscountAmount,
		&commission.NetAmount,
		&commission.CommissionRate,
		&commission.CommissionAmount,
		&commission.Status,
		&commission.ApprovedAt,
		&commission.PaidAt,
		&commission.Notes,
		&commission.CreatedAt,
		&commission.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("commission not found or not approved")
		}
		logger.Errorf("MyWellTax adapter failed to mark commission %s as paid: %v", commissionID, err)
		return nil, fmt.Errorf("failed to mark commission as paid: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully marked commission %s as paid", commissionID)
	return commission, nil
}

// CancelCommission cancels a commission with a reason
func (a *MyWellTaxAdapter) CancelCommission(db *sql.DB, schemaPrefix string, commissionID string, reason string) (*types.Commission, error) {
	query := fmt.Sprintf(`
		UPDATE %s.commissions
		SET status = 'CANCELLED', notes = $2, updated_at = NOW()
		WHERE id = $1 AND status IN ('PENDING', 'APPROVED')
		RETURNING id, affiliate_id, filing_id, user_id, discount_code_id, payment_id,
		          order_amount, discount_amount, net_amount, commission_rate,
		          commission_amount, status, approved_at, paid_at, notes,
		          created_at, updated_at
	`, schemaPrefix)

	logger.Infof("MyWellTax adapter cancelling commission %s with reason: %s", commissionID, reason)

	commission := &types.Commission{}
	err := db.QueryRow(query, commissionID, reason).Scan(
		&commission.ID,
		&commission.AffiliateID,
		&commission.FilingID,
		&commission.UserID,
		&commission.DiscountCodeID,
		&commission.PaymentID,
		&commission.OrderAmount,
		&commission.DiscountAmount,
		&commission.NetAmount,
		&commission.CommissionRate,
		&commission.CommissionAmount,
		&commission.Status,
		&commission.ApprovedAt,
		&commission.PaidAt,
		&commission.Notes,
		&commission.CreatedAt,
		&commission.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("commission not found or already paid/cancelled")
		}
		logger.Errorf("MyWellTax adapter failed to cancel commission %s: %v", commissionID, err)
		return nil, fmt.Errorf("failed to cancel commission: %w", err)
	}

	logger.Infof("MyWellTax adapter successfully cancelled commission %s", commissionID)
	return commission, nil
}
