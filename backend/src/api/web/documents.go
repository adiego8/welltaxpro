package webapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"welltaxpro/src/internal/storage"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
)

// uploadDocument handles document upload for a filing (admin only)
func (api *API) uploadDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	filingID := vars["filingId"]

	logger.Infof("Upload document request for filing %s in tenant %s", filingID, tenantID)

	// Parse multipart form with max size
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.Errorf("Failed to parse multipart form: %v", err)
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logger.Errorf("Failed to get file from form: %v", err)
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	documentType := r.FormValue("type")
	if documentType == "" {
		http.Error(w, "Document type is required", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("userId")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Validate user ID and filing ID are valid UUIDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	filingUUID, err := uuid.Parse(filingID)
	if err != nil {
		http.Error(w, "Invalid filing ID", http.StatusBadRequest)
		return
	}

	// Get tenant config for storage settings
	tc, err := api.store.GetTenantConfig(tenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant config: %v", err)
		http.Error(w, "Failed to get tenant configuration", http.StatusInternalServerError)
		return
	}

	// Create storage provider using factory (handles Secret Manager, file, or ADC)
	storageProvider, err := storage.NewStorageProviderForTenant(context.Background(), tc)
	if err != nil {
		logger.Errorf("Failed to create storage provider: %v", err)
		http.Error(w, "Failed to initialize storage", http.StatusInternalServerError)
		return
	}

	// Calculate file hash for deduplication
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("Failed to read file: %v", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	hasher := sha256.New()
	hasher.Write(fileBytes)
	fileHash := hex.EncodeToString(hasher.Sum(nil))[:16] // Use first 16 chars

	// Generate storage path: {userId}/{type}/{filename_hash}.ext
	ext := filepath.Ext(header.Filename)
	baseName := strings.TrimSuffix(header.Filename, ext)
	storagePath := fmt.Sprintf("%s/%s/%s_%s%s", userID, documentType, baseName, fileHash, ext)

	// Upload to GCS
	fileReader := strings.NewReader(string(fileBytes))
	metadata := map[string]string{
		"tenant_id":     tenantID,
		"filing_id":     filingID,
		"user_id":       userID,
		"document_type": documentType,
		"original_name": header.Filename,
	}

	if err := storageProvider.Upload(context.Background(), tc.StorageBucket, storagePath, fileReader, metadata); err != nil {
		logger.Errorf("Failed to upload to storage: %v", err)
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// Create document record in database
	document := &types.Document{
		ID:       uuid.New(),
		UserID:   userUUID,
		FilingID: &filingUUID,
		Name:     header.Filename,
		FilePath: storagePath,
		Type:     documentType,
	}

	createdDoc, err := api.store.CreateDocument(tenantID, document)
	if err != nil {
		logger.Errorf("Failed to create document record: %v", err)
		// Try to clean up uploaded file
		storageProvider.Delete(context.Background(), tc.StorageBucket, storagePath)
		http.Error(w, "Failed to create document record", http.StatusInternalServerError)
		return
	}

	logger.Infof("Successfully uploaded document %s", createdDoc.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdDoc); err != nil {
		logger.Errorf("Failed to encode document response: %v", err)
	}
}

// getDocuments returns all documents for a filing (admin only)
func (api *API) getDocuments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	filingID := vars["filingId"]

	logger.Infof("Fetching documents for filing %s in tenant %s", filingID, tenantID)

	documents, err := api.store.GetDocumentsByFilingID(tenantID, filingID)
	if err != nil {
		logger.Errorf("Failed to get documents: %v", err)
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(documents); err != nil {
		logger.Errorf("Failed to encode documents response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// downloadDocument generates a signed URL for document download (admin only)
func (api *API) downloadDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	documentID := vars["documentId"]

	logger.Infof("Download request for document %s in tenant %s", documentID, tenantID)

	// Get document record
	document, err := api.store.GetDocumentByID(tenantID, documentID)
	if err != nil {
		logger.Errorf("Failed to get document: %v", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Get tenant config for storage settings
	tc, err := api.store.GetTenantConfig(tenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant config: %v", err)
		http.Error(w, "Failed to get tenant configuration", http.StatusInternalServerError)
		return
	}

	// Create storage provider using factory (handles Secret Manager, file, or ADC)
	storageProvider, err := storage.NewStorageProviderForTenant(context.Background(), tc)
	if err != nil {
		logger.Errorf("Failed to create storage provider: %v", err)
		http.Error(w, "Failed to initialize storage", http.StatusInternalServerError)
		return
	}

	// Generate signed URL (valid for 15 minutes)
	signedURL, err := storageProvider.GetSignedURL(context.Background(), tc.StorageBucket, document.FilePath, 15*time.Minute)
	if err != nil {
		logger.Errorf("Failed to generate signed URL: %v", err)
		http.Error(w, "Failed to generate download URL", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"url":       signedURL,
		"expiresIn": "15m",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode download response: %v", err)
	}
}

// deleteDocument removes a document and its storage file (admin only)
func (api *API) deleteDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	documentID := vars["documentId"]

	logger.Infof("Delete request for document %s in tenant %s", documentID, tenantID)

	// Get document record first (need file path for storage deletion)
	document, err := api.store.GetDocumentByID(tenantID, documentID)
	if err != nil {
		logger.Errorf("Failed to get document: %v", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Get tenant config for storage settings
	tc, err := api.store.GetTenantConfig(tenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant config: %v", err)
		http.Error(w, "Failed to get tenant configuration", http.StatusInternalServerError)
		return
	}

	// Create storage provider using factory (handles Secret Manager, file, or ADC)
	storageProvider, err := storage.NewStorageProviderForTenant(context.Background(), tc)
	if err != nil {
		logger.Errorf("Failed to create storage provider: %v", err)
		http.Error(w, "Failed to initialize storage", http.StatusInternalServerError)
		return
	}

	// Delete from storage
	if err := storageProvider.Delete(context.Background(), tc.StorageBucket, document.FilePath); err != nil {
		logger.Errorf("Failed to delete from storage: %v", err)
		// Continue anyway - database record is more important
	}

	// Delete database record
	if err := api.store.DeleteDocument(tenantID, documentID); err != nil {
		logger.Errorf("Failed to delete document record: %v", err)
		http.Error(w, "Failed to delete document", http.StatusInternalServerError)
		return
	}

	logger.Infof("Successfully deleted document %s", documentID)
	w.WriteHeader(http.StatusNoContent)
}
