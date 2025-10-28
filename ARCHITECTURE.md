# WellTaxPro Architecture

## Overview

WellTaxPro is a multi-tenant, provider-agnostic admin platform for tax accounting firms. The system reads data from tenant databases using a pluggable adapter pattern.

## Key Architectural Decisions

### 1. Multi-Tenant Database Architecture

**Pattern**: Database-per-tenant with connection registry

```
┌─────────────────────────────────────┐
│   WellTaxPro Database (Metadata)    │
│   - tenant_connections table        │
│   - stores DB credentials           │
│   - stores adapter type             │
└──────────────┬──────────────────────┘
               │
         ┌─────┴──────┬──────────┐
         │            │          │
    ┌────▼───┐   ┌───▼────┐  ┌──▼────┐
    │MyWell  │   │TenantB │  │TenantC│
    │Tax DB  │   │DB      │  │DB     │
    └────────┘   └────────┘  └───────┘
```

**Why this approach:**
- Strong data isolation (IRS compliance)
- Each tenant can have different schemas
- Individual backup/restore per tenant
- Scales horizontally by tenant

### 2. Adapter Pattern for Schema Abstraction

**Problem**: Different tax platforms have different database schemas.

**Solution**: Adapter interface per platform.

```go
type ClientAdapter interface {
    GetClients(db *sql.DB, schemaPrefix string) ([]*types.Client, error)
    GetClientByID(db *sql.DB, schemaPrefix string, clientID string) (*types.Client, error)
}
```

**Implemented Adapters:**
- `MyWellTaxAdapter` - taxes.user table schema

**Future Adapters:**
- `DrakeAdapter` - for Drake Tax software
- `LacerteAdapter` - for Lacerte
- `CustomAdapter` - for firms with custom systems

**Flow:**
1. Request comes in for `GET /api/v1/mywelltax/clients`
2. Store looks up `mywelltax` tenant → gets `adapter_type: "mywelltax"`
3. Creates `MyWellTaxAdapter`
4. Adapter executes MyWellTax-specific SQL queries
5. Returns standardized `Client` objects

### 3. Connection Pool Management with TTL

**Problem**: Storing connections for 100 tenants = 100 open pools = resource exhaustion

**Solution**: TTL-based eviction

```go
type tenantConnection struct {
    db         *sql.DB
    lastAccess time.Time
}
```

**Eviction Strategy:**
- Background goroutine checks every 1 minute
- Closes connections idle for > 5 minutes
- Connections re-established on next request

**Benefits:**
- Prevents memory leaks
- Optimizes resource usage
- Scales to hundreds of tenants

## Directory Structure

```
backend/
├── src/
│   ├── cmd/
│   │   ├── main/           # Server entry point
│   │   ├── server/         # HTTP server logic
│   │   └── provisioner/    # Database migrations
│   ├── internal/
│   │   ├── adapter/        # Tenant-specific adapters
│   │   │   ├── adapter.go        # Interface & factory
│   │   │   └── mywelltax.go      # MyWellTax implementation
│   │   ├── store/          # Database access layer
│   │   │   ├── store.go          # Connection management
│   │   │   ├── tenant.go         # Tenant connection lookup
│   │   │   └── client.go         # Client queries (uses adapters)
│   │   └── types/          # Domain models
│   │       ├── tenant.go         # TenantConnection
│   │       └── client.go         # Client
│   └── api/
│       └── web/            # REST API handlers
│           ├── webapi.go         # Router setup
│           └── clients.go        # Client endpoints
```

## API Endpoints

### Get Clients
```
GET /api/v1/{tenantId}/clients
```

Returns all clients for a tenant using the tenant's configured adapter.

### Get Client by ID
```
GET /api/v1/{tenantId}/clients/{clientId}
```

Returns a specific client.

## Database Schema

### tenant_connections (WellTaxPro DB)

```sql
CREATE TABLE tenant_connections (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(100) UNIQUE NOT NULL,
    tenant_name VARCHAR(255) NOT NULL,

    -- Database connection
    db_host VARCHAR(255) NOT NULL,
    db_port INTEGER NOT NULL,
    db_user VARCHAR(100) NOT NULL,
    db_password TEXT NOT NULL,
    db_name VARCHAR(100) NOT NULL,
    db_sslmode VARCHAR(20) NOT NULL,

    -- Adapter configuration
    schema_prefix VARCHAR(100) NOT NULL,
    adapter_type VARCHAR(50) NOT NULL DEFAULT 'mywelltax',

    -- Metadata
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Adding a New Tenant

1. **Insert tenant connection**:
```sql
INSERT INTO tenant_connections (
    tenant_id, tenant_name, db_host, db_port, db_user, db_password,
    db_name, db_sslmode, schema_prefix, adapter_type
) VALUES (
    'newclient', 'New Client Inc', 'localhost', 5432, 'readonly_user', 'password',
    'newclient_db', 'require', 'public', 'mywelltax'
);
```

2. **API automatically works**:
```bash
curl http://localhost:8081/api/v1/newclient/clients
```

## Adding a New Adapter

1. **Create adapter file**: `src/internal/adapter/drake.go`

```go
type DrakeAdapter struct{}

func (a *DrakeAdapter) GetAdapterType() string {
    return "drake"
}

func (a *DrakeAdapter) GetClients(db *sql.DB, schemaPrefix string) ([]*types.Client, error) {
    // Drake-specific SQL
    query := fmt.Sprintf(`
        SELECT client_id, fname, lname, email_address
        FROM %s.clients  -- Drake uses "clients" table
        WHERE is_active = 't'
    `, schemaPrefix)

    // Map Drake columns to WellTaxPro Client type
    // ...
}
```

2. **Register in factory**: `src/internal/adapter/adapter.go`

```go
func NewAdapter(adapterType string) (ClientAdapter, error) {
    switch adapterType {
    case "mywelltax":
        return &MyWellTaxAdapter{}, nil
    case "drake":
        return &DrakeAdapter{}, nil  // Add this
    default:
        return nil, fmt.Errorf("unknown adapter: %s", adapterType)
    }
}
```

3. **Update tenant config**:
```sql
UPDATE tenant_connections
SET adapter_type = 'drake'
WHERE tenant_id = 'drakeclient';
```

## Security Considerations (TODO)

⚠️ **Production blockers:**

1. **Encrypt DB passwords**: Currently plain text in `tenant_connections`
   - Use AWS Secrets Manager, HashiCorp Vault, or encrypted columns

2. **Add authentication**: No auth layer currently
   - Implement JWT/OAuth for API access
   - Add tenant-to-user mapping

3. **Add authorization**: Anyone can access any tenant
   - Implement role-based access control (RBAC)
   - Validate user has access to requested tenant

4. **Rate limiting**: Prevent abuse
   - Per-tenant rate limits
   - Global API rate limits

## Performance Optimizations

- **Connection pooling**: Max 5 connections per tenant DB
- **TTL eviction**: 5-minute idle timeout
- **Lazy loading**: Connections created on first request
- **Read-only mode**: Prevents accidental data modification

## Next Steps

1. Add authentication middleware
2. Encrypt tenant credentials
3. Implement authorization layer
4. Add audit logging
5. Create second adapter (Drake or Lacerte)
6. Add tenant health monitoring
