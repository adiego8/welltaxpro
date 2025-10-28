package signature

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"welltaxpro/src/internal/secrets"

	"github.com/golang-jwt/jwt"
	"github.com/google/logger"
)

type AccessToken struct {
	Token  string `json:"access_token"`
	Type   string `json:"token_type"`
	Expiry int    `json:"expires_in"`
}

// Auto-generated using https://transform.tools/json-to-go
type AccountId struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Created    string `json:"created"`
	Email      string `json:"email"`
	Accounts   []struct {
		AccountID    string `json:"account_id"`
		IsDefault    bool   `json:"is_default"`
		AccountName  string `json:"account_name"`
		BaseURI      string `json:"base_uri"`
		Organization struct {
			OrganizationID string `json:"organization_id"`
			Links          []struct {
				Rel  string `json:"rel"`
				Href string `json:"href"`
			} `json:"links"`
		} `json:"organization"`
	} `json:"accounts"`
}

// makeDSToken creates a DocuSign JWT access token using tenant configuration
// privateKeySecret is the GCP Secret Manager path to the RSA private key
func makeDSToken(ctx context.Context, integrationKey, clientId, privateKeySecret string) (string, error) {
	logger.Info("Getting DS Token")

	// Create a new JWT claim. Set your integration key, impersonated user GUID, time of issue, expiry time, account server, and required scopes
	rawJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   integrationKey,
		"sub":   clientId,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Unix() + 3600,
		"aud":   "account.docusign.com",
		"scope": "signature impersonation",
	})

	// Get private key from Secret Manager or local file
	var RSAPrivateKey []byte
	var err error

	// Check if it's a Secret Manager path (starts with "projects/") or a local file path
	if strings.HasPrefix(privateKeySecret, "projects/") {
		// Use Secret Manager
		logger.Infof("Reading DocuSign private key from Secret Manager: %s", privateKeySecret)
		secretManager, err := secrets.GetSecretManager(ctx)
		if err != nil {
			logger.Errorf("Failed to get Secret Manager: %v", err)
			return "", fmt.Errorf("failed to get secret manager: %w", err)
		}

		RSAPrivateKey, err = secretManager.GetSecret(ctx, privateKeySecret)
		if err != nil {
			logger.Errorf("Failed to get DocuSign private key from Secret Manager: %v", err)
			return "", fmt.Errorf("failed to get private key: %w", err)
		}
	} else {
		// Read from local file
		logger.Infof("Reading DocuSign private key from local file: %s", privateKeySecret)
		RSAPrivateKey, err = os.ReadFile(privateKeySecret)
		if err != nil {
			logger.Errorf("Failed to read DocuSign private key from file: %v", err)
			return "", fmt.Errorf("failed to read private key file: %w", err)
		}
	}

	// Load the private key into JWT library
	rsaPrivate, err := jwt.ParseRSAPrivateKeyFromPEM(RSAPrivateKey)
	if err != nil {
		logger.Errorf("Failed to parse RSA private key: %v", err)
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Generate the signed JSON Web Token assertion with an RSA private key
	tokenString, err := rawJWT.SignedString(rsaPrivate)
	if err != nil {
		logger.Errorf("Failed creating signed JSON with RSA Key: %v", err)
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Submit the JWT to the account server and request access token
	resp, err := http.PostForm("https://account.docusign.com/oauth/token",
		url.Values{
			"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
			"assertion":  {tokenString},
		})
	if err != nil {
		logger.Errorf("Request Failed: %v", err)
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	logger.Infof("Response from Auth status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Failed reading response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Log the raw response for debugging
	logger.Infof("DocuSign auth response body: %s", string(body))

	// Decode the response to JSON
	var token AccessToken
	jsonErr := json.Unmarshal(body, &token)
	if jsonErr != nil {
		logger.Errorf("There was an error decoding the json. err = %v", jsonErr)
		return "", fmt.Errorf("failed to decode token: %w", jsonErr)
	}

	if token.Token == "" {
		logger.Errorf("Empty token received. Full response: %s", string(body))
		return "", fmt.Errorf("failed to get DS token: empty token in response")
	}

	return token.Token, nil
}

// getAPIAccId retrieves the API account ID GUID used to make all subsequent API calls
func getAPIAccId(DSAccessToken string) (string, error) {
	client := &http.Client{}
	// Use http.NewRequest in order to set custom headers
	req, err := http.NewRequest("GET", "https://account.docusign.com/oauth/userinfo", nil)
	if err != nil {
		logger.Errorf("Request Failed: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+DSAccessToken)

	// Since http.NewRequest is being used, client.Do is needed to execute the request
	res, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed connecting to client: %v", err)
		return "", fmt.Errorf("failed to get user info: %w", err)
	}
	defer res.Body.Close()

	logger.Info("Reading signature account id")

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Errorf("Failed reading response body: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Decode the response to JSON
	var accountId AccountId
	jsonErr := json.Unmarshal(body, &accountId)
	if jsonErr != nil {
		logger.Errorf("There was an error decoding the json. err = %v", jsonErr)
		return "", fmt.Errorf("failed to decode account info: %w", jsonErr)
	}

	if len(accountId.Accounts) == 0 {
		return "", fmt.Errorf("no accounts found for user")
	}

	return accountId.Accounts[0].AccountID, nil
}
