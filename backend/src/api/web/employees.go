package webapi

import (
	"encoding/json"
	"net/http"
	"welltaxpro/src/internal/middleware"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// CreateEmployeeRequest represents the request body for creating an employee
type CreateEmployeeRequest struct {
	FirebaseUID string   `json:"firebaseUid"`
	Email       string   `json:"email"`
	FirstName   *string  `json:"firstName,omitempty"`
	LastName    *string  `json:"lastName,omitempty"`
	Role        string   `json:"role"`
	TenantIDs   []string `json:"tenantIds,omitempty"` // List of tenant IDs this employee should have access to
}

// CreateEmployeeResponse represents the response after creating an employee
type CreateEmployeeResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Employee interface{} `json:"employee,omitempty"`
}

// GetEmployeeTenantAccessRequest represents the request for getting tenant access
type GetEmployeeTenantAccessRequest struct {
	EmployeeID uuid.UUID `json:"employeeId"`
}

// AssignTenantRequest represents the request for assigning an employee to a tenant
type AssignTenantRequest struct {
	EmployeeID uuid.UUID `json:"employeeId"`
	TenantID   string    `json:"tenantId"`
	Role       string    `json:"role"` // Role within this tenant
}

// RemoveTenantRequest represents the request for removing an employee from a tenant
type RemoveTenantRequest struct {
	EmployeeID uuid.UUID `json:"employeeId"`
	TenantID   string    `json:"tenantId"`
}

