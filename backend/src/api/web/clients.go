package webapi

import (
	"encoding/json"
	"net/http"

	"github.com/google/logger"
	"github.com/gorilla/mux"
)

// getClients returns all clients for a tenant
func (api *API) getClients(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	if tenantID == "" {
		logger.Warning("getClients called without tenant ID")
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	logger.Infof("[getClients] Starting request - TenantID: %s, Method: %s, Path: %s", tenantID, r.Method, r.URL.Path)

	clients, err := api.store.GetClients(tenantID)
	if err != nil {
		logger.Errorf("[getClients] FAILED - TenantID: %s, Error: %v", tenantID, err)
		http.Error(w, "failed to fetch clients", http.StatusInternalServerError)
		return
	}

	logger.Infof("[getClients] SUCCESS - TenantID: %s, ClientCount: %d", tenantID, len(clients))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clients); err != nil {
		logger.Errorf("[getClients] Failed to encode response - TenantID: %s, Error: %v", tenantID, err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getClient returns a specific client by ID for a tenant
func (api *API) getClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	clientID := vars["clientId"]

	if tenantID == "" || clientID == "" {
		http.Error(w, "tenant ID and client ID are required", http.StatusBadRequest)
		return
	}

	logger.Infof("Fetching client %s for tenant: %s", clientID, tenantID)

	client, err := api.store.GetClientByID(tenantID, clientID)
	if err != nil {
		logger.Errorf("Failed to get client %s for tenant %s: %v", clientID, tenantID, err)
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(client); err != nil {
		logger.Errorf("Failed to encode client response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getClientComprehensive returns all data for a specific client (filings, dependents, etc.)
func (api *API) getClientComprehensive(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	clientID := vars["clientId"]

	if tenantID == "" || clientID == "" {
		http.Error(w, "tenant ID and client ID are required", http.StatusBadRequest)
		return
	}

	logger.Infof("Fetching comprehensive data for client %s (tenant: %s)", clientID, tenantID)

	clientData, err := api.store.GetClientComprehensive(tenantID, clientID)
	if err != nil {
		logger.Errorf("Failed to get comprehensive data for client %s (tenant %s): %v", clientID, tenantID, err)
		http.Error(w, "failed to fetch client data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clientData); err != nil {
		logger.Errorf("Failed to encode comprehensive client response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getFilings returns clients with their filings (paginated, no filtering)
func (api *API) getFilings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]

	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	// Get pagination parameters (default: limit=100, offset=0)
	limit := 100
	offset := 0

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := json.Number(limitParam).Int64(); err == nil && parsedLimit > 0 {
			limit = int(parsedLimit)
		}
	}

	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if parsedOffset, err := json.Number(offsetParam).Int64(); err == nil && parsedOffset >= 0 {
			offset = int(parsedOffset)
		}
	}

	logger.Infof("Fetching filings for tenant %s with pagination - limit: %d, offset: %d", tenantID, limit, offset)

	clientsData, err := api.store.GetClientsByFilings(tenantID, limit, offset)
	if err != nil {
		logger.Errorf("Failed to get filings for tenant %s: %v", tenantID, err)
		http.Error(w, "failed to fetch filings", http.StatusInternalServerError)
		return
	}

	logger.Infof("Successfully fetched %d clients with their filings", len(clientsData))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clientsData); err != nil {
		logger.Errorf("Failed to encode filings response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
