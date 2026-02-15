# Quick Start Guide

## Start the Application

### Option 1: Development Mode (with hot reload)
```bash
make dev-up          # Start PostgreSQL and Redis
make migrate-up      # Run database migrations
make dev             # Start app with hot reload (Air)
```

### Option 2: Direct Run
```bash
make dev-up          # Start PostgreSQL and Redis
make migrate-up      # Run database migrations
go run ./cmd/app/main.go
```

### Option 3: Docker (Full Stack)
```bash
docker compose up --build
```

## Test the Application

### Run automated tests
```bash
./test-api.sh
```

### Manual API tests

**Health Check:**
```bash
curl http://localhost:8080/health
```

**Upload Sample Collection:**
```bash
curl -X POST http://localhost:8080/api/v1/collections/upload \
  -H "Content-Type: application/json" \
  -d @sample-collection.json
```

**List Collections:**
```bash
curl http://localhost:8080/api/v1/collections
```

**Get Collection Tree (replace :id with actual ID):**
```bash
curl http://localhost:8080/api/v1/collections/1/tree
```

**Execute Request (replace :id with actual item ID):**
```bash
curl -X POST http://localhost:8080/api/v1/items/1/execute
```

## Common Commands

```bash
# Development
make dev              # Run with hot reload
make dev-up           # Start Docker services
make dev-down         # Stop Docker services

# Database
make migrate-up       # Apply migrations
make migrate-down     # Rollback last migration
make migrate-create   # Create new migration
make migrate-status   # Check migration status

# Code Quality
make test             # Run tests
make lint             # Run linter

# Docker
docker compose up -d postgres redis    # Start only DB services
docker compose up --build              # Build and start all
docker compose down                    # Stop all services
docker compose logs -f app             # View app logs
```

## Project Overview

### All TODO Items Implemented ✅

**Backend (Go + PostgreSQL):**
- ✅ Go module initialized
- ✅ HTTP server with Gin framework
- ✅ PostgreSQL connection pool configured
- ✅ Database migrations with Goose
- ✅ Environment configuration (.env support)
- ✅ Collections and collection_items tables
- ✅ Indexes and cascading deletes
- ✅ Health check endpoint

**Postman Import:**
- ✅ JSON upload endpoint
- ✅ Postman v2.0/v2.1 schema validation
- ✅ Unsupported version rejection
- ✅ Collection items validation
- ✅ Folder vs request validation
- ✅ HTTP method validation (GET, POST, PUT, PATCH, DELETE)
- ✅ URL scheme validation (http/https only)
- ✅ Protocol rejection (file://, ftp://, ws://)
- ✅ Request body size limit
- ✅ Header count limit
- ✅ Postman scripts stripped (event, auth, test)
- ✅ Recursive folder/request import
- ✅ Hierarchy preservation
- ✅ Item order preservation (sort_order)
- ✅ Request method, URL, headers (JSONB), body stored

**Collection Retrieval:**
- ✅ Recursive CTE for tree structure
- ✅ Nested JSON tree output
- ✅ Collection list endpoint
- ✅ Item details endpoint

**Request Execution:**
- ✅ Item type validation before execution
- ✅ HTTP request building with headers and body
- ✅ Safe HTTP request execution
- ✅ Request duration measurement
- ✅ Status, headers, body, duration returned

**SSRF Protection:**
- ✅ Private IP range blocking
- ✅ Localhost and metadata IP blocking
- ✅ DNS resolution validation
- ✅ URL scheme validation

**Limits & Safety:**
- ✅ Execution timeout
- ✅ Response size limit
- ✅ Redirect limit with validation
- ✅ Request body size limit
- ✅ Header count limit

**Middleware & Error Handling:**
- ✅ Structured error responses
- ✅ Panic recovery middleware
- ✅ Request logging middleware
- ✅ Rate limiting (in-memory, per-IP)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/collections/upload` | Upload Postman collection |
| GET | `/api/v1/collections` | List all collections |
| GET | `/api/v1/collections/:id/tree` | Get collection tree |
| GET | `/api/v1/items/:id` | Get item details |
| POST | `/api/v1/items/:id/execute` | Execute request (rate limited) |

## Configuration (.env)

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=dev
DB_PASSWORD=localdb
DB_NAME=postman_runner_db
DB_SSLMODE=disable

PORT=8080

REQUEST_TIMEOUT=30s
MAX_REQUEST_SIZE=10485760        # 10MB
MAX_RESPONSE_SIZE=52428800       # 50MB
MAX_HEADER_COUNT=50
MAX_REDIRECTS=5

RATE_LIMIT_RPS=10                # Requests per second
RATE_LIMIT_BURST=20              # Burst capacity
```

## Security Features

### SSRF Protection
- Blocks localhost, private IPs, link-local, metadata endpoints
- Validates URLs before execution
- Only allows http/https schemes

### Rate Limiting
- Per-IP rate limiting (in-memory)
- Applied to execution endpoint
- Configurable RPS and burst

### Input Validation
- Postman schema validation
- HTTP method whitelist
- URL scheme validation
- Size and count limits

## Troubleshooting

### PostgreSQL connection failed
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Start PostgreSQL
docker compose up -d postgres

# Check logs
docker logs postman-runner-postgres
```

### Port 8080 already in use
```bash
# Find process using port 8080
lsof -i :8080

# Change port in .env
PORT=8081
```

### Migrations not applied
```bash
# Check migration status
make migrate-status

# Apply migrations
make migrate-up
```

## Files Structure

```
.
├── cmd/app/main.go                     # Entry point
├── internal/
│   ├── config/config.go                # Config loader
│   ├── db/db.go                        # DB connection
│   ├── handlers/                       # HTTP handlers
│   ├── middleware/                     # Middleware
│   ├── models/models.go                # Data models
│   └── validator/                      # Validation logic
├── migrations/                         # SQL migrations
├── docker-compose.yml                  # Docker services
├── makefile                            # Dev commands
├── .env                                # Config (gitignored)
├── sample-collection.json              # Test collection
├── test-api.sh                         # API test script
└── README.md                           # Full documentation
```

## Next Steps

1. **Test the application**: `./test-api.sh`
2. **Import your own collections**: Use the upload endpoint
3. **Execute requests**: Test your APIs
4. **Customize configuration**: Edit `.env` file
5. **Deploy to production**: Use Docker or build binary

## Support

- Check the full [README.md](README.md) for detailed documentation
- Review [sample-collection.json](sample-collection.json) for collection format
- Run `./test-api.sh` to verify everything works