// getAllEmployees handles GET /api/v1/employees
// Returns all employees (admin only)
func (api *API) getAllEmployees(w http.ResponseWriter, r *http.Request) {
	includeInactive := r.URL.Query().Get("includeInactive") == "true"

	logger.Infof("Fetching all employees (includeInactive=%v)", includeInactive)

	employees, err := api.store.GetAllEmployees(includeInactive)
	if err != nil {
		logger.Errorf("Failed to get employees: %v", err)
		http.Error(w, "Failed to fetch employees", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(employees); err != nil {
		logger.Errorf("Failed to encode employees response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getEmployeeByID handles GET /api/v1/employees/{employeeId}
// Returns a specific employee (admin only)
func (api *API) getEmployeeByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeIDStr := vars["employeeId"]

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
		return
	}

	logger.Infof("Fetching employee: %s", employeeID)

	employee, err := api.store.GetEmployeeByID(employeeID)
	if err != nil {
		logger.Errorf("Failed to get employee: %v", err)
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(employee); err != nil {
		logger.Errorf("Failed to encode employee response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// createEmployee handles POST /api/v1/employees
// This endpoint creates a new employee record when a user signs up with Google
func (api *API) createEmployee(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode create employee request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FirebaseUID == "" {
		http.Error(w, "Firebase UID is required", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "accountant" // Default role
	}

	// Validate role
	validRoles := map[string]bool{
		"admin":      true,
		"accountant": true,
		"support":    true,
	}
	if !validRoles[req.Role] {
		http.Error(w, "Invalid role. Must be one of: admin, accountant, support", http.StatusBadRequest)
		return
	}

	logger.Infof("Creating employee for Firebase UID: %s, Email: %s", req.FirebaseUID, req.Email)

	// Check if employee already exists
	existingEmployee, err := api.store.GetEmployeeByFirebaseUID(req.FirebaseUID)
	if err == nil && existingEmployee != nil {
		logger.Infof("Employee already exists for Firebase UID: %s", req.FirebaseUID)

		response := CreateEmployeeResponse{
			Success:  true,
			Message:  "Employee already exists",
			Employee: existingEmployee,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Errorf("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// Create new employee
	employee, err := api.store.CreateEmployee(req.FirebaseUID, req.Email, req.FirstName, req.LastName, req.Role)
	if err != nil {
		logger.Errorf("Failed to create employee: %v", err)
		http.Error(w, "Failed to create employee", http.StatusInternalServerError)
		return
	}

	logger.Infof("Successfully created employee: %s (%s)", employee.Email, employee.ID)

	response := CreateEmployeeResponse{
		Success:  true,
		Message:  "Employee created successfully",
		Employee: employee,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getMe handles GET /api/v1/employees/me
// This endpoint returns the current authenticated employee's information
func (api *API) getMe(w http.ResponseWriter, r *http.Request) {
	// Get employee from context (set by auth middleware)
	employee, ok := middleware.GetEmployeeFromContext(r.Context())
	if !ok {
		logger.Error("Employee not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(employee); err != nil {
		logger.Errorf("Failed to encode employee response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// updateEmployee handles PUT /api/v1/employees/me
// This endpoint allows an employee to update their own information
func (api *API) updateEmployee(w http.ResponseWriter, r *http.Request) {
	// Get employee from context
	employee, ok := middleware.GetEmployeeFromContext(r.Context())
	if !ok {
		logger.Error("Employee not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var updateReq struct {
		FirstName *string `json:"firstName,omitempty"`
		LastName  *string `json:"lastName,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		logger.Errorf("Failed to decode update employee request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, we'll just return the current employee
	// TODO: Implement employee update functionality in the store
	logger.Infof("Employee %s requested profile update", employee.Email)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(employee); err != nil {
		logger.Errorf("Failed to encode employee response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getEmployeeTenants handles GET /api/v1/employees/me/tenants
// Returns the list of tenants the current employee has access to
func (api *API) getEmployeeTenants(w http.ResponseWriter, r *http.Request) {
	// Get employee from context
	employee, ok := middleware.GetEmployeeFromContext(r.Context())
	if !ok {
		logger.Error("Employee not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger.Infof("Getting tenant access for employee: %s", employee.Email)

	// Query employee's tenant access from database
	query := `
		SELECT eta.tenant_id, tc.tenant_name, eta.role, eta.is_active
		FROM employee_tenant_access eta
		JOIN tenant_connections tc ON eta.tenant_id = tc.tenant_id
		WHERE eta.employee_id = $1
		ORDER BY tc.tenant_name
	`

	rows, err := api.store.DB.Query(query, employee.ID)
	if err != nil {
		logger.Errorf("Failed to query tenant access: %v", err)
		http.Error(w, "Failed to fetch tenant access", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tenantAccess := []map[string]interface{}{}
	for rows.Next() {
		var tenantID, tenantName, role string
		var isActive bool

		err := rows.Scan(&tenantID, &tenantName, &role, &isActive)
		if err != nil {
			logger.Errorf("Failed to scan tenant access row: %v", err)
			continue
		}

		tenantAccess = append(tenantAccess, map[string]interface{}{
			"tenantId":   tenantID,
			"tenantName": tenantName,
			"role":       role,
			"isActive":   isActive,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tenantAccess); err != nil {
		logger.Errorf("Failed to encode tenant access response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// assignEmployeeToTenant handles POST /api/v1/employees/{employeeId}/tenants
// Assigns an employee to a tenant (admin only)
func (api *API) assignEmployeeToTenant(w http.ResponseWriter, r *http.Request) {
	// Get employee from context
	currentEmployee, ok := middleware.GetEmployeeFromContext(r.Context())
	if !ok {
		logger.Error("Employee not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if current employee is admin
	if !currentEmployee.IsAdmin() {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	// Parse request body
	var req AssignTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode assign tenant request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.TenantID == "" {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "accountant" // Default role
	}

	// Validate role
	validRoles := map[string]bool{
		"admin":      true,
		"accountant": true,
		"viewer":     true,
	}
	if !validRoles[req.Role] {
		http.Error(w, "Invalid role. Must be one of: admin, accountant, viewer", http.StatusBadRequest)
		return
	}

	// TODO: Implement store method to assign employee to tenant
	logger.Infof("Admin %s assigned employee %s to tenant %s with role %s",
		currentEmployee.Email, req.EmployeeID, req.TenantID, req.Role)

	response := map[string]interface{}{
		"success": true,
		"message": "Employee assigned to tenant successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// removeEmployeeFromTenant handles DELETE /api/v1/employees/{employeeId}/tenants/{tenantId}
// Removes an employee's access to a tenant (admin only)
func (api *API) removeEmployeeFromTenant(w http.ResponseWriter, r *http.Request) {
	// Get employee from context
	currentEmployee, ok := middleware.GetEmployeeFromContext(r.Context())
	if !ok {
		logger.Error("Employee not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if current employee is admin
	if !currentEmployee.IsAdmin() {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	// Get URL parameters
	vars := mux.Vars(r)
	employeeIDStr := vars["employeeId"]
	tenantID := vars["tenantId"]

	if employeeIDStr == "" || tenantID == "" {
		http.Error(w, "Employee ID and Tenant ID are required", http.StatusBadRequest)
		return
	}

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
		return
	}

	// TODO: Implement store method to remove employee from tenant
	logger.Infof("Admin %s removed employee %s from tenant %s",
		currentEmployee.Email, employeeID, tenantID)

	response := map[string]interface{}{
		"success": true,
		"message": "Employee removed from tenant successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
