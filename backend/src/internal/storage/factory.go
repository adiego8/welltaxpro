package storage

import (
	"context"
	"fmt"
	"os"
	"welltaxpro/src/internal/secrets"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
)

// NewStorageProviderForTenant creates a storage provider for a tenant with priority cascade:
// 1. Try StorageCredentialsSecret (fetch from Secret Manager)
// 2. Fallback to StorageCredentialsPath (read from file - local dev)
// 3. Fallback to ADC (Application Default Credentials)
func NewStorageProviderForTenant(ctx context.Context, tc *types.TenantConnection) (StorageProvider, error) {
	if tc.StorageProvider != "gcs" {
		return nil, fmt.Errorf("unsupported storage provider: %s", tc.StorageProvider)
	}

	// Priority 1: Try Secret Manager (production)
	if tc.StorageCredentialsSecret != "" {
		logger.Infof("Attempting to create GCS provider from Secret Manager: %s", tc.StorageCredentialsSecret)

		secretManager, err := secrets.GetSecretManager(ctx)
		if err != nil {
			logger.Warningf("Failed to initialize Secret Manager, falling back: %v", err)
		} else {
			secretData, err := secretManager.GetSecret(ctx, tc.StorageCredentialsSecret)
			if err != nil {
				logger.Warningf("Failed to fetch secret from Secret Manager, falling back: %v", err)
			} else {
				provider, err := NewGCSProviderFromJSON(ctx, secretData)
				if err != nil {
					logger.Warningf("Failed to create GCS client from secret JSON, falling back: %v", err)
				} else {
					logger.Infof("Successfully created GCS provider from Secret Manager for tenant %s", tc.TenantID)
					return provider, nil
				}
			}
		}
	}

	// Priority 2: Try file path (local development)
	if tc.StorageCredentialsPath != "" {
		logger.Infof("Attempting to create GCS provider from file: %s", tc.StorageCredentialsPath)

		// Check if file exists
		if _, err := os.Stat(tc.StorageCredentialsPath); err == nil {
			provider, err := NewGCSProviderFromFile(ctx, tc.StorageCredentialsPath)
			if err != nil {
				logger.Warningf("Failed to create GCS client from file, falling back to ADC: %v", err)
			} else {
				logger.Infof("Successfully created GCS provider from file for tenant %s", tc.TenantID)
				return provider, nil
			}
		} else {
			logger.Infof("Credentials file not found at %s, falling back to ADC", tc.StorageCredentialsPath)
		}
	}

	// Priority 3: Use Application Default Credentials (ADC)
	logger.Infof("Attempting to create GCS provider using ADC for tenant %s", tc.TenantID)
	provider, err := NewGCSProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS provider with ADC: %w", err)
	}

	logger.Infof("Successfully created GCS provider using ADC for tenant %s", tc.TenantID)
	return provider, nil
}
