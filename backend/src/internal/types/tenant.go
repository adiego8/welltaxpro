package types

import (
	"fmt"

	"github.com/google/uuid"
)

// TenantConnection represents a tenant's database connection configuration
type TenantConnection struct {
	ID           uuid.UUID `json:"id"`
	TenantID     string    `json:"tenantId"`
	TenantName   string    `json:"tenantName"`
	DBHost       string    `json:"dbHost"`
	DBPort       int       `json:"dbPort"`
	DBUser       string    `json:"dbUser"`
	DBPassword   string    `json:"-"` // Never expose in JSON
	DBName       string    `json:"dbName"`
	DBSslMode    string    `json:"dbSslMode"`
	SchemaPrefix             string  `json:"schemaPrefix"`
	AdapterType              string  `json:"adapterType"` // Adapter to use (mywelltax, drake, lacerte, etc.)
	StorageProvider          string  `json:"storageProvider"` // Storage provider (gcs, s3, azure)
	StorageBucket            string  `json:"storageBucket"` // Bucket/container name for document storage
	StorageCredentialsSecret string  `json:"-"` // GCP Secret Manager path (e.g., "projects/PROJECT/secrets/NAME/versions/VERSION")
	StorageCredentialsPath   string  `json:"-"` // Fallback: Path to service account JSON file (never exposed in JSON)
	DocuSignIntegrationKey   string  `json:"docusignIntegrationKey"` // DocuSign Integration Key
	DocuSignClientID         string  `json:"docusignClientId"` // DocuSign Client ID / User ID for JWT auth
	DocuSignPrivateKeySecret string  `json:"-"` // GCP Secret Manager path to DocuSign RSA private key (never exposed in JSON)
	DocuSignAPIURL           string  `json:"docusignApiUrl"` // DocuSign API base URL (demo or production)
	IsActive                 bool    `json:"isActive"`
	CreatedAt              string  `json:"createdAt"`
	UpdatedAt              string  `json:"updatedAt"`
	CreatedBy              *string `json:"createdBy"`
	Notes                  *string `json:"notes"`
}

// GetConnectionString returns a PostgreSQL connection string for this tenant
func (tc *TenantConnection) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s binary_parameters=yes",
		tc.DBHost, tc.DBPort, tc.DBUser, tc.DBPassword, tc.DBName, tc.DBSslMode)
}
