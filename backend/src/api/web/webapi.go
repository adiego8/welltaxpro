package webapi

import (
	"context"
	"net/http"
	"welltaxpro/src/internal/auth"
	"welltaxpro/src/internal/middleware"
	"welltaxpro/src/internal/notification"
	"welltaxpro/src/internal/store"
	"welltaxpro/src/internal/types"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type API struct {
	context              context.Context
	Router               *mux.Router
	store                *store.Store
	authMiddleware       *middleware.AuthMiddleware
	tenantUserAuthMiddleware *middleware.TenantUserAuthMiddleware
	auditMiddleware      *middleware.AuditMiddleware
	emailService         *notification.EmailService
}

// NewAPI creates and returns a new API instance
func NewAPI(ctx context.Context, s *store.Store, authClient *auth.Auth, emailService *notification.EmailService) *API {
	authMw := middleware.NewAuthMiddleware(authClient, s)
	tenantUserAuthMw := middleware.NewTenantUserAuthMiddleware(authClient)
	auditMw := middleware.NewAuditMiddleware(s)

	return &API{
		context:              ctx,
		Router:               mux.NewRouter(),
		store:                s,
		authMiddleware:       authMw,
		tenantUserAuthMiddleware: tenantUserAuthMw,
		auditMiddleware:      auditMw,
		emailService:         emailService,
	}
}

// CORSHandler wraps the router with CORS middleware
func (api *API) CORSHandler(corsConfig CORSConfig) http.Handler {
	// Set secure defaults if not configured
	allowedOrigins := corsConfig.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}

	allowedMethods := corsConfig.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	allowedHeaders := corsConfig.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{"Content-Type", "Authorization"}
	}

	corsOptions := []handlers.CORSOption{
		handlers.AllowedOrigins(allowedOrigins),
		handlers.AllowedMethods(allowedMethods),
		handlers.AllowedHeaders(allowedHeaders),
	}

	if corsConfig.AllowCredentials {
		corsOptions = append(corsOptions, handlers.AllowCredentials())
	}

	corsHandler := handlers.CORS(corsOptions...)

	// Wrap with security headers middleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Apply CORS handler
		corsHandler(api.Router).ServeHTTP(w, r)
	})
}

