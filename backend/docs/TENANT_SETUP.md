# WellTaxPro Tenant Setup Guide

This document provides all SQL queries required to set up a new tenant in WellTaxPro.

## Prerequisites

Before running these queries, ensure you have:

1. **Tenant's database** already created and accessible
2. **GCP Service Account** credentials for storage access
3. **DocuSign credentials** (Integration Key, Client ID, Private Key)
4. **Storage bucket** created (GCS/S3/Azure)

## Complete Setup Example: MyWellTax Tenant

### 1. Initial Tenant Connection Setup

```sql
-- Create tenant connection with all configurations
INSERT INTO tenant_connections (
    -- Identity
    tenant_id,
    tenant_name,

    -- Database Connection
    db_host,
    db_port,
    db_user,
    db_password,
    db_name,
    db_sslmode,
    schema_prefix,

    -- Platform Configuration
    adapter_type,
    is_active,
    created_by,
    notes,

    -- Storage Configuration
    storage_provider,
    storage_bucket,
    storage_credentials_secret,
    storage_credentials_path,

    -- DocuSign Configuration
    docusign_integration_key,
    docusign_client_id,
    docusign_private_key_secret,
    docusign_api_url
) VALUES (
    -- Identity
    'mywelltax',                           -- Unique tenant identifier (lowercase, no spaces)
    'MyWellTax',                           -- Display name

    -- Database Connection
    'localhost',                           -- Or remote host: 'db.example.com'
    5432,                                  -- PostgreSQL default port
    'postgres',                            -- Database user
    'password',                            -- Database password (use strong password in production)
    'mywelltax',                          -- Name of tenant's database
    'disable',                             -- Use 'require' for production
    'taxes',                               -- Schema prefix within tenant DB

    -- Platform Configuration
    'mywelltax',                          -- Adapter type (determines data mapping)
    true,                                  -- is_active: true to enable tenant
    'system',                              -- created_by: admin user or 'system'
    'Pilot tenant - MyWellTax admin platform',

    -- Storage Configuration (GCS)
    'gcs',                                                                      -- Provider: 'gcs', 's3', or 'azure'
    'mywelltax-documents-dev',                                                 -- Bucket/container name
    'projects/welltaxpro-prod/secrets/mywelltax-gcs-sa/versions/latest',     -- GCP Secret Manager path to credentials
    '/tmp/mywelltax-sa.json',                                                  -- Fallback: local file path for dev

    -- DocuSign Configuration
    '02662059-8436-404f-80ee-7078d642422b',    -- DocuSign Integration Key
    'f1e95a79-5239-4a94-8855-8f69d9f57912',    -- DocuSign Client ID (User ID for JWT)
    '/var/opt/private.key',                     -- Path to RSA private key or Secret Manager path
    'https://demo.docusign.net/restapi'         -- Demo environment (use production URL for prod)
);
```

### 2. Verify Tenant Setup

```sql
-- Check tenant was created successfully
SELECT
    tenant_id,
    tenant_name,
    db_host,
    db_name,
    schema_prefix,
    adapter_type,
    is_active,
    storage_provider,
    storage_bucket
FROM tenant_connections
WHERE tenant_id = 'mywelltax';
```

### 3. Update Storage Configuration (if needed)

```sql
-- Update storage settings for an existing tenant
UPDATE tenant_connections
SET
    storage_provider = 'gcs',
    storage_bucket = 'mywelltax-documents-dev',
    storage_credentials_secret = 'projects/welltaxpro-prod/secrets/mywelltax-gcs-sa/versions/latest',
    storage_credentials_path = '/tmp/mywelltax-sa.json',
    updated_at = NOW()
WHERE tenant_id = 'mywelltax';
```

### 4. Update DocuSign Configuration (if needed)

```sql
-- Update DocuSign settings for an existing tenant
UPDATE tenant_connections
SET
    docusign_integration_key = '02662059-8436-404f-80ee-7078d642422b',
    docusign_client_id = 'f1e95a79-5239-4a94-8855-8f69d9f57912',
    docusign_private_key_secret = '/var/opt/private.key',
    docusign_api_url = 'https://demo.docusign.net/restapi',
    updated_at = NOW()
WHERE tenant_id = 'mywelltax';
```

### 5. Deactivate/Reactivate Tenant

```sql
-- Deactivate tenant (prevents access without deleting data)
UPDATE tenant_connections
SET
    is_active = false,
    updated_at = NOW()
WHERE tenant_id = 'mywelltax';

-- Reactivate tenant
UPDATE tenant_connections
SET
    is_active = true,
    updated_at = NOW()
WHERE tenant_id = 'mywelltax';
```

## Configuration Reference

### Storage Providers

