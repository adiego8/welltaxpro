-- Rollback initial schema setup
-- Drops all tables in reverse order (respecting foreign key dependencies)

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS tenant_users;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS employee_tenant_access;
DROP TABLE IF EXISTS tenant_connections;
DROP TABLE IF EXISTS employees;
