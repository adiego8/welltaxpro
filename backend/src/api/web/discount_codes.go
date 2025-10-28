package webapi

import (
	"encoding/json"
	"net/http"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// getDiscountCodes returns all discount codes for a tenant, optionally filtered by affiliate (admin only)
func (api *API) getDiscountCodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	affiliateID := r.URL.Query().Get("affiliateId")
	activeOnly := r.URL.Query().Get("active") == "true"

	var affiliateIDPtr *string
	if affiliateID != "" {
		affiliateIDPtr = &affiliateID
	}

	logger.Infof("Fetching discount codes for tenant: %s (affiliateId=%v, activeOnly=%v)", tenantID, affiliateID, activeOnly)

	codes, err := api.store.GetDiscountCodes(tenantID, affiliateIDPtr, activeOnly)
	if err != nil {
		logger.Errorf("Failed to get discount codes: %v", err)
		http.Error(w, "Failed to fetch discount codes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(codes); err != nil {
		logger.Errorf("Failed to encode discount codes response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getDiscountCode returns a specific discount code by ID (admin only)
func (api *API) getDiscountCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	codeID := vars["codeId"]

	logger.Infof("Fetching discount code %s for tenant %s", codeID, tenantID)

	code, err := api.store.GetDiscountCodeByID(tenantID, codeID)
	if err != nil {
		logger.Errorf("Failed to get discount code: %v", err)
		http.Error(w, "Discount code not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(code); err != nil {
		logger.Errorf("Failed to encode discount code response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// validateDiscountCode validates a discount code by code string (admin only)
func (api *API) validateDiscountCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	codeStr := r.URL.Query().Get("code")

	if codeStr == "" {
		http.Error(w, "code query parameter required", http.StatusBadRequest)
		return
	}

	logger.Infof("Validating discount code %s for tenant %s", codeStr, tenantID)

	code, err := api.store.GetDiscountCodeByCode(tenantID, codeStr)
	if err != nil {
		logger.Errorf("Failed to validate discount code: %v", err)
		http.Error(w, "Discount code not found", http.StatusNotFound)
		return
	}

	// Check if code is valid (active, not expired, not max uses)
	if !code.IsValid() {
		logger.Warningf("Discount code %s is not valid", codeStr)
		http.Error(w, "Discount code is not valid or has expired", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(code); err != nil {
		logger.Errorf("Failed to encode discount code response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// createDiscountCode creates a new discount code for an affiliate (admin only)
func (api *API) createDiscountCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	type CreateDiscountCodeRequest struct {
		Code            string   `json:"code"`
		Description     *string  `json:"description"`
		DiscountType    string   `json:"discountType"` // PERCENTAGE or FIXED_AMOUNT
		DiscountValue   float64  `json:"discountValue"`
		MaxUses         *int     `json:"maxUses"`
		ValidFrom       *string  `json:"validFrom"`
		ValidUntil      *string  `json:"validUntil"`
		AffiliateID     string   `json:"affiliateId"`
		CommissionRate  *float64 `json:"commissionRate"`
	}

	var input CreateDiscountCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.Code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}
	if input.DiscountType != types.DiscountTypePercentage && input.DiscountType != types.DiscountTypeFixedAmount {
		http.Error(w, "discountType must be PERCENTAGE or FIXED_AMOUNT", http.StatusBadRequest)
		return
	}
	if input.DiscountValue <= 0 {
		http.Error(w, "discountValue must be greater than 0", http.StatusBadRequest)
		return
	}
	if input.AffiliateID == "" {
		http.Error(w, "affiliateId is required", http.StatusBadRequest)
		return
	}

	logger.Infof("Creating discount code %s for tenant %s, affiliate %s", input.Code, tenantID, input.AffiliateID)

	affiliateUUID, err := uuid.Parse(input.AffiliateID)
	if err != nil {
		http.Error(w, "Invalid affiliate ID", http.StatusBadRequest)
		return
	}

	discountCode := &types.DiscountCode{
		Code:            input.Code,
		Description:     input.Description,
		DiscountType:    input.DiscountType,
		DiscountValue:   input.DiscountValue,
		MaxUses:         input.MaxUses,
		CurrentUses:     0,
		ValidFrom:       input.ValidFrom,
		ValidUntil:      input.ValidUntil,
		IsActive:        true,
		IsAffiliateCode: true,
		AffiliateID:     &affiliateUUID,
		CommissionRate:  input.CommissionRate,
	}

	// Use affiliate's default commission rate if not specified
	if discountCode.CommissionRate == nil {
		affiliate, err := api.store.GetAffiliateByID(tenantID, input.AffiliateID)
		if err != nil {
			logger.Errorf("Failed to get affiliate: %v", err)
			http.Error(w, "Affiliate not found", http.StatusNotFound)
			return
		}
		discountCode.CommissionRate = &affiliate.DefaultCommissionRate
	}

	created, err := api.store.CreateDiscountCode(tenantID, discountCode)
	if err != nil {
		logger.Errorf("Failed to create discount code: %v", err)
		http.Error(w, "Failed to create discount code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(created); err != nil {
		logger.Errorf("Failed to encode discount code response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// updateDiscountCode updates an existing discount code (admin only)
func (api *API) updateDiscountCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	codeID := vars["codeId"]

	type UpdateDiscountCodeRequest struct {
		Code           string   `json:"code"`
		Description    *string  `json:"description"`
		DiscountType   string   `json:"discountType"`
		DiscountValue  float64  `json:"discountValue"`
		MaxUses        *int     `json:"maxUses"`
		ValidFrom      *string  `json:"validFrom"`
		ValidUntil     *string  `json:"validUntil"`
		IsActive       bool     `json:"isActive"`
		CommissionRate *float64 `json:"commissionRate"`
	}

	var input UpdateDiscountCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infof("Updating discount code %s for tenant %s", codeID, tenantID)

	discountCode := &types.DiscountCode{
		Code:           input.Code,
		Description:    input.Description,
		DiscountType:   input.DiscountType,
		DiscountValue:  input.DiscountValue,
		MaxUses:        input.MaxUses,
		ValidFrom:      input.ValidFrom,
		ValidUntil:     input.ValidUntil,
		IsActive:       input.IsActive,
		CommissionRate: input.CommissionRate,
	}

	updated, err := api.store.UpdateDiscountCode(tenantID, codeID, discountCode)
	if err != nil {
		logger.Errorf("Failed to update discount code: %v", err)
		http.Error(w, "Failed to update discount code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		logger.Errorf("Failed to encode discount code response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// deactivateDiscountCode deactivates a discount code (admin only)
func (api *API) deactivateDiscountCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	codeID := vars["codeId"]

	logger.Infof("Deactivating discount code %s for tenant %s", codeID, tenantID)

	if err := api.store.DeactivateDiscountCode(tenantID, codeID); err != nil {
		logger.Errorf("Failed to deactivate discount code: %v", err)
		http.Error(w, "Failed to deactivate discount code", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
