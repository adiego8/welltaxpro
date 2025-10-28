package auth

import (
	"context"

	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/logger"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type contextKey string

const (
	// AuthorizationKey is the authorization header key.
	AuthorizationKey contextKey = "Authorization"
	// EmployeeContextKey stores the authenticated employee in request context
	EmployeeContextKey contextKey = "Employee"
)

// Initialize Firebase App and Auth
type Auth struct {
	Client      *auth.Client
	FirebaseKey string
}

type firebaseTokenResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
}

var firebaseAuth *auth.Client

func InitAuth(firebaseKey, serviceAccountPath string) (*Auth, error) {
	// Initialize Firebase SDK using a service account key file
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	// Initialize Firebase Auth
	firebaseAuth, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
		return nil, err
	}

	// Make sure firebaseAuth is initialized
	if firebaseAuth == nil {
		return nil, fmt.Errorf("firebaseAuth is not initialized")
	}

	return &Auth{
		Client:      firebaseAuth,
		FirebaseKey: firebaseKey,
	}, nil
}

// ValidateToken to ensure that token provided is valid and user can
// access the API, it returns the token UID.
func (a *Auth) ValidateToken(ctx context.Context, token string) (*string, error) {
	logger.Info("Verifying token")

	// Remove Bearer prefix if present
	cleanToken := token
	if len(token) > 7 && token[:7] == "Bearer " {
		cleanToken = token[7:]
	}

	// Check if this is already an ID token by trying to verify it directly first
	decodedToken, err := firebaseAuth.VerifyIDToken(ctx, cleanToken)
	if err != nil {
		// If direct verification fails, try to exchange as custom token
		logger.Info("Direct token verification failed, attempting custom token exchange")

		exchangedToken, exchangeErr := exchangeCustomTokenForIDToken(cleanToken, a.FirebaseKey)
		if exchangeErr != nil {
			logger.Errorf("Failed to exchange custom token: %v", exchangeErr)
			logger.Errorf("Original verification error: %v", err)
			return nil, err // Return the original verification error
		}

		if exchangedToken == "" {
			logger.Error("Custom token exchange returned empty token")
			return nil, err
		}

		logger.Info("Custom token exchanged successfully")

		// Verify the exchanged token
		decodedToken, err = firebaseAuth.VerifyIDToken(ctx, exchangedToken)
		if err != nil {
			logger.Errorf("Error verifying exchanged token: %v", err)
			return nil, err
		}
	} else {
		logger.Info("Token verified directly as ID token")
	}

	return &decodedToken.UID, nil
}

func exchangeCustomTokenForIDToken(customToken, firebaseAPIKey string) (string, error) {
	// Firebase REST API endpoint for exchanging custom token
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=%s", firebaseAPIKey)

	// Request payload
	payload := map[string]string{
		"token":             customToken,
		"returnSecureToken": "true",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Make the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to make request to Firebase: %v", err)
	}
	defer resp.Body.Close()

	// Decode the response
	var tokenResponse firebaseTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	// Return the ID token
	return tokenResponse.IDToken, nil
}

type AuthUser struct {
	UID           string
	Email         string
	EmailVerified bool
}

func (a *Auth) CreateUser(ctx context.Context, email string, password string) (*AuthUser, error) {
	// Create a new user
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password)

	u, err := a.Client.CreateUser(ctx, params)
	if err != nil {
		logger.Error("failed creating user in auth package")
		return nil, err
	}

	return &AuthUser{UID: u.UID, Email: u.Email, EmailVerified: u.EmailVerified}, nil
}

type SignInRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

type SignInResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
}

type RefreshTokenRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	UserID       string `json:"user_id"`
	TokenType    string `json:"token_type"`
}

func (a *Auth) DeleteUser(ctx context.Context, uid string) error {
	logger.Info("Deleting user")
	err := a.Client.DeleteUser(ctx, uid)
	if err != nil {
		logger.Infof("Failed deleting user with uid: %s\n Error: %v", uid, err)
		return err
	}

	return nil
}

func (a *Auth) SignInWithEmailAndPassword(email, password string) (*SignInResponse, error) {
	logger.Info("Sign in with email and password")

	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", a.FirebaseKey)

	// Create the sign-in request payload
	signInRequest := SignInRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	requestBody, err := json.Marshal(signInRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sign-in request: %v", err)
	}

	// Make the HTTP POST request to Firebase Identity Toolkit API
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make sign-in request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseBody bytes.Buffer
		_, _ = responseBody.ReadFrom(resp.Body)
		logger.Errorf("Sign-in request failed with status: %v, response: %s", resp.StatusCode, responseBody.String())
		return nil, fmt.Errorf("sign-in request failed with status: %v", resp.StatusCode)
	}

	// Decode the response body into SignInResponse struct
	var signInResponse SignInResponse
	if err := json.NewDecoder(resp.Body).Decode(&signInResponse); err != nil {
		return nil, fmt.Errorf("failed to decode sign-in response: %v", err)
	}

	return &signInResponse, nil
}

// RefreshToken uses Firebase's refresh token to get a new ID token
func (a *Auth) RefreshToken(refreshToken string) (*RefreshTokenResponse, error) {
	logger.Info("Refreshing Firebase token")

	url := fmt.Sprintf("https://securetoken.googleapis.com/v1/token?key=%s", a.FirebaseKey)

	// Create the refresh token request payload
	refreshRequest := RefreshTokenRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}

	requestBody, err := json.Marshal(refreshRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refresh token request: %v", err)
	}

	// Make the HTTP POST request to Firebase Secure Token API
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make refresh token request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseBody bytes.Buffer
		_, _ = responseBody.ReadFrom(resp.Body)
		logger.Errorf("Refresh token request failed with status: %v, response: %s", resp.StatusCode, responseBody.String())
		return nil, fmt.Errorf("refresh token request failed with status: %v", resp.StatusCode)
	}

	// Decode the response body into RefreshTokenResponse struct
	var refreshResponse RefreshTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshResponse); err != nil {
		return nil, fmt.Errorf("failed to decode refresh token response: %v", err)
	}

	logger.Info("Token refreshed successfully")
	return &refreshResponse, nil
}