// InitRoutes initializes the routes and handlers
func (api *API) InitRoutes() {
	// Health check (no auth required)
	api.Router.HandleFunc("/health", api.healthCheck).Methods(http.MethodGet)

	// Tenant management endpoints (admin only)
	api.Router.Handle("/api/v1/admin/tenants",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getAllTenants),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/admin/tenants",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.createTenant),
			),
		),
	).Methods(http.MethodPost)

	api.Router.Handle("/api/v1/admin/tenants/{tenantId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getTenant),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/admin/tenants/{tenantId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.updateTenant),
			),
		),
	).Methods(http.MethodPut)

	api.Router.Handle("/api/v1/admin/tenants/{tenantId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.deleteTenant),
			),
		),
	).Methods(http.MethodDelete)

	// Employee management endpoints
	// Create employee (public endpoint for user signup)
	api.Router.HandleFunc("/api/v1/employees", api.createEmployee).Methods(http.MethodPost)

	// Get all employees (admin only)
	api.Router.Handle("/api/v1/employees",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getAllEmployees),
			),
		),
	).Methods(http.MethodGet)

	// Get current employee info (requires auth)
	api.Router.Handle("/api/v1/employees/me",
		api.authMiddleware.Authenticate(
			http.HandlerFunc(api.getMe),
		),
	).Methods(http.MethodGet)

	// Update current employee info (requires auth)
	api.Router.Handle("/api/v1/employees/me",
		api.authMiddleware.Authenticate(
			http.HandlerFunc(api.updateEmployee),
		),
	).Methods(http.MethodPut)

	// Get current employee's tenant access (requires auth)
	api.Router.Handle("/api/v1/employees/me/tenants",
		api.authMiddleware.Authenticate(
			http.HandlerFunc(api.getEmployeeTenants),
		),
	).Methods(http.MethodGet)

	// Get employee by ID (admin only)
	api.Router.Handle("/api/v1/employees/{employeeId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getEmployeeByID),
			),
		),
	).Methods(http.MethodGet)

	// Assign employee to tenant (admin only)
	api.Router.Handle("/api/v1/employees/{employeeId}/tenants",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.assignEmployeeToTenant),
			),
		),
	).Methods(http.MethodPost)

	// Remove employee from tenant (admin only)
	api.Router.Handle("/api/v1/employees/{employeeId}/tenants/{tenantId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.removeEmployeeFromTenant),
			),
		),
	).Methods(http.MethodDelete)

	// Admin API for tenant clients (auth + audit required)
	api.Router.Handle("/api/v1/{tenantId}/clients",
		api.authMiddleware.Authenticate(
			api.auditMiddleware.LogAccess(types.AuditActionView, types.AuditResourceClient)(
				http.HandlerFunc(api.getClients),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/clients/{clientId}",
		api.authMiddleware.Authenticate(
			api.auditMiddleware.LogAccess(types.AuditActionView, types.AuditResourceClient)(
				http.HandlerFunc(api.getClient),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/clients/{clientId}/comprehensive",
		api.authMiddleware.Authenticate(
			api.auditMiddleware.LogAccess(types.AuditActionView, types.AuditResourceClient)(
				http.HandlerFunc(api.getClientComprehensive),
			),
		),
	).Methods(http.MethodGet)

	// Filings endpoint (filtered by status/year)
	api.Router.Handle("/api/v1/{tenantId}/filings",
		api.authMiddleware.Authenticate(
			api.auditMiddleware.LogAccess(types.AuditActionView, types.AuditResourceClient)(
				http.HandlerFunc(api.getFilings),
			),
		),
	).Methods(http.MethodGet)

	// Admin affiliate management (auth + admin required)
	api.Router.Handle("/api/v1/{tenantId}/affiliates",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getAffiliates),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/affiliates",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.createAffiliate),
			),
		),
	).Methods(http.MethodPost)

	api.Router.Handle("/api/v1/{tenantId}/affiliates/{affiliateId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getAffiliate),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/affiliates/{affiliateId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.updateAffiliate),
			),
		),
	).Methods(http.MethodPut)

	api.Router.Handle("/api/v1/{tenantId}/affiliates/{affiliateId}/generate-token",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.generateAffiliateToken),
			),
		),
	).Methods(http.MethodPost)

	api.Router.Handle("/api/v1/{tenantId}/affiliates/{affiliateId}/tokens",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getAffiliateTokens),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/affiliates/{affiliateId}/tokens/{tokenId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.revokeAffiliateToken),
			),
		),
	).Methods(http.MethodDelete)

	api.Router.Handle("/api/v1/{tenantId}/commissions",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getCommissions),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/commissions/{commissionId}/approve",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.approveCommission),
			),
		),
	).Methods(http.MethodPut)

	api.Router.Handle("/api/v1/{tenantId}/commissions/{commissionId}/mark-paid",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.markCommissionPaid),
			),
		),
	).Methods(http.MethodPut)

	api.Router.Handle("/api/v1/{tenantId}/commissions/{commissionId}/cancel",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.cancelCommission),
			),
		),
	).Methods(http.MethodPut)

	// Discount code management (admin only)
	api.Router.Handle("/api/v1/{tenantId}/discount-codes",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getDiscountCodes),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/discount-codes",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.createDiscountCode),
			),
		),
	).Methods(http.MethodPost)

	api.Router.Handle("/api/v1/{tenantId}/discount-codes/validate",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.validateDiscountCode),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/discount-codes/{codeId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.getDiscountCode),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/discount-codes/{codeId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.updateDiscountCode),
			),
		),
	).Methods(http.MethodPut)

	api.Router.Handle("/api/v1/{tenantId}/discount-codes/{codeId}/deactivate",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.deactivateDiscountCode),
			),
		),
	).Methods(http.MethodPut)

	// Document management endpoints (admin only with audit)
	api.Router.Handle("/api/v1/{tenantId}/filings/{filingId}/documents",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				api.auditMiddleware.LogAccess(types.AuditActionUpload, types.AuditResourceDocument)(
					http.HandlerFunc(api.uploadDocument),
				),
			),
		),
	).Methods(http.MethodPost)

	api.Router.Handle("/api/v1/{tenantId}/filings/{filingId}/documents",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				api.auditMiddleware.LogAccess(types.AuditActionView, types.AuditResourceDocument)(
					http.HandlerFunc(api.getDocuments),
				),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/documents/{documentId}/download",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				api.auditMiddleware.LogAccess(types.AuditActionDownload, types.AuditResourceDocument)(
					http.HandlerFunc(api.downloadDocument),
				),
			),
		),
	).Methods(http.MethodGet)

	api.Router.Handle("/api/v1/{tenantId}/documents/{documentId}",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				api.auditMiddleware.LogAccess(types.AuditActionDelete, types.AuditResourceDocument)(
					http.HandlerFunc(api.deleteDocument),
				),
			),
		),
	).Methods(http.MethodDelete)

	// Signature endpoints (admin only)
	api.Router.Handle("/api/v1/{tenantId}/signature/send",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.sendSignatureRequest),
			),
		),
	).Methods(http.MethodPost)

	// Filing management endpoints (admin only)
	api.Router.Handle("/api/v1/{tenantId}/filings/{filingId}/complete",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.markFilingCompleted),
			),
		),
	).Methods(http.MethodPut)

	// Tenant User Portal endpoints (Firebase-authenticated client access)
	// Auto-register tenant user on first sign-in (requires Firebase auth)
	api.Router.Handle("/api/v1/{tenantId}/user/register",
		api.tenantUserAuthMiddleware.Authenticate(
			http.HandlerFunc(api.autoRegisterTenantUser),
		),
	).Methods(http.MethodPost)

	// Manual registration by admin (admin only) - links Firebase UID to client record
	api.Router.Handle("/api/v1/{tenantId}/users/register",
		api.authMiddleware.Authenticate(
			api.authMiddleware.RequireAdmin(
				http.HandlerFunc(api.registerTenantUser),
			),
		),
	).Methods(http.MethodPost)

	// Get tenant user's own profile and data (requires Firebase auth, tenant user only)
	api.Router.Handle("/api/v1/{tenantId}/user/profile",
		api.tenantUserAuthMiddleware.Authenticate(
			http.HandlerFunc(api.getTenantUserProfile),
		),
	).Methods(http.MethodGet)

	// Download tenant user's own document (requires Firebase auth, tenant user only)
	api.Router.Handle("/api/v1/{tenantId}/user/documents/{documentId}/download",
		api.tenantUserAuthMiddleware.Authenticate(
			http.HandlerFunc(api.downloadTenantUserDocument),
		),
	).Methods(http.MethodGet)

	// Public affiliate endpoints (token-based, no Firebase auth)
	api.Router.HandleFunc("/api/v1/{tenantId}/affiliates/{affiliateId}/dashboard", api.getAffiliateDashboard).Methods(http.MethodGet)
	api.Router.HandleFunc("/api/v1/{tenantId}/affiliates/{affiliateId}/stats", api.getAffiliateStatsPublic).Methods(http.MethodGet)
	api.Router.HandleFunc("/api/v1/{tenantId}/affiliates/{affiliateId}/commissions", api.getAffiliateCommissionsPublic).Methods(http.MethodGet)
}

// healthCheck returns 200 OK if service is running
func (api *API) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
