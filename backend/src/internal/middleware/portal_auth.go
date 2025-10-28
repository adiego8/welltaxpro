package middleware

import (
	"context"
	"net/http"
	"welltaxpro/src/internal/portal"

	"github.com/google/logger"
)

type contextKey string

const PortalClaimsContextKey contextKey = "portalClaims"

// PortalAuthMiddleware validates portal JWT tokens
type PortalAuthMiddleware struct {
	jwtSecret string
}

// NewPortalAuthMiddleware creates a new portal auth middleware
func NewPortalAuthMiddleware(jwtSecret string) *PortalAuthMiddleware {
	return &PortalAuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// Authenticate validates the portal token and adds claims to context
func (m *PortalAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from query parameter or Authorization header
		token := r.URL.Query().Get("token")
		if token == "" {
			// Try Authorization header as fallback
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}
		}

		if token == "" {
			logger.Warning("Missing portal token")
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := portal.ValidatePortalToken(token, m.jwtSecret)
		if err != nil {
			logger.Errorf("Invalid portal token: %v", err)
			http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Verify token type is session (not magic_link)
		if claims.TokenType != "session" {
			logger.Warningf("Attempted to use %s token for portal access (client: %s)", claims.TokenType, claims.ClientID)
			http.Error(w, "Unauthorized: Invalid token type. Please complete verification first.", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), PortalClaimsContextKey, claims)
		logger.Infof("Authenticated portal access for client %s in tenant %s", claims.ClientID, claims.TenantID)

		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetPortalClaimsFromContext retrieves portal claims from context
func GetPortalClaimsFromContext(ctx context.Context) (*portal.PortalClaims, bool) {
	claims, ok := ctx.Value(PortalClaimsContextKey).(*portal.PortalClaims)
	return claims, ok
}
