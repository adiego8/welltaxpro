-- Initial schema setup for WellTaxPro
-- This migration creates all core tables with proper constraints and indexes

-- ============================================================================
-- Employees Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) NOT NULL DEFAULT 'accountant',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_employee_role CHECK (role IN ('admin', 'accountant', 'viewer'))
);

CREATE INDEX idx_employees_firebase_uid ON employees(firebase_uid);
CREATE INDEX idx_employees_email ON employees(email);
CREATE INDEX idx_employees_active ON employees(is_active, role);

COMMENT ON TABLE employees IS 'WellTaxPro employees who can access multiple tenants';
COMMENT ON COLUMN employees.role IS 'Global role: admin, accountant, viewer';

-- ============================================================================
-- Tenant Connections Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS tenant_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(100) NOT NULL UNIQUE,
    tenant_name VARCHAR(255) NOT NULL,
    db_host VARCHAR(255) NOT NULL,
    db_port INTEGER NOT NULL DEFAULT 5432,
    db_user VARCHAR(100) NOT NULL,
    db_password TEXT NOT NULL,
    db_name VARCHAR(100) NOT NULL,
    db_sslmode VARCHAR(20) NOT NULL DEFAULT 'require',
    schema_prefix VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    notes TEXT,
    adapter_type VARCHAR(50) NOT NULL DEFAULT 'mywelltax',
    storage_provider VARCHAR(20),
    storage_bucket VARCHAR(255),
    storage_credentials_secret VARCHAR(500),
    storage_credentials_path VARCHAR(500),
    docusign_integration_key VARCHAR(255),
    docusign_client_id VARCHAR(255),
    docusign_private_key_secret VARCHAR(500),
    docusign_api_url VARCHAR(255) DEFAULT 'https://demo.docusign.net/restapi',

    CONSTRAINT chk_adapter_type CHECK (adapter_type IN ('mywelltax', 'drake', 'lacerte', 'proseries', 'ultratax')),
    CONSTRAINT chk_storage_provider CHECK (storage_provider IS NULL OR storage_provider IN ('gcs', 's3', 'azure')),
    CONSTRAINT chk_db_sslmode CHECK (db_sslmode IN ('disable', 'require', 'verify-ca', 'verify-full'))
);

CREATE INDEX idx_tenant_id ON tenant_connections(tenant_id);
CREATE INDEX idx_is_active ON tenant_connections(is_active);
CREATE INDEX idx_adapter_type ON tenant_connections(adapter_type);
CREATE INDEX idx_tenant_connections_storage_provider ON tenant_connections(storage_provider);

COMMENT ON TABLE tenant_connections IS 'Registry of tenant database connections for multi-tenant access';
COMMENT ON COLUMN tenant_connections.tenant_id IS 'Unique identifier for tenant (used in API routes)';
COMMENT ON COLUMN tenant_connections.db_password IS 'Database password - should be encrypted at rest in production';
COMMENT ON COLUMN tenant_connections.schema_prefix IS 'Schema or table prefix used in tenant database';
COMMENT ON COLUMN tenant_connections.adapter_type IS 'Specifies which adapter implementation to use for this tenant';
COMMENT ON COLUMN tenant_connections.storage_provider IS 'Cloud storage provider: gcs, s3, or azure';
COMMENT ON COLUMN tenant_connections.storage_bucket IS 'Bucket or container name for document storage';
COMMENT ON COLUMN tenant_connections.storage_credentials_secret IS 'GCP Secret Manager path (e.g., projects/PROJECT/secrets/NAME/versions/VERSION)';
COMMENT ON COLUMN tenant_connections.storage_credentials_path IS 'Fallback: Path to service account JSON file (for local dev)';
COMMENT ON COLUMN tenant_connections.docusign_integration_key IS 'DocuSign Integration Key (from DocuSign app)';
COMMENT ON COLUMN tenant_connections.docusign_client_id IS 'DocuSign Client ID / User ID for JWT auth';
COMMENT ON COLUMN tenant_connections.docusign_private_key_secret IS 'GCP Secret Manager path to DocuSign RSA private key';
COMMENT ON COLUMN tenant_connections.docusign_api_url IS 'DocuSign API base URL (demo or production)';

