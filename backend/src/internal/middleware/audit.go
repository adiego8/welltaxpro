package middleware

import (
	"net"
	"net/http"
	"strings"
	"welltaxpro/src/internal/store"

	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AuditMiddleware logs access for compliance
type AuditMiddleware struct {
	store *store.Store
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(store *store.Store) *AuditMiddleware {
	return &AuditMiddleware{
		store: store,
	}
}

// LogAccess logs the API access to audit trail
func (m *AuditMiddleware) LogAccess(action, resourceType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get employee from context
			employee, ok := GetEmployeeFromContext(r.Context())
			if !ok {
				// If no employee context, skip audit logging (unauthenticated request)
				next.ServeHTTP(w, r)
				return
			}

			// Get route variables
			vars := mux.Vars(r)
			tenantID := vars["tenantId"]
			clientID := vars["clientId"]

			// Parse client ID if present
			var clientUUID *uuid.UUID
			if clientID != "" {
				parsed, err := uuid.Parse(clientID)
				if err == nil {
					clientUUID = &parsed
				}
			}

			// Get IP address
			ipAddress := getIPAddress(r)

			// Get user agent
			userAgent := r.UserAgent()

			// Build details
			details := map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"query":  r.URL.RawQuery,
			}

			// Log the audit entry
			err := m.store.CreateAuditLog(
				employee.ID,
				tenantID,
				clientUUID,
				action,
				resourceType,
				nil, // resource_id can be populated by specific handlers if needed
				details,
				&ipAddress,
				&userAgent,
			)

			if err != nil {
				logger.Errorf("Failed to log audit entry: %v", err)
				// Don't fail the request if audit logging fails
			} else {
				logger.Infof("Audit: %s %s %s by %s", action, resourceType, tenantID, employee.Email)
			}

			// Continue with the request
			next.ServeHTTP(w, r)
		})
	}
}

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) string {
	// Try X-Forwarded-For header first (for requests behind proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	// RemoteAddr format is "IP:port" or "[IPv6]:port", we only want the IP
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	// If SplitHostPort fails, return as-is (shouldn't happen)
	return r.RemoteAddr
}
