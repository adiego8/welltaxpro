package signature

import (
	"context"
	"fmt"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

type Signature struct {
	TaxPayerEmail      string
	TaxPayerName       string
	SpouseName         string
	SpouseEmail        string
	TaxPayerSsn        string
	TaxPayerSpouseName *string
	TaxPayerSpouseSsn  *string
	GrossIncome        float64
	TotalTax           float64
	TaxWithHeld        float64
	Refund             float64
	Owed               float64
	SignatureDate      *string
	SpouseSignature    bool
}

// SignDocument requests a signature from DocuSign using tenant configuration
// pdfPath is the path to the Form 8879 PDF file to sign
func SignDocument(ctx context.Context, tc *types.TenantConnection, pdfPath string, s *Signature) error {
	logger.Info("Starting Signature Request")

	// Validate tenant has DocuSign configured
	if tc.DocuSignIntegrationKey == "" || tc.DocuSignClientID == "" || tc.DocuSignPrivateKeySecret == "" {
		return fmt.Errorf("tenant %s does not have DocuSign configured", tc.TenantID)
	}

	// Get DocuSign access token using JWT
	dSAccessToken, err := makeDSToken(ctx, tc.DocuSignIntegrationKey, tc.DocuSignClientID, tc.DocuSignPrivateKeySecret)
	if err != nil {
		logger.Errorf("Failed to retrieve token: %v", err)
		return fmt.Errorf("failed to get DocuSign token: %w", err)
	}

	maskedToken := fmt.Sprintf("%s...%s", dSAccessToken[:3], dSAccessToken[len(dSAccessToken)-3:])
	logger.Infof("Getting account with token: %s", maskedToken)

	// Get DocuSign account ID
	dSAccountId, err := getAPIAccId(dSAccessToken)
	if err != nil {
		logger.Errorf("Failed to get API Account ID: %v", err)
		return fmt.Errorf("failed to get account ID: %w", err)
	}

	logger.Info("Signature auth completed")

	// Build envelope API URL
	apiURL := fmt.Sprintf("%s/v2.1/accounts/%s/envelopes", tc.DocuSignAPIURL, dSAccountId)

	// Send envelope for signature
	err = sendEnvelope(ctx, dSAccessToken, apiURL, tc, pdfPath, s)
	if err != nil {
		logger.Errorf("Failed to request signature: %v", err)
		return fmt.Errorf("failed to send envelope: %w", err)
	}

	logger.Info("Signature request sent successfully")
	return nil
}
