package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/logger"
	"google.golang.org/api/option"
)

// StorageProvider defines the interface for cloud storage operations
type StorageProvider interface {
	Upload(ctx context.Context, bucket, path string, file io.Reader, metadata map[string]string) error
	Download(ctx context.Context, bucket, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, path string) error
	GetSignedURL(ctx context.Context, bucket, path string, expiration time.Duration) (string, error)
}

// GCSProvider implements StorageProvider for Google Cloud Storage
type GCSProvider struct {
	client *storage.Client
}

// NewGCSProvider creates a new GCS storage provider using Application Default Credentials (ADC)
func NewGCSProvider(ctx context.Context) (*GCSProvider, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client with ADC: %w", err)
	}

	return &GCSProvider{client: client}, nil
}

// NewGCSProviderFromJSON creates a new GCS storage provider from service account JSON
func NewGCSProviderFromJSON(ctx context.Context, jsonData []byte) (*GCSProvider, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client from JSON: %w", err)
	}

	logger.Info("Created GCS client from JSON credentials")
	return &GCSProvider{client: client}, nil
}

// NewGCSProviderFromFile creates a new GCS storage provider from a credentials file
func NewGCSProviderFromFile(ctx context.Context, credentialsPath string) (*GCSProvider, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client from file: %w", err)
	}

	logger.Infof("Created GCS client from file: %s", credentialsPath)
	return &GCSProvider{client: client}, nil
}

// Upload uploads a file to GCS
func (g *GCSProvider) Upload(ctx context.Context, bucket, path string, file io.Reader, metadata map[string]string) error {
	logger.Infof("Uploading file to gs://%s/%s", bucket, path)

	wc := g.client.Bucket(bucket).Object(path).NewWriter(ctx)

	// Set metadata
	if metadata != nil {
		wc.Metadata = metadata
	}

	// Copy file content
	if _, err := io.Copy(wc, file); err != nil {
		wc.Close()
		return fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	logger.Infof("Successfully uploaded file to gs://%s/%s", bucket, path)
	return nil
}

// Download retrieves a file from GCS
func (g *GCSProvider) Download(ctx context.Context, bucket, path string) (io.ReadCloser, error) {
	logger.Infof("Downloading file from gs://%s/%s", bucket, path)

	rc, err := g.client.Bucket(bucket).Object(path).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read from GCS: %w", err)
	}

	return rc, nil
}

// Delete removes a file from GCS
func (g *GCSProvider) Delete(ctx context.Context, bucket, path string) error {
	logger.Infof("Deleting file from gs://%s/%s", bucket, path)

	if err := g.client.Bucket(bucket).Object(path).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}

	logger.Infof("Successfully deleted file from gs://%s/%s", bucket, path)
	return nil
}

// GetSignedURL generates a signed URL for temporary access to a file
func (g *GCSProvider) GetSignedURL(ctx context.Context, bucket, path string, expiration time.Duration) (string, error) {
	logger.Infof("Generating signed URL for gs://%s/%s (expires in %v)", bucket, path, expiration)

	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := g.client.Bucket(bucket).SignedURL(path, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// Close closes the GCS client
func (g *GCSProvider) Close() error {
	return g.client.Close()
}