-- ============================================================================
-- Employee Tenant Access Table (Multi-tenant access control)
-- ============================================================================
CREATE TABLE IF NOT EXISTS employee_tenant_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID,

    CONSTRAINT fk_employee FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_connections(tenant_id) ON DELETE CASCADE,
    CONSTRAINT chk_tenant_role CHECK (role IN ('admin', 'accountant', 'viewer')),
    CONSTRAINT uq_employee_tenant UNIQUE (employee_id, tenant_id)
);

CREATE INDEX idx_employee_tenant_employee ON employee_tenant_access(employee_id);
CREATE INDEX idx_employee_tenant_tenant ON employee_tenant_access(tenant_id);
CREATE INDEX idx_employee_tenant_active ON employee_tenant_access(is_active);

COMMENT ON TABLE employee_tenant_access IS 'Controls which employees can access which tenants and their role within each tenant';
COMMENT ON COLUMN employee_tenant_access.role IS 'Role within this specific tenant: admin, accountant, viewer';
COMMENT ON COLUMN employee_tenant_access.is_active IS 'Whether this employee can currently access this tenant';

-- ============================================================================
-- Audit Logs Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    client_id UUID,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_audit_employee FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    CONSTRAINT fk_audit_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_connections(tenant_id) ON DELETE CASCADE
);

CREATE INDEX idx_audit_timestamp ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_employee_time ON audit_logs(employee_id, created_at DESC);
CREATE INDEX idx_audit_tenant_time ON audit_logs(tenant_id, created_at DESC);
CREATE INDEX idx_audit_client_time ON audit_logs(client_id, created_at DESC);
CREATE INDEX idx_audit_action_time ON audit_logs(action, created_at DESC);
CREATE INDEX idx_audit_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_employee_client_time ON audit_logs(employee_id, client_id, created_at DESC);

COMMENT ON TABLE audit_logs IS 'Audit trail of all employee actions for IRS compliance';
COMMENT ON COLUMN audit_logs.action IS 'Action performed: view, create, update, delete, download, etc.';
COMMENT ON COLUMN audit_logs.resource_type IS 'Type of resource: client, document, filing, etc.';

-- ============================================================================
-- Portal Magic Tokens Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS portal_magic_tokens (
    id UUID PRIMARY KEY,
    client_id UUID NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    email TEXT NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_magic_token_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_connections(tenant_id) ON DELETE CASCADE
);

CREATE INDEX idx_magic_tokens_client ON portal_magic_tokens(client_id);
CREATE INDEX idx_magic_tokens_expiry ON portal_magic_tokens(expires_at);
CREATE INDEX idx_magic_tokens_used ON portal_magic_tokens(used);
CREATE INDEX idx_magic_tokens_tenant ON portal_magic_tokens(tenant_id);

COMMENT ON TABLE portal_magic_tokens IS 'Tracks one-time use magic link tokens with 24-hour expiry for portal access';
COMMENT ON COLUMN portal_magic_tokens.id IS 'JWT ID (jti claim) for token tracking';
COMMENT ON COLUMN portal_magic_tokens.used IS 'Marks if token has been exchanged for session token';

-- ============================================================================
-- Cleanup Function for Expired Magic Tokens
-- ============================================================================
CREATE OR REPLACE FUNCTION cleanup_expired_magic_tokens() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM portal_magic_tokens
    WHERE expires_at < NOW() - INTERVAL '7 days';
END;
$$;

COMMENT ON FUNCTION cleanup_expired_magic_tokens IS 'Removes expired magic tokens older than 7 days (run via cron)';
