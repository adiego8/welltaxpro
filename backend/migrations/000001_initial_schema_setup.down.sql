-- Rollback initial schema setup
-- Drops all tables in reverse order (respecting foreign key dependencies)

-- Drop function first
DROP FUNCTION IF EXISTS cleanup_expired_magic_tokens();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS portal_magic_tokens;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS employee_tenant_access;
DROP TABLE IF EXISTS tenant_connections;
DROP TABLE IF EXISTS employees;
