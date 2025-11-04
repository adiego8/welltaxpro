package middleware

import (
	"context"
	"net/http"
	"strings"
	"welltaxpro/src/internal/auth"

	"github.com/google/logger"
)

type contextKey string

const FirebaseUIDContextKey contextKey = "firebaseUID"

// TenantUserAuthMiddleware validates Firebase token for tenant users (clients)
// Unlike AuthMiddleware, this does not require an employee record
type TenantUserAuthMiddleware struct {
	auth *auth.Auth
}

// NewTenantUserAuthMiddleware creates a new tenant user auth middleware
func NewTenantUserAuthMiddleware(authClient *auth.Auth) *TenantUserAuthMiddleware {
	return &TenantUserAuthMiddleware{
		auth: authClient,
	}
}

// Authenticate validates the Firebase token and stores the Firebase UID in context
func (m *TenantUserAuthMiddleware) Authenticate(next http.Handler) http.Handler {
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

		// Add Firebase UID to request context
		ctx := context.WithValue(r.Context(), FirebaseUIDContextKey, *firebaseUID)
		logger.Infof("Authenticated tenant user with Firebase UID: %s", *firebaseUID)

		// Call next handler with Firebase UID in context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetFirebaseUIDFromContext retrieves the Firebase UID from context
func GetFirebaseUIDFromContext(ctx context.Context) (string, error) {
	firebaseUID, ok := ctx.Value(FirebaseUIDContextKey).(string)
	if !ok {
		return "", http.ErrNoCookie // Using standard error for "not found"
	}
	return firebaseUID, nil
}
