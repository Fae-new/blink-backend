# Postman-Compatible API Runner

A robust, production-ready API runner that imports Postman collections and executes HTTP requests with comprehensive security protections.

## Features

✅ **Postman Collection Import**
- Validates Postman v2.0 and v2.1 collection schemas
- Recursively imports folders and requests
- Preserves hierarchy and item order
- Stores requests with full metadata

✅ **Request Execution**
- Execute individual requests by ID
- Measure request duration
- Return full response (status, headers, body)
- SSRF protection (blocks private IPs, localhost, cloud metadata endpoints)
- Configurable timeouts and limits

✅ **Security Features**
- SSRF protection validates all URLs
- Blocks localhost, private IP ranges, and metadata endpoints
- Request/response size limits
- Header count limits
- Redirect limits with validation
- In-memory rate limiting per IP
- Panic recovery middleware

✅ **Database Schema**
- PostgreSQL with proper indexing
- Recursive folder structure support
- Cascading deletes
- JSONB for flexible header storage

## Tech Stack

- **Language**: Go 1.24
- **Framework**: Gin HTTP framework
- **Database**: PostgreSQL 16.6
- **Caching**: Redis (ready for future use)
- **Migrations**: Goose
- **Development**: Air (hot reload), Docker Compose

## Project Structure

```
.
├── cmd/
│   └── app/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration loader
│   ├── db/
│   │   └── db.go                   # Database connection pool
│   ├── handlers/
│   │   ├── health.go               # Health check endpoint
│   │   ├── collection.go           # Collection import
│   │   ├── tree.go                 # Collection tree retrieval
│   │   └── execution.go            # Request execution
│   ├── middleware/
│   │   ├── logger.go               # Request logging
│   │   └── ratelimit.go            # IP-based rate limiting
│   ├── models/
│   │   └── models.go               # Data models
│   └── validator/
│       ├── validator.go            # Postman collection validation
│       └── ssrf.go                 # SSRF protection
├── migrations/
│   ├── 00001_create_collections_table.sql
│   └── 00002_create_collection_items_table.sql
├── docker-compose.yml              # Docker services
├── dockerfile                      # Multi-stage Docker build
├── makefile                        # Development automation
├── .env                            # Environment variables
├── .air.toml                       # Hot reload configuration
├── go.mod
└── go.sum
```

## Getting Started

### Prerequisites

- Go 1.24+
- Docker and Docker Compose
- Make
- Goose (for migrations)

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd "Postman Without bugs"
```

2. **Start services (PostgreSQL + Redis)**
```bash
make dev-up
```

3. **Run migrations**
```bash
make migrate-up
```

4. **Start the development server**
```bash
make dev  # with hot reload
# OR
go run ./cmd/app/main.go
```

The server will start on `http://localhost:8080`

### Using Docker

**Build and run all services:**
```bash
docker compose up --build
```

This will:
- Build the Go application
- Start PostgreSQL
- Start Redis
- Run the application on port 8080

## API Endpoints

### Health Check
```
GET /health
```

### Collections

**Upload Postman Collection**
```
POST /api/v1/collections/upload
Content-Type: application/json

<Postman Collection JSON>
```

**List Collections**
```
GET /api/v1/collections
```

**Get Collection Tree**
```
GET /api/v1/collections/:id/tree
```

### Items

**Get Item Details**
```
GET /api/v1/items/:id
```

**Execute Request**
```
POST /api/v1/items/:id/execute
```
Returns:
```json
{
  "status": 200,
  "headers": { ... },
  "body": "...",
  "duration_ms": 123
}
```

## Configuration

Environment variables (`.env`):

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=dev
DB_PASSWORD=localdb
DB_NAME=postman_runner_db
DB_SSLMODE=disable

# Server
PORT=8080

# Request Execution
REQUEST_TIMEOUT=30s
MAX_REQUEST_SIZE=10485760        # 10MB
MAX_RESPONSE_SIZE=52428800       # 50MB
MAX_HEADER_COUNT=50
MAX_REDIRECTS=5

# Rate Limiting
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20
```

## Security Features

### SSRF Protection

The application blocks:
- Localhost (`127.0.0.1`, `localhost`, `::1`)
- Private IP ranges (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`)
- Link-local addresses (`169.254.0.0/16`, `fe80::/10`)
- Cloud metadata endpoints (`169.254.169.254`)
- Non-HTTP(S) schemes (`file://`, `ftp://`, `ws://`)

### Rate Limiting

- Per-IP rate limiting
- Configurable requests per second (RPS)
- Configurable burst size
- Applied to execution endpoint

### Input Validation

- Postman schema validation (v2.0, v2.1 only)
- HTTP method validation (GET, POST, PUT, PATCH, DELETE only)
- URL scheme validation
- Header count limits
- Request/response size limits

## Development

### Makefile Commands

```bash
make dev              # Run with hot reload (Air)
make dev-up           # Start Docker services
make dev-down         # Stop Docker services
make migrate-up       # Run migrations
make migrate-down     # Rollback migrations
make migrate-create   # Create new migration
make migrate-status   # Check migration status
make test             # Run tests
make lint             # Run linter
```

### Database Migrations

**Create a new migration:**
```bash
make migrate-create
# Enter migration name when prompted
```

**Apply migrations:**
```bash
make migrate-up
```

**Rollback last migration:**
```bash
make migrate-down
```

## Database Schema

### Collections Table
```sql
CREATE TABLE collections (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Collection Items Table
```sql
CREATE TABLE collection_items (
    id SERIAL PRIMARY KEY,
    collection_id INTEGER NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES collection_items(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    item_type VARCHAR(20) NOT NULL CHECK (item_type IN ('folder', 'request')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    method VARCHAR(10) CHECK (method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE')),
    url TEXT,
    headers JSONB,
    body TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Example: Importing a Postman Collection

```bash
curl -X POST http://localhost:8080/api/v1/collections/upload \
  -H "Content-Type: application/json" \
  -d @my-collection.json
```

Response:
```json
{
  "collection_id": 1,
  "message": "Collection imported successfully"
}
```

## Example: Executing a Request

```bash
curl -X POST http://localhost:8080/api/v1/items/1/execute
```

Response:
```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json",
    "Content-Length": "123"
  },
  "body": "{ ... }",
  "duration_ms": 245
}
```

## Production Deployment

1. **Set environment variables for production**
2. **Use production database credentials**
3. **Set `GIN_MODE=release`**
4. **Configure proper rate limits**
5. **Use HTTPS reverse proxy (nginx/traefik)**
6. **Monitor logs and metrics**

## TODO/Future Enhancements

- [ ] Authentication & authorization
- [ ] Collection sharing & permissions
- [ ] Scheduled request execution
- [ ] Request history & analytics
- [ ] Variable/environment support
- [ ] Pre-request scripts (safe subset)
- [ ] Response assertions/tests
- [ ] Webhook notifications
- [ ] Export to other formats

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.
