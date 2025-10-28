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
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	logger.Infof("Fetching clients for tenant: %s", tenantID)

	clients, err := api.store.GetClients(tenantID)
	if err != nil {
		logger.Errorf("Failed to get clients for tenant %s: %v", tenantID, err)
		http.Error(w, "failed to fetch clients", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clients); err != nil {
		logger.Errorf("Failed to encode clients response: %v", err)
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
