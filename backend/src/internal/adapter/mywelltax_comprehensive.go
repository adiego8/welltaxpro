package adapter

import (
	"database/sql"
	"fmt"
	"welltaxpro/src/internal/crypto"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// GetClientComprehensive retrieves all data related to a MyWellTax client
func (a *MyWellTaxAdapter) GetClientComprehensive(db *sql.DB, schemaPrefix string, clientID string) (*types.ClientComprehensive, error) {
	logger.Infof("MyWellTax adapter fetching comprehensive data for client %s", clientID)

	comprehensive := &types.ClientComprehensive{}

	// 1. Get basic client info
	client, err := a.GetClientByID(db, schemaPrefix, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	comprehensive.Client = client

	// 2. Get spouse (optional)
	spouse, _ := a.getSpouse(db, schemaPrefix, clientID)
	comprehensive.Spouse = spouse

	// 3. Get dependents (optional)
	dependents, _ := a.getDependents(db, schemaPrefix, clientID)
	comprehensive.Dependents = dependents

	// 4. Get all filings with related data
	filings, _ := a.getFilingsWithRelatedData(db, schemaPrefix, clientID)
	comprehensive.Filings = filings

	logger.Infof("Successfully fetched comprehensive data for client %s (%d filings)", clientID, len(comprehensive.Filings))
	return comprehensive, nil
}

func (a *MyWellTaxAdapter) getSpouse(db *sql.DB, schemaPrefix string, clientID string) (*types.Spouse, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, first_name, middle_name, last_name, email, phone, dob, ssn, is_death, death_date, created_at
		FROM %s.spouse WHERE user_id = $1 LIMIT 1
	`, schemaPrefix)

	row := db.QueryRow(query, clientID)
	spouse := &types.Spouse{}
	var ssnEncrypted string
	err := row.Scan(&spouse.ID, &spouse.UserID, &spouse.FirstName, &spouse.MiddleName, &spouse.LastName, &spouse.Email, &spouse.Phone, &spouse.Dob, &ssnEncrypted, &spouse.IsDeath, &spouse.DeathDate, &spouse.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Mask SSN for API response
	spouse.Ssn = crypto.MaskSSN(ssnEncrypted)
	return spouse, nil
}

func (a *MyWellTaxAdapter) getDependents(db *sql.DB, schemaPrefix string, clientID string) ([]*types.Dependent, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, first_name, middle_name, last_name, dob, ssn, relationship, time_with_applicant, exclusive_claim, created_at, updated_at
		FROM %s.dependent WHERE user_id = $1
	`, schemaPrefix)

	rows, err := db.Query(query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dependents []*types.Dependent
	for rows.Next() {
		dep := &types.Dependent{}
		var ssnEncrypted string
		if err := rows.Scan(&dep.ID, &dep.UserID, &dep.FirstName, &dep.MiddleName, &dep.LastName, &dep.Dob, &ssnEncrypted, &dep.Relationship, &dep.TimeWithApplicant, &dep.ExclusiveClaim, &dep.CreatedAt, &dep.UpdatedAt); err != nil {
			return nil, err
		}
		// Mask SSN for API response
		dep.Ssn = crypto.MaskSSN(ssnEncrypted)

		// Fetch required document types for this dependent
		docs, err := a.getDependentDocuments(db, schemaPrefix, dep.ID)
		if err != nil {
			logger.Warningf("Failed to get dependent documents for %s: %v", dep.ID, err)
		} else {
			dep.Documents = docs
		}

		dependents = append(dependents, dep)
	}
	return dependents, rows.Err()
}

// getDependentDocuments retrieves the list of required document types for a dependent
func (a *MyWellTaxAdapter) getDependentDocuments(db *sql.DB, schemaPrefix string, dependentID uuid.UUID) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT record_name
		FROM %s.dependent_document_map
		WHERE dependent_id = $1
		ORDER BY created_at
	`, schemaPrefix)

	rows, err := db.Query(query, dependentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []string
	for rows.Next() {
		var recordName string
		if err := rows.Scan(&recordName); err != nil {
			return nil, err
		}
		documents = append(documents, recordName)
	}
	return documents, rows.Err()
}

func (a *MyWellTaxAdapter) getFilingsWithRelatedData(db *sql.DB, schemaPrefix string, clientID string) ([]*types.Filing, error) {
	query := fmt.Sprintf(`
		SELECT id, year, user_id, marital_status, spouse, source_of_income, deductions, income, marketplace_insurance, created_at, updated_at
		FROM %s.filing WHERE user_id = $1 ORDER BY year DESC
	`, schemaPrefix)

	logger.Infof("Fetching filings for client %s with query: %s", clientID, query)

	rows, err := db.Query(query, clientID)
	if err != nil {
		logger.Errorf("Failed to query filings: %v", err)
		return nil, err
	}
	defer rows.Close()

	var filings []*types.Filing
	for rows.Next() {
		filing := &types.Filing{}
		err := rows.Scan(&filing.ID, &filing.Year, &filing.UserID, &filing.MaritalStatus, &filing.SpouseID, pq.Array(&filing.SourceOfIncome), pq.Array(&filing.Deductions), &filing.Income, &filing.MarketplaceInsurance, &filing.CreatedAt, &filing.UpdatedAt)
		if err != nil {
			logger.Errorf("Failed to scan filing row: %v", err)
			return nil, err
		}

		logger.Infof("Found filing: year=%d, id=%s", filing.Year, filing.ID)

		// Fetch related data with error logging
		filing.Status, err = a.getFilingStatus(db, schemaPrefix, filing.ID)
		if err != nil {
			logger.Warningf("Failed to get filing status for %s: %v", filing.ID, err)
		}

		filing.Documents, err = a.getFilingDocuments(db, schemaPrefix, filing.ID)
		if err != nil {
			logger.Warningf("Failed to get filing documents for %s: %v", filing.ID, err)
		}

		filing.Properties, err = a.getFilingProperties(db, schemaPrefix, filing.ID)
		if err != nil {
			logger.Warningf("Failed to get filing properties for %s: %v", filing.ID, err)
		}

		filing.IRAContributions, err = a.getFilingIRAContributions(db, schemaPrefix, filing.ID)
		if err != nil {
			logger.Warningf("Failed to get IRA contributions for %s: %v", filing.ID, err)
		}

		filing.Charities, err = a.getFilingCharities(db, schemaPrefix, filing.ID)
		if err != nil {
			logger.Warningf("Failed to get charities for %s: %v", filing.ID, err)
		}

		filing.Childcares, err = a.getFilingChildcares(db, schemaPrefix, filing.ID)
		if err != nil {
			logger.Warningf("Failed to get childcares for %s: %v", filing.ID, err)
		}

		filing.Payments, err = a.getFilingPayments(db, schemaPrefix, filing.ID)
		if err != nil {
			logger.Warningf("Failed to get payments for %s: %v", filing.ID, err)
		}

		filing.Discounts, err = a.getFilingDiscounts(db, schemaPrefix, filing.ID)
		if err != nil {
			logger.Warningf("Failed to get discounts for %s: %v", filing.ID, err)
		}

		filings = append(filings, filing)
	}

	logger.Infof("Fetched %d filings for client %s", len(filings), clientID)
	return filings, rows.Err()
}

func (a *MyWellTaxAdapter) getFilingStatus(db *sql.DB, schemaPrefix string, filingID uuid.UUID) (*types.FilingStatus, error) {
	query := fmt.Sprintf(`SELECT id, filing_id, latest_step, is_completed, status FROM %s.filing_status WHERE filing_id = $1`, schemaPrefix)
	row := db.QueryRow(query, filingID)
	status := &types.FilingStatus{}
	err := row.Scan(&status.ID, &status.FilingID, &status.LatestStep, &status.IsCompleted, &status.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return status, err
}

func (a *MyWellTaxAdapter) getFilingDocuments(db *sql.DB, schemaPrefix string, filingID uuid.UUID) ([]*types.Document, error) {
	query := fmt.Sprintf(`SELECT id, user_id, filing_id, name, file_path, type, created_at, updated_at FROM %s.document WHERE filing_id = $1`, schemaPrefix)
	rows, err := db.Query(query, filingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []*types.Document
	for rows.Next() {
		doc := &types.Document{}
		if err := rows.Scan(&doc.ID, &doc.UserID, &doc.FilingID, &doc.Name, &doc.FilePath, &doc.Type, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, rows.Err()
}

func (a *MyWellTaxAdapter) getFilingProperties(db *sql.DB, schemaPrefix string, filingID uuid.UUID) ([]*types.Property, error) {
	query := fmt.Sprintf(`
		SELECT p.id, p.user_id, p.address1, p.address2, p.state, p.city, p.zipcode, p.purchase_price, p.closing_cost, p.purchase_date, p.rents, p.royalties, p.updated_at, p.created_at
		FROM %s.property p JOIN %s.filing_property_map fpm ON fpm.property_id = p.id WHERE fpm.filing_id = $1
	`, schemaPrefix, schemaPrefix)

	rows, err := db.Query(query, filingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []*types.Property
	for rows.Next() {
		prop := &types.Property{}
		if err := rows.Scan(&prop.ID, &prop.UserID, &prop.Address1, &prop.Address2, &prop.State, &prop.City, &prop.Zipcode, &prop.PurchasePrice, &prop.ClosingCost, &prop.PurchaseDate, &prop.Rents, &prop.Royalties, &prop.UpdatedAt, &prop.CreatedAt); err != nil {
			return nil, err
		}
		prop.Expenses, _ = a.getPropertyExpenses(db, schemaPrefix, prop.ID)
		properties = append(properties, prop)
	}
	return properties, rows.Err()
}

func (a *MyWellTaxAdapter) getPropertyExpenses(db *sql.DB, schemaPrefix string, propertyID uuid.UUID) ([]*types.Expense, error) {
	query := fmt.Sprintf(`SELECT id, property_id, name, amount, created_at FROM %s.expense WHERE property_id = $1`, schemaPrefix)
	rows, err := db.Query(query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*types.Expense
	for rows.Next() {
		exp := &types.Expense{}
		if err := rows.Scan(&exp.ID, &exp.PropertyID, &exp.Name, &exp.Amount, &exp.CreatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, exp)
	}
	return expenses, rows.Err()
}

func (a *MyWellTaxAdapter) getFilingIRAContributions(db *sql.DB, schemaPrefix string, filingID uuid.UUID) ([]*types.IRAContribution, error) {
	query := fmt.Sprintf(`SELECT id, filing_id, account_type, amount FROM %s.ira_contribution WHERE filing_id = $1`, schemaPrefix)
	rows, err := db.Query(query, filingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributions []*types.IRAContribution
	for rows.Next() {
		ira := &types.IRAContribution{}
		if err := rows.Scan(&ira.ID, &ira.FilingID, &ira.AccountType, &ira.Amount); err != nil {
			return nil, err
		}
		contributions = append(contributions, ira)
	}
	return contributions, rows.Err()
}

func (a *MyWellTaxAdapter) getFilingCharities(db *sql.DB, schemaPrefix string, filingID uuid.UUID) ([]*types.Charity, error) {
	query := fmt.Sprintf(`SELECT id, user_id, filing_id, name, contribution FROM %s.charity WHERE filing_id = $1`, schemaPrefix)
	rows, err := db.Query(query, filingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charities []*types.Charity
	for rows.Next() {
		charity := &types.Charity{}
		if err := rows.Scan(&charity.ID, &charity.UserID, &charity.FilingID, &charity.Name, &charity.Contribution); err != nil {
			return nil, err
		}
		charities = append(charities, charity)
	}
	return charities, rows.Err()
}

func (a *MyWellTaxAdapter) getFilingChildcares(db *sql.DB, schemaPrefix string, filingID uuid.UUID) ([]*types.Childcare, error) {
	query := fmt.Sprintf(`
		SELECT c.id, c.user_id, c.name, c.amount, c.tax_id, c.address1, c.address2, c.city, c.state, c.zipcode
		FROM %s.childcare c JOIN %s.filing_childcare_map fcm ON fcm.childcare_id = c.id WHERE fcm.filing_id = $1
	`, schemaPrefix, schemaPrefix)

	rows, err := db.Query(query, filingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var childcares []*types.Childcare
	for rows.Next() {
		cc := &types.Childcare{}
		if err := rows.Scan(&cc.ID, &cc.UserID, &cc.Name, &cc.Amount, &cc.TaxID, &cc.Address1, &cc.Address2, &cc.City, &cc.State, &cc.Zipcode); err != nil {
			return nil, err
		}
		childcares = append(childcares, cc)
	}
	return childcares, rows.Err()
}

func (a *MyWellTaxAdapter) getFilingPayments(db *sql.DB, schemaPrefix string, filingID uuid.UUID) ([]*types.Payment, error) {
	query := fmt.Sprintf(`
		SELECT id, filing_id, stripe_session_id, amount, original_amount, discount_amount, discount_code, status, created_at, updated_at
		FROM %s.payment WHERE filing_id = $1 ORDER BY created_at DESC
	`, schemaPrefix)

	rows, err := db.Query(query, filingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*types.Payment
	for rows.Next() {
		payment := &types.Payment{}
		var amountCents float64
		var originalAmountCents, discountAmountCents *float64

		if err := rows.Scan(&payment.ID, &payment.FilingID, &payment.StripeSessionID, &amountCents, &originalAmountCents, &discountAmountCents, &payment.DiscountCode, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt); err != nil {
			return nil, err
		}

		// Convert cents to dollars (data is stored as cents but in decimal format)
		payment.Amount = amountCents / 100.0
		if originalAmountCents != nil {
			dollars := *originalAmountCents / 100.0
			payment.OriginalAmount = &dollars
		}
		if discountAmountCents != nil {
			dollars := *discountAmountCents / 100.0
			payment.DiscountAmount = &dollars
		}

		payment.Items, _ = a.getPaymentItems(db, schemaPrefix, payment.ID)
		payments = append(payments, payment)
	}
	return payments, rows.Err()
}

func (a *MyWellTaxAdapter) getPaymentItems(db *sql.DB, schemaPrefix string, paymentID uuid.UUID) ([]*types.PaymentItem, error) {
	query := fmt.Sprintf(`SELECT id, payment_id, price_id, name, quantity, unit_amount FROM %s.payment_item WHERE payment_id = $1`, schemaPrefix)
	rows, err := db.Query(query, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*types.PaymentItem
	for rows.Next() {
		item := &types.PaymentItem{}
		var unitAmountCents float64
		if err := rows.Scan(&item.ID, &item.PaymentID, &item.PriceID, &item.Name, &item.Quantity, &unitAmountCents); err != nil {
			return nil, err
		}
		// Convert cents to dollars (data is stored as cents but in decimal format)
		item.UnitAmount = unitAmountCents / 100.0
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *MyWellTaxAdapter) getFilingDiscounts(db *sql.DB, schemaPrefix string, filingID uuid.UUID) ([]*types.FilingDiscount, error) {
	query := fmt.Sprintf(`
		SELECT fd.id, fd.filing_id, fd.discount_code_id, fd.original_amount, fd.discount_amount, fd.final_amount, fd.applied_at, dc.code
		FROM %s.filing_discounts fd LEFT JOIN %s.discount_codes dc ON dc.id = fd.discount_code_id WHERE fd.filing_id = $1
	`, schemaPrefix, schemaPrefix)

	rows, err := db.Query(query, filingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discounts []*types.FilingDiscount
	for rows.Next() {
		discount := &types.FilingDiscount{}
		var originalAmountCents, discountAmountCents, finalAmountCents int64
		if err := rows.Scan(&discount.ID, &discount.FilingID, &discount.DiscountCodeID, &originalAmountCents, &discountAmountCents, &finalAmountCents, &discount.AppliedAt, &discount.Code); err != nil {
			return nil, err
		}
		// Convert cents to dollars
		discount.OriginalAmount = float64(originalAmountCents) / 100.0
		discount.DiscountAmount = float64(discountAmountCents) / 100.0
		discount.FinalAmount = float64(finalAmountCents) / 100.0
		discounts = append(discounts, discount)
	}
	return discounts, rows.Err()
}