**Google Cloud Storage (GCS)**:
```sql
storage_provider = 'gcs'
storage_bucket = 'your-bucket-name'
storage_credentials_secret = 'projects/PROJECT_ID/secrets/SECRET_NAME/versions/latest'
```

**Amazon S3**:
```sql
storage_provider = 's3'
storage_bucket = 'your-bucket-name'
storage_credentials_secret = 'projects/PROJECT_ID/secrets/aws-credentials/versions/latest'
```

**Azure Blob Storage**:
```sql
storage_provider = 'azure'
storage_bucket = 'your-container-name'
storage_credentials_secret = 'projects/PROJECT_ID/secrets/azure-credentials/versions/latest'
```

### DocuSign Environments

**Demo/Sandbox**:
```sql
docusign_api_url = 'https://demo.docusign.net/restapi'
```

**Production**:
```sql
docusign_api_url = 'https://www.docusign.net/restapi'
```

### SSL Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `disable` | No SSL | Local development only |
| `require` | SSL required | Production (recommended) |
| `verify-ca` | Verify CA | High security environments |
| `verify-full` | Full verification | Maximum security |

## Security Best Practices

### 1. Use Secret Manager for Credentials

**Store DocuSign Private Key in Secret Manager**:
```bash
# Upload private key to GCP Secret Manager
gcloud secrets create mywelltax-docusign-key \
  --data-file=/path/to/private.key \
  --project=welltaxpro-prod

# Update database to use Secret Manager path
UPDATE tenant_connections
SET docusign_private_key_secret = 'projects/welltaxpro-prod/secrets/mywelltax-docusign-key/versions/latest'
WHERE tenant_id = 'mywelltax';
```

**Store GCS Service Account in Secret Manager**:
```bash
# Upload service account JSON
gcloud secrets create mywelltax-gcs-sa \
  --data-file=/path/to/service-account.json \
  --project=welltaxpro-prod

# Update database
UPDATE tenant_connections
SET storage_credentials_secret = 'projects/welltaxpro-prod/secrets/mywelltax-gcs-sa/versions/latest'
WHERE tenant_id = 'mywelltax';
```

### 2. Use Strong Database Passwords

```sql
-- Generate strong password
-- Use a password manager or: openssl rand -base64 32

UPDATE tenant_connections
SET
    db_password = 'YOUR_STRONG_PASSWORD_HERE',
    updated_at = NOW()
WHERE tenant_id = 'mywelltax';
```

### 3. Enable SSL for Production

```sql
UPDATE tenant_connections
SET
    db_sslmode = 'require',
    updated_at = NOW()
WHERE tenant_id = 'mywelltax';
```

## Multi-Tenant Isolation

WellTaxPro ensures tenant isolation through:

1. **Separate Databases**: Each tenant has its own database
2. **Schema Prefixes**: Each tenant uses a specific schema within their database
3. **Connection Pooling**: Separate connection pools per tenant
4. **Storage Buckets**: Dedicated storage buckets per tenant (recommended)

## Testing Tenant Setup

### 1. Test Database Connection

```bash
# Test connection to tenant database
PGPASSWORD=password psql -h localhost -U postgres -d mywelltax -c "SELECT current_database(), current_schema();"
```

### 2. Test Storage Access

```bash
# List bucket contents (requires gcloud auth)
gsutil ls gs://mywelltax-documents-dev/
```

### 3. Test via API

```bash
# Get clients for tenant (requires Firebase auth token)
curl -H "Authorization: Bearer YOUR_ID_TOKEN" \
  http://localhost:8081/api/v1/mywelltax/clients
```

## Common Operations

### List All Tenants

```sql
SELECT
    tenant_id,
    tenant_name,
    adapter_type,
    is_active,
    storage_provider,
    created_at
FROM tenant_connections
ORDER BY created_at DESC;
```

### Get Full Tenant Configuration

```sql
SELECT * FROM tenant_connections
WHERE tenant_id = 'mywelltax';
```

### Update Tenant Metadata

```sql
UPDATE tenant_connections
SET
    tenant_name = 'MyWellTax LLC',
    notes = 'Updated company name',
    updated_at = NOW()
WHERE tenant_id = 'mywelltax';
```

## Troubleshooting

### Connection Issues

```sql
-- Verify connection details
SELECT
    tenant_id,
    db_host,
    db_port,
    db_user,
    db_name,
    db_sslmode
FROM tenant_connections
WHERE tenant_id = 'mywelltax';
```

### Storage Issues

```sql
-- Verify storage configuration
SELECT
    tenant_id,
    storage_provider,
    storage_bucket,
    storage_credentials_secret,
    storage_credentials_path
FROM tenant_connections
WHERE tenant_id = 'mywelltax';
```

### DocuSign Issues

