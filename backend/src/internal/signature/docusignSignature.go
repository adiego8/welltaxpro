package signature

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"welltaxpro/src/internal/storage"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

type EnvelopeDefinition struct {
	EmailSubject string     `json:"emailSubject"`
	Documents    []Document `json:"documents"`
	Recipients   Recipients `json:"recipients"`
	Status       string     `json:"status"`
}

type Document struct {
	DocumentBase64 string `json:"documentBase64"`
	Name           string `json:"name"`
	FileExtension  string `json:"fileExtension"`
	DocumentID     string `json:"documentId"`
}

type Recipients struct {
	Signers []Signer `json:"signers"`
}

type Signer struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	RecipientID  string `json:"recipientId"`
	ClientUserID string `json:"clientUserId"`
	Tabs         Tabs   `json:"tabs"`
}

type Tabs struct {
	SignHereTabs   []SignHere   `json:"signHereTabs"`
	DateSignedTabs []DateSigned `json:"dateSignedTabs"`
	TextTabs       []Text       `json:"textTabs"`
}

type SignHere struct {
	XPosition  string `json:"xPosition"`
	YPosition  string `json:"yPosition"`
	DocumentID string `json:"documentId"`
	PageNumber string `json:"pageNumber"`
}

type DateSigned struct {
	XPosition  string `json:"xPosition"`
	YPosition  string `json:"yPosition"`
	DocumentID string `json:"documentId"`
	PageNumber string `json:"pageNumber"`
}

type Text struct {
	XPosition  string `json:"xPosition"`
	YPosition  string `json:"yPosition"`
	DocumentID string `json:"documentId"`
	PageNumber string `json:"pageNumber"`
	Value      string `json:"value"`
	Locked     bool   `json:"locked"`
}

// Auto-generated using https://transform.tools/json-to-go
type EnvelopeID struct {
	EnvelopeID     string    `json:"envelopeId"`
	URI            string    `json:"uri"`
	StatusDateTime time.Time `json:"statusDateTime"`
	Status         string    `json:"status"`
}

// encodePDFToBase64 reads a PDF file and encodes it to a Base64 string
// Handles both local file paths and GCS URLs
func encodePDFToBase64(ctx context.Context, tc *types.TenantConnection, filePath string) (string, error) {
	var pdfBytes []byte
	var err error

	// Check if it's a GCS URL
	if strings.HasPrefix(filePath, "https://storage.googleapis.com/") || strings.HasPrefix(filePath, "gs://") {
		// Parse GCS URL to extract bucket and path
		bucket, path := parseGCSURL(filePath)
		if bucket == "" || path == "" {
			return "", fmt.Errorf("invalid GCS URL format: %s", filePath)
		}

		logger.Infof("Downloading PDF from GCS: gs://%s/%s", bucket, path)

		// Create storage provider
		storageProvider, err := storage.NewStorageProviderForTenant(ctx, tc)
		if err != nil {
			return "", fmt.Errorf("failed to create storage provider: %w", err)
		}

		// Download file from GCS
		reader, err := storageProvider.Download(ctx, bucket, path)
		if err != nil {
			return "", fmt.Errorf("failed to download from GCS: %w", err)
		}
		defer reader.Close()

		// Read all bytes
		pdfBytes, err = io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read GCS file: %w", err)
		}

		logger.Infof("Successfully downloaded PDF from GCS (%d bytes)", len(pdfBytes))
	} else {
		// Read from local file
		logger.Infof("Reading PDF from local file: %s", filePath)
		pdfBytes, err = os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("error reading PDF file: %w", err)
		}
	}

	// Encode to Base64
	base64Str := base64.StdEncoding.EncodeToString(pdfBytes)

	return base64Str, nil
}

