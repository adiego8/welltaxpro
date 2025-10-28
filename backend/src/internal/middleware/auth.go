package middleware

import (
	"context"
	"net/http"
	"strings"
	"welltaxpro/src/internal/auth"
	"welltaxpro/src/internal/store"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

// AuthMiddleware validates Firebase token and loads employee context
type AuthMiddleware struct {
	auth  *auth.Auth
	store *store.Store
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authClient *auth.Auth, store *store.Store) *AuthMiddleware {
	return &AuthMiddleware{
		auth:  authClient,
		store: store,
	}
}

// Authenticate validates the token and loads employee into request context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warning("Missing Authorization header")
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix if present
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token with Firebase
		firebaseUID, err := m.auth.ValidateToken(r.Context(), token)
		if err != nil {
			logger.Errorf("Token validation failed: %v", err)
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Load employee from database
		employee, err := m.store.GetEmployeeByFirebaseUID(*firebaseUID)
		if err != nil {
			logger.Errorf("Failed to load employee for firebase UID %s: %v", *firebaseUID, err)
			http.Error(w, "Unauthorized: Employee not found", http.StatusUnauthorized)
			return
		}

		// Check if employee is active
		if !employee.IsActive {
			logger.Warningf("Inactive employee attempted access: %s", employee.Email)
			http.Error(w, "Unauthorized: Account inactive", http.StatusUnauthorized)
			return
		}

		// Add employee to request context
		ctx := context.WithValue(r.Context(), auth.EmployeeContextKey, employee)
		logger.Infof("Authenticated employee: %s (%s)", employee.Email, employee.Role)

		// Call next handler with employee context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetEmployeeFromContext retrieves the authenticated employee from context
func GetEmployeeFromContext(ctx context.Context) (*types.Employee, bool) {
	employee, ok := ctx.Value(auth.EmployeeContextKey).(*types.Employee)
	return employee, ok
}

// RequireRole is a middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			employee, ok := GetEmployeeFromContext(r.Context())
			if !ok {
				logger.Error("Employee not found in context")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if employee.Role != role && employee.Role != "admin" {
				logger.Warningf("Employee %s lacks required role %s", employee.Email, role)
				http.Error(w, "Forbidden: Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireRole("admin")(next)
}