```sql
-- Verify DocuSign configuration
SELECT
    tenant_id,
    docusign_integration_key,
    docusign_client_id,
    docusign_private_key_secret,
    docusign_api_url
FROM tenant_connections
WHERE tenant_id = 'mywelltax';
```

## Adapter Types

Current supported adapters:

- `mywelltax`: MyWellTax schema (pilot implementation)
- Future: Additional adapters for other tax software schemas

## Migration Notes

When migrating an existing tax platform to WellTaxPro:

1. **Database**: Keep existing database, just add connection details
2. **Storage**: Migrate documents to new bucket OR point to existing bucket
3. **Schema**: Ensure `schema_prefix` matches existing schema name
4. **Adapter**: May require custom adapter development for different schemas

## Complete Setup Script

Save this as `setup_tenant.sql` and run with:
```bash
PGPASSWORD=password psql -h localhost -U postgres -d welltaxpro -f setup_tenant.sql
```

```sql
-- Complete tenant setup in one transaction
BEGIN;

-- Insert tenant
INSERT INTO tenant_connections (
    tenant_id, tenant_name,
    db_host, db_port, db_user, db_password, db_name, db_sslmode, schema_prefix,
    adapter_type, is_active, created_by, notes,
    storage_provider, storage_bucket, storage_credentials_secret, storage_credentials_path,
    docusign_integration_key, docusign_client_id, docusign_private_key_secret, docusign_api_url
) VALUES (
    'mywelltax', 'MyWellTax',
    'localhost', 5432, 'postgres', 'password', 'mywelltax', 'disable', 'taxes',
    'mywelltax', true, 'system', 'Pilot tenant - MyWellTax admin platform',
    'gcs', 'mywelltax-documents-dev',
    'projects/welltaxpro-prod/secrets/mywelltax-gcs-sa/versions/latest',
    '/tmp/mywelltax-sa.json',
    '02662059-8436-404f-80ee-7078d642422b',
    'f1e95a79-5239-4a94-8855-8f69d9f57912',
    '/var/opt/private.key',
    'https://demo.docusign.net/restapi'
)
ON CONFLICT (tenant_id) DO UPDATE SET
    tenant_name = EXCLUDED.tenant_name,
    db_host = EXCLUDED.db_host,
    db_port = EXCLUDED.db_port,
    db_user = EXCLUDED.db_user,
    db_password = EXCLUDED.db_password,
    db_name = EXCLUDED.db_name,
    db_sslmode = EXCLUDED.db_sslmode,
    schema_prefix = EXCLUDED.schema_prefix,
    adapter_type = EXCLUDED.adapter_type,
    is_active = EXCLUDED.is_active,
    notes = EXCLUDED.notes,
    storage_provider = EXCLUDED.storage_provider,
    storage_bucket = EXCLUDED.storage_bucket,
    storage_credentials_secret = EXCLUDED.storage_credentials_secret,
    storage_credentials_path = EXCLUDED.storage_credentials_path,
    docusign_integration_key = EXCLUDED.docusign_integration_key,
    docusign_client_id = EXCLUDED.docusign_client_id,
    docusign_private_key_secret = EXCLUDED.docusign_private_key_secret,
    docusign_api_url = EXCLUDED.docusign_api_url,
    updated_at = NOW();

-- Verify
SELECT
    tenant_id,
    tenant_name,
    is_active,
    storage_bucket,
    docusign_integration_key IS NOT NULL as has_docusign
FROM tenant_connections
WHERE tenant_id = 'mywelltax';

COMMIT;
```

## Additional Configuration

### Portal Configuration

Portal settings are configured in `config.yaml`:

```yaml
portal:
  jwtSecret: "your-secret-key-for-jwt-signing"
  baseURL: "https://your-domain.com"  # Or http://localhost:3000 for dev
```

### SendGrid Configuration

Email settings are configured in `config.yaml`:

```yaml
sendgrid:
  apiKey: "SG.your-api-key"
  defaultFromEmail: "support@welltaxpro.com"
  defaultFromName: "MyWellTax"
```

---

## Summary Checklist

- [ ] Tenant database created and accessible
- [ ] Storage bucket created (GCS/S3/Azure)
- [ ] GCP Service Account created with bucket access
- [ ] Service Account JSON stored in Secret Manager
- [ ] DocuSign Integration Key obtained
- [ ] DocuSign Client ID (User ID) obtained
- [ ] DocuSign Private Key generated and stored
- [ ] Tenant connection added to `tenant_connections` table
- [ ] Database connection tested
- [ ] Storage access tested
- [ ] Portal JWT secret configured in `config.yaml`
- [ ] SendGrid API key configured in `config.yaml`
- [ ] SSL enabled for production databases
- [ ] Strong passwords used for all credentials

---

**Last Updated**: October 15, 2025
**WellTaxPro Version**: v0.1.0
