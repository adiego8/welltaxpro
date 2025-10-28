package adapter

import (
	"database/sql"
	"fmt"
	"time"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
)

// CreateDocument creates a new document record in the tenant's database
func (a *MyWellTaxAdapter) CreateDocument(db *sql.DB, schemaPrefix string, document *types.Document) (*types.Document, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s.document (id, user_id, name, file_path, type, filing_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, name, file_path, type, filing_id, created_at, updated_at
	`, schemaPrefix)

	logger.Infof("Creating document in %s.document", schemaPrefix)

	// Generate ID if not provided
	if document.ID == uuid.Nil {
		document.ID = uuid.New()
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	document.CreatedAt = now

	var filingID *uuid.UUID
	var createdAt, updatedAt string
	var updatedAtPtr *string

	err := db.QueryRow(
		query,
		document.ID,
		document.UserID,
		document.Name,
		document.FilePath,
		document.Type,
		document.FilingID,
		document.CreatedAt,
		document.UpdatedAt,
	).Scan(
		&document.ID,
		&document.UserID,
		&document.Name,
		&document.FilePath,
		&document.Type,
		&filingID,
		&createdAt,
		&updatedAtPtr,
	)

	if err != nil {
		logger.Errorf("Failed to create document: %v", err)
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	document.FilingID = filingID
	document.CreatedAt = createdAt
	if updatedAtPtr != nil {
		updatedAt = *updatedAtPtr
		document.UpdatedAt = &updatedAt
	}

	logger.Infof("Successfully created document: %s", document.ID)
	return document, nil
}

// GetDocumentByID retrieves a specific document by ID
func (a *MyWellTaxAdapter) GetDocumentByID(db *sql.DB, schemaPrefix string, documentID string) (*types.Document, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, name, file_path, type, filing_id, created_at, updated_at
		FROM %s.document
		WHERE id = $1
	`, schemaPrefix)

	logger.Infof("Fetching document %s from %s.document", documentID, schemaPrefix)

	var document types.Document
	var filingID *uuid.UUID
	var updatedAtPtr *string

	err := db.QueryRow(query, documentID).Scan(
		&document.ID,
		&document.UserID,
		&document.Name,
		&document.FilePath,
		&document.Type,
		&filingID,
		&document.CreatedAt,
		&updatedAtPtr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Errorf("Document not found: %s", documentID)
			return nil, fmt.Errorf("document not found")
		}
		logger.Errorf("Failed to fetch document: %v", err)
		return nil, fmt.Errorf("failed to fetch document: %w", err)
	}

	document.FilingID = filingID
	if updatedAtPtr != nil {
		document.UpdatedAt = updatedAtPtr
	}

	return &document, nil
}

// GetDocumentsByFilingID retrieves all documents associated with a filing
func (a *MyWellTaxAdapter) GetDocumentsByFilingID(db *sql.DB, schemaPrefix string, filingID string) ([]*types.Document, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, name, file_path, type, filing_id, created_at, updated_at
		FROM %s.document
		WHERE filing_id = $1
		ORDER BY created_at DESC
	`, schemaPrefix)

	logger.Infof("Fetching documents for filing %s from %s.document", filingID, schemaPrefix)

	rows, err := db.Query(query, filingID)
	if err != nil {
		logger.Errorf("Failed to query documents: %v", err)
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	documents := make([]*types.Document, 0)
	for rows.Next() {
		var document types.Document
		var filingIDPtr *uuid.UUID
		var updatedAtPtr *string

		if err := rows.Scan(
			&document.ID,
			&document.UserID,
			&document.Name,
			&document.FilePath,
			&document.Type,
			&filingIDPtr,
			&document.CreatedAt,
			&updatedAtPtr,
		); err != nil {
			logger.Errorf("Failed to scan document: %v", err)
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}

		document.FilingID = filingIDPtr
		if updatedAtPtr != nil {
			document.UpdatedAt = updatedAtPtr
		}

		documents = append(documents, &document)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("Error iterating documents: %v", err)
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	logger.Infof("Found %d documents for filing %s", len(documents), filingID)
	return documents, nil
}

// DeleteDocument removes a document record from the tenant's database
func (a *MyWellTaxAdapter) DeleteDocument(db *sql.DB, schemaPrefix string, documentID string) error {
	query := fmt.Sprintf(`
		DELETE FROM %s.document
		WHERE id = $1
	`, schemaPrefix)

	logger.Infof("Deleting document %s from %s.document", documentID, schemaPrefix)

	result, err := db.Exec(query, documentID)
	if err != nil {
		logger.Errorf("Failed to delete document: %v", err)
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("Failed to get rows affected: %v", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Errorf("Document not found: %s", documentID)
		return fmt.Errorf("document not found")
	}

	logger.Infof("Successfully deleted document: %s", documentID)
	return nil
}
