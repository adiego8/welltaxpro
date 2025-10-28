package secrets

import (
	"context"
	"fmt"
	"sync"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/google/logger"
)

// CachedSecret holds a secret value with expiration
type CachedSecret struct {
	Data      []byte
	ExpiresAt time.Time
}

// SecretManager manages secrets with in-memory caching
type SecretManager struct {
	client *secretmanager.Client
	cache  map[string]*CachedSecret
	mutex  sync.RWMutex
	ttl    time.Duration
}

var (
	instance *SecretManager
	initErr  error
	once     sync.Once
)

// GetSecretManager returns the singleton SecretManager instance
func GetSecretManager(ctx context.Context) (*SecretManager, error) {
	once.Do(func() {
		instance, initErr = newSecretManager(ctx)
	})
	return instance, initErr
}

// newSecretManager creates a new SecretManager
func newSecretManager(ctx context.Context) (*SecretManager, error) {
	// Uses Application Default Credentials (ADC)
	// - Cloud Run: Automatically uses service account
	// - Local dev: Uses `gcloud auth application-default login`
	// - Override: Set GOOGLE_APPLICATION_CREDENTIALS env var
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	logger.Info("Secret Manager client initialized successfully")

	return &SecretManager{
		client: client,
		cache:  make(map[string]*CachedSecret),
		ttl:    1 * time.Hour, // 1 hour cache TTL
	}, nil
}

// GetSecret retrieves a secret by its full path, using cache if available
// secretPath format: "projects/PROJECT_ID/secrets/SECRET_NAME/versions/VERSION"
func (sm *SecretManager) GetSecret(ctx context.Context, secretPath string) ([]byte, error) {
	// Check cache first
	sm.mutex.RLock()
	if cached, exists := sm.cache[secretPath]; exists {
		if time.Now().Before(cached.ExpiresAt) {
			sm.mutex.RUnlock()
			logger.Infof("Secret cache hit: %s", secretPath)
			return cached.Data, nil
		}
		// Expired, will refetch
		logger.Infof("Secret cache expired: %s", secretPath)
	}
	sm.mutex.RUnlock()

	// Cache miss or expired - fetch from Secret Manager
	logger.Infof("Fetching secret from Secret Manager: %s", secretPath)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	result, err := sm.client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Errorf("Failed to access secret %s: %v", secretPath, err)
		return nil, fmt.Errorf("failed to access secret: %w", err)
	}

	secretData := result.Payload.Data

	// Store in cache
	sm.mutex.Lock()
	sm.cache[secretPath] = &CachedSecret{
		Data:      secretData,
		ExpiresAt: time.Now().Add(sm.ttl),
	}
	sm.mutex.Unlock()

	logger.Infof("Secret fetched and cached: %s (expires in %v)", secretPath, sm.ttl)
	return secretData, nil
}

// ClearCache removes a specific secret from cache (useful for testing/rotation)
func (sm *SecretManager) ClearCache(secretPath string) {
	sm.mutex.Lock()
	delete(sm.cache, secretPath)
	sm.mutex.Unlock()
	logger.Infof("Cleared cache for secret: %s", secretPath)
}

// ClearAllCache removes all cached secrets
func (sm *SecretManager) ClearAllCache() {
	sm.mutex.Lock()
	sm.cache = make(map[string]*CachedSecret)
	sm.mutex.Unlock()
	logger.Info("Cleared all secret cache")
}

// Close closes the Secret Manager client
func (sm *SecretManager) Close() error {
	return sm.client.Close()
}
