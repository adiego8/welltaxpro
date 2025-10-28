package portal

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/logger"
	"github.com/google/uuid"
)

// PortalClaims represents the JWT claims for client portal access
type PortalClaims struct {
	ClientID  string `json:"client_id"`
	TenantID  string `json:"tenant_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"` // "magic_link" or "session"
	jwt.RegisteredClaims
}

// GeneratePortalToken creates a JWT token for client portal access
func GeneratePortalToken(clientID, tenantID, email, jwtSecret string) (string, error) {
	// Token expires in 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &PortalClaims{
		ClientID: clientID,
		TenantID: tenantID,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "welltaxpro",
			Subject:   fmt.Sprintf("portal:%s:%s", tenantID, clientID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		logger.Errorf("Failed to sign portal token: %v", err)
		return "", err
	}

	logger.Infof("Generated portal token for client %s in tenant %s", clientID, tenantID)
	return tokenString, nil
}

// ValidatePortalToken validates and parses a portal access token
func ValidatePortalToken(tokenString, jwtSecret string) (*PortalClaims, error) {
	claims := &PortalClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		logger.Errorf("Failed to parse portal token: %v", err)
		return nil, err
	}

	if !token.Valid {
		logger.Error("Portal token is invalid")
		return nil, fmt.Errorf("invalid token")
	}

	logger.Infof("Validated portal token for client %s in tenant %s", claims.ClientID, claims.TenantID)
	return claims, nil
}

// GenerateMagicLinkToken creates a JWT token for magic link (24 hours)
// This token is one-time use and must be exchanged for a session token
func GenerateMagicLinkToken(clientID, tenantID, email, jwtSecret string) (tokenString string, tokenID string, err error) {
	// Generate unique UUID for tracking
	tokenUUID := uuid.New()
	tokenID = tokenUUID.String()

	// Token expires in 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &PortalClaims{
		ClientID:  clientID,
		TenantID:  tenantID,
		Email:     email,
		TokenType: "magic_link",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "welltaxpro",
			Subject:   fmt.Sprintf("magic:%s:%s", tenantID, clientID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(jwtSecret))
	if err != nil {
		logger.Errorf("Failed to sign magic link token: %v", err)
		return "", "", err
	}

	logger.Infof("Generated magic link token (ID: %s) for client %s in tenant %s", tokenID, clientID, tenantID)
	return tokenString, tokenID, nil
}

// GenerateSessionToken creates a session JWT token (2 hours)
// This token is issued after magic link exchange and SSN verification
func GenerateSessionToken(clientID, tenantID, email, jwtSecret string) (string, error) {
	// Token expires in 2 hours
	expirationTime := time.Now().Add(2 * time.Hour)

	claims := &PortalClaims{
		ClientID:  clientID,
		TenantID:  tenantID,
		Email:     email,
		TokenType: "session",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "welltaxpro",
			Subject:   fmt.Sprintf("session:%s:%s", tenantID, clientID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		logger.Errorf("Failed to sign session token: %v", err)
		return "", err
	}

	logger.Infof("Generated session token for client %s in tenant %s", clientID, tenantID)
	return tokenString, nil
}

// GenerateSecureToken generates a cryptographically secure random token
// This can be used as an additional layer of security if needed
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
