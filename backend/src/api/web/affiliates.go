package webapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// getAffiliates returns all affiliates for a tenant (admin only)
func (api *API) getAffiliates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	activeOnly := r.URL.Query().Get("active") == "true"

	logger.Infof("Fetching affiliates for tenant: %s", tenantID)

	affiliates, err := api.store.GetAffiliates(tenantID, activeOnly)
	if err != nil {
		logger.Errorf("Failed to get affiliates: %v", err)
		http.Error(w, "Failed to fetch affiliates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(affiliates); err != nil {
		logger.Errorf("Failed to encode affiliates response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getAffiliate returns a specific affiliate by ID (admin only)
func (api *API) getAffiliate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	affiliateID := vars["affiliateId"]

	logger.Infof("Fetching affiliate %s for tenant %s", affiliateID, tenantID)

	affiliate, err := api.store.GetAffiliateByID(tenantID, affiliateID)
	if err != nil {
		logger.Errorf("Failed to get affiliate: %v", err)
		http.Error(w, "Affiliate not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(affiliate); err != nil {
		logger.Errorf("Failed to encode affiliate response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// createAffiliate creates a new affiliate (admin only)
func (api *API) createAffiliate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	var input types.Affiliate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infof("Creating affiliate for tenant %s: %s %s", tenantID, input.FirstName, input.LastName)

	// Set defaults if not provided
	if input.PayoutMethod == "" {
		input.PayoutMethod = types.PayoutMethodManual
	}
	if input.PayoutThreshold == 0 {
		input.PayoutThreshold = 100.00
	}
	if input.DefaultCommissionRate == 0 {
		input.DefaultCommissionRate = 15.00
	}
	input.IsActive = true

	affiliate, err := api.store.CreateAffiliate(tenantID, &input)
	if err != nil {
		logger.Errorf("Failed to create affiliate: %v", err)
		http.Error(w, "Failed to create affiliate", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(affiliate); err != nil {
		logger.Errorf("Failed to encode affiliate response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// updateAffiliate updates an existing affiliate (admin only)
func (api *API) updateAffiliate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	affiliateID := vars["affiliateId"]

	var input types.Affiliate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infof("Updating affiliate %s for tenant %s", affiliateID, tenantID)

	affiliate, err := api.store.UpdateAffiliate(tenantID, affiliateID, &input)
	if err != nil {
		logger.Errorf("Failed to update affiliate: %v", err)
		http.Error(w, "Failed to update affiliate", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(affiliate); err != nil {
		logger.Errorf("Failed to encode affiliate response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// generateAffiliateToken generates a new access token for an affiliate (admin only)
func (api *API) generateAffiliateToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	affiliateID := vars["affiliateId"]

	type TokenRequest struct {
		ExpiresAt *time.Time `json:"expiresAt"`
		Notes     *string    `json:"notes"`
	}

	var input TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infof("Generating token for affiliate %s in tenant %s", affiliateID, tenantID)

	affiliateUUID, err := uuid.Parse(affiliateID)
	if err != nil {
		http.Error(w, "Invalid affiliate ID", http.StatusBadRequest)
		return
	}

	plainToken, token, err := api.store.GenerateAffiliateToken(tenantID, affiliateUUID, input.ExpiresAt, input.Notes)
	if err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return both the token object and the plain token (only time we send it)
	response := map[string]interface{}{
		"token":     plainToken, // This is the only time the plain token is sent
		"tokenInfo": token,
		"accessUrl": fmt.Sprintf("/affiliates/%s/%s/dashboard?token=%s", tenantID, affiliateID, plainToken),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode token response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getAffiliateTokens returns all tokens for an affiliate (admin only)
func (api *API) getAffiliateTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	affiliateID := vars["affiliateId"]

	activeOnly := r.URL.Query().Get("active") == "true"

	logger.Infof("Fetching tokens for affiliate %s in tenant %s", affiliateID, tenantID)

	affiliateUUID, err := uuid.Parse(affiliateID)
	if err != nil {
		http.Error(w, "Invalid affiliate ID", http.StatusBadRequest)
		return
	}

	tokens, err := api.store.GetAffiliateTokens(tenantID, affiliateUUID, activeOnly)
	if err != nil {
		logger.Errorf("Failed to get tokens: %v", err)
		http.Error(w, "Failed to fetch tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		logger.Errorf("Failed to encode tokens response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// revokeAffiliateToken revokes a specific token (admin only)
func (api *API) revokeAffiliateToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	tokenID := vars["tokenId"]

	logger.Infof("Revoking token %s in tenant %s", tokenID, tenantID)

	tokenUUID, err := uuid.Parse(tokenID)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	if err := api.store.RevokeAffiliateToken(tenantID, tokenUUID); err != nil {
		logger.Errorf("Failed to revoke token: %v", err)
		http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getCommissions returns commissions with optional filters (admin only)
func (api *API) getCommissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	affiliateID := r.URL.Query().Get("affiliateId")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	// Make affiliateID optional - if not provided, fetch all commissions
	var affiliateIDPtr *string
	if affiliateID != "" {
		affiliateIDPtr = &affiliateID
		logger.Infof("Fetching commissions for affiliate %s in tenant %s", affiliateID, tenantID)
	} else {
		logger.Infof("Fetching all commissions in tenant %s", tenantID)
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	commissions, err := api.store.GetCommissionsByAffiliate(tenantID, affiliateIDPtr, statusPtr, limit)
	if err != nil {
		logger.Errorf("Failed to get commissions: %v", err)
		http.Error(w, "Failed to fetch commissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commissions); err != nil {
		logger.Errorf("Failed to encode commissions response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// approveCommission approves a pending commission (admin only)
func (api *API) approveCommission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	commissionID := vars["commissionId"]

	logger.Infof("Approving commission %s in tenant %s", commissionID, tenantID)

	commission, err := api.store.ApproveCommission(tenantID, commissionID)
	if err != nil {
		logger.Errorf("Failed to approve commission: %v", err)
		http.Error(w, "Failed to approve commission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commission); err != nil {
		logger.Errorf("Failed to encode commission response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// markCommissionPaid marks an approved commission as paid (admin only)
func (api *API) markCommissionPaid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	commissionID := vars["commissionId"]

	logger.Infof("Marking commission %s as paid in tenant %s", commissionID, tenantID)

	commission, err := api.store.MarkCommissionPaid(tenantID, commissionID)
	if err != nil {
		logger.Errorf("Failed to mark commission as paid: %v", err)
		http.Error(w, "Failed to mark commission as paid", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commission); err != nil {
		logger.Errorf("Failed to encode commission response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// cancelCommission cancels a commission with a reason (admin only)
func (api *API) cancelCommission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	commissionID := vars["commissionId"]

	// Parse request body for cancellation reason
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode cancel request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		http.Error(w, "Cancellation reason is required", http.StatusBadRequest)
		return
	}

	logger.Infof("Cancelling commission %s in tenant %s with reason: %s", commissionID, tenantID, req.Reason)

	commission, err := api.store.CancelCommission(tenantID, commissionID, req.Reason)
	if err != nil {
		logger.Errorf("Failed to cancel commission: %v", err)
		http.Error(w, "Failed to cancel commission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commission); err != nil {
		logger.Errorf("Failed to encode commission response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
