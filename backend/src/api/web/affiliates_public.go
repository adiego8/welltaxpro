package webapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// validateAffiliateToken validates the token and verifies it matches the affiliate ID
func (api *API) validateAffiliateToken(tenantID, affiliateID, token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	// Validate token and get affiliate ID
	tokenAffiliateID, err := api.store.ValidateAffiliateToken(tenantID, token)
	if err != nil {
		return false, err
	}

	// Verify token's affiliate ID matches URL parameter
	expectedAffiliateID, err := uuid.Parse(affiliateID)
	if err != nil {
		return false, err
	}

	return tokenAffiliateID == expectedAffiliateID, nil
}

// getAffiliateDashboard returns complete dashboard data for an affiliate (token-based, public)
func (api *API) getAffiliateDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	affiliateID := vars["affiliateId"]
	token := r.URL.Query().Get("token")

	logger.Infof("Fetching affiliate dashboard for %s in tenant %s", affiliateID, tenantID)

	// Validate token
	valid, err := api.validateAffiliateToken(tenantID, affiliateID, token)
	if err != nil {
		logger.Errorf("Failed to validate token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	if !valid {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Get affiliate info
	affiliate, err := api.store.GetAffiliateByID(tenantID, affiliateID)
	if err != nil {
		logger.Errorf("Failed to get affiliate: %v", err)
		http.Error(w, "Affiliate not found", http.StatusNotFound)
		return
	}

	// Get affiliate stats
	stats, err := api.store.GetAffiliateStats(tenantID, affiliateID)
	if err != nil {
		logger.Errorf("Failed to get affiliate stats: %v", err)
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	// Get recent commissions (last 20)
	commissions, err := api.store.GetCommissionsByAffiliate(tenantID, &affiliateID, nil, 20)
	if err != nil {
		logger.Errorf("Failed to get commissions: %v", err)
		http.Error(w, "Failed to fetch commissions", http.StatusInternalServerError)
		return
	}

	// Build dashboard response
	dashboard := map[string]interface{}{
		"affiliate":   affiliate,
		"stats":       stats,
		"commissions": commissions,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		logger.Errorf("Failed to encode dashboard response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getAffiliateStatsPublic returns statistics for an affiliate (token-based, public)
func (api *API) getAffiliateStatsPublic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	affiliateID := vars["affiliateId"]
	token := r.URL.Query().Get("token")

	logger.Infof("Fetching affiliate stats for %s in tenant %s", affiliateID, tenantID)

	// Validate token
	valid, err := api.validateAffiliateToken(tenantID, affiliateID, token)
	if err != nil {
		logger.Errorf("Failed to validate token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	if !valid {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Get affiliate stats
	stats, err := api.store.GetAffiliateStats(tenantID, affiliateID)
	if err != nil {
		logger.Errorf("Failed to get affiliate stats: %v", err)
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		logger.Errorf("Failed to encode stats response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getAffiliateCommissionsPublic returns commissions for an affiliate (token-based, public)
func (api *API) getAffiliateCommissionsPublic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	affiliateID := vars["affiliateId"]
	token := r.URL.Query().Get("token")

	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	logger.Infof("Fetching affiliate commissions for %s in tenant %s", affiliateID, tenantID)

	// Validate token
	valid, err := api.validateAffiliateToken(tenantID, affiliateID, token)
	if err != nil {
		logger.Errorf("Failed to validate token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	if !valid {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	// Get commissions
	commissions, err := api.store.GetCommissionsByAffiliate(tenantID, &affiliateID, statusPtr, limit)
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
