package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"welltaxpro/src/internal/signature"

	"github.com/google/logger"
	"github.com/gorilla/mux"
)

// SignatureRequest represents the request body for signature endpoint
type SignatureRequest struct {
	PDFPath            string   `json:"pdfPath"`
	TaxPayerEmail      string   `json:"taxPayerEmail"`
	TaxPayerName       string   `json:"taxPayerName"`
	TaxPayerSsn        string   `json:"taxPayerSsn"`
	SpouseName         string   `json:"spouseName,omitempty"`
	SpouseEmail        string   `json:"spouseEmail,omitempty"`
	TaxPayerSpouseName *string  `json:"taxPayerSpouseName,omitempty"`
	TaxPayerSpouseSsn  *string  `json:"taxPayerSpouseSsn,omitempty"`
	GrossIncome        float64  `json:"grossIncome"`
	TotalTax           float64  `json:"totalTax"`
	TaxWithHeld        float64  `json:"taxWithHeld"`
	Refund             float64  `json:"refund"`
	Owed               float64  `json:"owed"`
	SignatureDate      *string  `json:"signatureDate,omitempty"`
	SpouseSignature    bool     `json:"spouseSignature"`
}

// sendSignatureRequest sends a document to DocuSign for signature (admin only)
func (api *API) sendSignatureRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	logger.Infof("Signature request for tenant %s", tenantID)

	// Parse request body
	var req SignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to parse request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PDFPath == "" {
		http.Error(w, "PDF path is required", http.StatusBadRequest)
		return
	}
	if req.TaxPayerEmail == "" || req.TaxPayerName == "" || req.TaxPayerSsn == "" {
		http.Error(w, "Taxpayer information is required", http.StatusBadRequest)
		return
	}
	if req.SpouseSignature && (req.SpouseEmail == "" || req.SpouseName == "") {
		http.Error(w, "Spouse information is required when spouse signature is needed", http.StatusBadRequest)
		return
	}

	// Get tenant config for DocuSign settings
	tc, err := api.store.GetTenantConfig(tenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant config: %v", err)
		http.Error(w, "Failed to get tenant configuration", http.StatusInternalServerError)
		return
	}

	// Create signature request
	sig := &signature.Signature{
		TaxPayerEmail:      req.TaxPayerEmail,
		TaxPayerName:       req.TaxPayerName,
		TaxPayerSsn:        req.TaxPayerSsn,
		SpouseName:         req.SpouseName,
		SpouseEmail:        req.SpouseEmail,
		TaxPayerSpouseName: req.TaxPayerSpouseName,
		TaxPayerSpouseSsn:  req.TaxPayerSpouseSsn,
		GrossIncome:        req.GrossIncome,
		TotalTax:           req.TotalTax,
		TaxWithHeld:        req.TaxWithHeld,
		Refund:             req.Refund,
		Owed:               req.Owed,
		SignatureDate:      req.SignatureDate,
		SpouseSignature:    req.SpouseSignature,
	}

	// Send to DocuSign
	if err := signature.SignDocument(context.Background(), tc, req.PDFPath, sig); err != nil {
		logger.Errorf("Failed to send signature request: %v", err)
		http.Error(w, "Failed to send signature request", http.StatusInternalServerError)
		return
	}

	logger.Infof("Successfully sent signature request for tenant %s", tenantID)

	// Return success response
	response := map[string]string{
		"status":  "sent",
		"message": "Signature request sent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}