// parseGCSURL extracts bucket and path from GCS URL
// Supports both https://storage.googleapis.com/{bucket}/{path} and gs://{bucket}/{path}
func parseGCSURL(url string) (bucket, path string) {
	if strings.HasPrefix(url, "gs://") {
		// gs://bucket/path/to/file.pdf
		parts := strings.SplitN(strings.TrimPrefix(url, "gs://"), "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	} else if strings.HasPrefix(url, "https://storage.googleapis.com/") {
		// https://storage.googleapis.com/bucket/path/to/file.pdf
		parts := strings.SplitN(strings.TrimPrefix(url, "https://storage.googleapis.com/"), "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return "", ""
}

func sendEnvelope(ctx context.Context, accessToken, apiURL string, tc *types.TenantConnection, pdfPath string, s *Signature) error {
	// Convert the PDF file to Base64
	docBase64, err := encodePDFToBase64(ctx, tc, pdfPath)
	if err != nil {
		logger.Errorf("Error encoding PDF: %v", err)
		return fmt.Errorf("failed to encode PDF: %w", err)
	}

	gi := strconv.FormatFloat(s.GrossIncome, 'f', 2, 64)
	tt := strconv.FormatFloat(s.TotalTax, 'f', 2, 64)
	tw := strconv.FormatFloat(s.TaxWithHeld, 'f', 2, 64)
	rf := strconv.FormatFloat(s.Refund, 'f', 2, 64)
	ow := strconv.FormatFloat(s.Owed, 'f', 2, 64)

	taxPayerTabs := []Text{
		{
			XPosition:  "85",
			YPosition:  "125",
			DocumentID: "1",
			PageNumber: "1",
			Value:      s.TaxPayerName,
			Locked:     true,
		},
		{
			XPosition:  "450",
			YPosition:  "128",
			DocumentID: "1",
			PageNumber: "1",
			Value:      s.TaxPayerSsn,
			Locked:     true,
		},
		// Tax Information
		{
			XPosition:  "502",
			YPosition:  "200",
			DocumentID: "1",
			PageNumber: "1",
			Value:      gi,
			Locked:     true,
		},
		{
			XPosition:  "502",
			YPosition:  "213",
			DocumentID: "1",
			PageNumber: "1",
			Value:      tt,
			Locked:     true,
		},
		{
			XPosition:  "502",
			YPosition:  "226",
			DocumentID: "1",
			PageNumber: "1",
			Value:      tw,
			Locked:     true,
		},
		{
			XPosition:  "502",
			YPosition:  "239",
			DocumentID: "1",
			PageNumber: "1",
			Value:      rf,
			Locked:     true,
		},
		{
			XPosition:  "502",
			YPosition:  "252",
			DocumentID: "1",
			PageNumber: "1",
			Value:      ow,
			Locked:     true,
		},
	}

	// Taxpayer Signer
	taxpayerSigner := Signer{
		Email:       s.TaxPayerEmail,
		Name:        s.TaxPayerName,
		RecipientID: "1",
		Tabs: Tabs{
			SignHereTabs: []SignHere{
				{
					XPosition:  "130",
					YPosition:  "450",
					DocumentID: "1",
					PageNumber: "1",
				},
			},
			DateSignedTabs: []DateSigned{
				{
					XPosition:  "450",
					YPosition:  "465",
					DocumentID: "1",
					PageNumber: "1",
				},
			},
			TextTabs: taxPayerTabs,
		},
	}

	// Spouse Signer (Only if SpouseSignature is required)
	var spouseSigner Signer
	if s.SpouseSignature {
		spouseSigner = Signer{
			Email:       s.SpouseEmail,
			Name:        s.SpouseName,
			RecipientID: "2",
			Tabs: Tabs{
				SignHereTabs: []SignHere{
					{
						XPosition:  "130",
						YPosition:  "580",
						DocumentID: "1",
						PageNumber: "1",
					},
				},
				DateSignedTabs: []DateSigned{
					{
						XPosition:  "450",
						YPosition:  "590",
						DocumentID: "1",
						PageNumber: "1",
					},
				},
				TextTabs: taxPayerTabs,
			},
		}
	}

	signers := []Signer{taxpayerSigner}
	if s.SpouseSignature {
		signers = append(signers, spouseSigner)
	}

	envelope := EnvelopeDefinition{
		EmailSubject: "Please sign this document",
		Documents: []Document{
			{
				DocumentBase64: docBase64,
				Name:           "Form 8879",
				FileExtension:  "pdf",
				DocumentID:     "1",
			},
		},
		Recipients: Recipients{
			Signers: signers,
		},
		Status: "sent",
	}

	// Convert struct to JSON
	jsonData, err := json.Marshal(envelope)
	if err != nil {
		logger.Errorf("Error encoding JSON: %v", err)
		return fmt.Errorf("failed to encode envelope: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Error creating request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Error sending request: %v", err)
		return fmt.Errorf("failed to send envelope: %w", err)
	}
	defer resp.Body.Close()

	logger.Infof("Response status: %s", resp.Status)

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Error reading response: %v", err)
		return fmt.Errorf("failed to read response: %w", err)
	}

	logger.Infof("Response: %s", string(body))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("DocuSign API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
