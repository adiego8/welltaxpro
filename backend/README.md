# WellTaxPro Backend

Multi-tenant admin platform for tax accounting firms.

## Quick Start

1. **Install dependencies**
```bash
make deps
```

2. **Configure database**
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your database credentials
```

3. **Run database migrations**
```bash
make provision
```

4. **Insert first tenant (MyWellTax)**
```sql
INSERT INTO tenant_connections (
    tenant_id, tenant_name, db_host, db_port, db_user, db_password,
    db_name, db_sslmode, schema_prefix, created_by, notes
) VALUES (
    'mywelltax', 'MyWellTax', 'localhost', 5432, 'readonly_user', 'PASSWORD',
    'mywelltax', 'require', 'taxes', 'system', 'Pilot tenant'
);
```

5. **Run server**
```bash
make run
```

## API Endpoints

### Get Clients
```
GET /api/v1/{tenantId}/clients
```

### Get Client by ID
```
GET /api/v1/{tenantId}/clients/{clientId}
```

### Health Check
```
GET /health
```

## Architecture

```
welltaxpro (database)
└── tenant_connections (table)
    ├── Connection info for each tenant

For each tenant request:
1. Lookup connection details from tenant_connections
2. Connect to tenant's database
3. Query tenant's data (e.g., taxes.user table)
```

## Development

- `make build` - Build server binary
- `make run` - Build and run server
- `make provision` - Run database migrations
- `make test` - Run tests
- `make fmt` - Format code
