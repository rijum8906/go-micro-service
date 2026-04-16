# User Service

The User Service is the identity and session management service for the Relay
backend. It exposes gRPC endpoints for authentication, user profile operations,
and session lifecycle management, backed by PostgreSQL for persistence and Redis
for token and cache workloads.

This README focuses on how to run and work on the service. Detailed system
design and contract documentation should live in:

- `ARCHITECTURE.md`
- `API.md`

## Responsibilities

This service owns:

- user registration and login
- access token and scoped token issuance
- session creation, refresh, and revocation
- user and profile data access
- auth-related persistence and validation flows

## Stack

- Go `1.26`
- gRPC
- PostgreSQL
- Redis
- `sqlc` for typed query generation
- Atlas for schema and migration management

## Repository Context

This service is part of a larger monorepo and depends on shared local modules:

- `../../packages/core`
- `../../packages/pb`

Run commands from this service directory unless noted otherwise. Docker builds
use the monorepo root as the build context.

## Project Layout

```text
.
├── app/                  # Application bootstrap and dependency wiring
├── cmd/main.go           # Service entrypoint
├── cmd/migrate/          # Programmatic migration runner
├── cmd/test-cli/         # Local test environment helper
├── db/                   # Schema, queries, migrations
├── internal/db/          # Generated sqlc code
├── internal/handlers/    # gRPC handlers
├── internal/repository/  # Repository implementations
├── internal/services/    # Business logic
└── internal/dto/         # Internal transport/data objects
```

## Prerequisites

Install or have access to:

- Go `1.26+`
- PostgreSQL
- Redis
- Docker, if you want to use the local test database flow
- Atlas CLI
- `sqlc`

You can install the schema tools with:

```bash
make setup
```

## Configuration

The service reads configuration from environment variables and will also load a
local `.env` file if present.

Core variables:

| Variable | Required | Description |
| --- | --- | --- |
| `APP_NAME` | Yes | Service/application name |
| `APP_ENV` | No | Environment name, defaults to `development` |
| `PORT` | No | gRPC listen port, defaults to `8080` |
| `DB_HOST` | Yes | PostgreSQL host |
| `DB_PORT` | No | PostgreSQL port, defaults to `5432` |
| `DB_USER` | Yes | PostgreSQL user |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DB_NAME` | Yes | PostgreSQL database name |
| `DB_SSL_MODE` | No | PostgreSQL SSL mode, defaults to `disable` |
| `REDIS_HOST` | Yes | Redis host |
| `REDIS_PORT` | No | Redis port, defaults to `6379` |
| `REDIS_PASS` | No | Redis password |
| `JWT_SECRET` | Yes | Secret used for auth token signing |
| `SCOPED_SECRET` | Yes | Secret used for scoped token signing |
| `SESSION_TTL` | No | Session token TTL, defaults to `15m` |
| `SCOPED_TOKEN_TTL` | No | Scoped token TTL, defaults to `10m` |
| `REFRESH_TOKEN_TTL` | No | Refresh token TTL, defaults to `600h` |

Minimal local example:

```env
APP_NAME=user-service
APP_ENV=development
PORT=8906

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=postgres
DB_SSL_MODE=disable

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=

JWT_SECRET=change-me
SCOPED_SECRET=change-me-too
SESSION_TTL=15m
SCOPED_TOKEN_TTL=10m
REFRESH_TOKEN_TTL=600h
```

## Local Use

For quick local setup and usage without requiring Atlas version tracking:

1. Start PostgreSQL and Redis.
2. Configure the required environment variables in `.env`.
3. Apply database migrations using embedded SQL:

   ```bash
   make db-sql-apply
   ```

4. Run the service:

   ```bash
   make run
   ```

This flow applies the SQL files in `db/migrations/` directly to your database.

## Development

For active development, schema changes, and migration management, use the Atlas-based workflow:

### Prerequisites

Ensure you have [Atlas](https://atlasgo.io/) installed.

### Workflow

1. **Initialize Database**: Create the main and development databases if they don't exist.

   ```bash
   make db-init
   ```

2. **Apply Migrations**: Sync your database with the current migration set.

   ```bash
   make db-apply
   ```

3. **Generate New Migration**: After changing `db/schema.sql`, generate a new migration file.

   ```bash
   make db-new NAME=your_change_name
   ```

4. **Other Commands**:
   - `make db-status`: Check migration status.
   - `make db-rollback`: Roll back the last migration.
   - `make db-reset`: Wipe the database and re-apply all migrations.

## Testing

Run the unit test suite:

```bash
make test
```

For the local integration-style flow:

```bash
make test-setup
make test-run
```

`make test-setup` uses the embedded SQL loader to apply migrations directly to the test database.

## Available Commands

| Command | Description | Flow |
| --- | --- | --- |
| `make run` | Run the user service | Both |
| `make db-sql-apply` | Apply embedded SQL migrations | **Local Use** |
| `make db-apply` | Apply Atlas-managed migrations | **Development** |
| `make db-new NAME=xxx` | Generate a new migration | **Development** |
| `make db-init` | Create necessary databases | **Development** |
| `make test` | Run all Go tests | Both |
| `make test-setup` | Prepare test environment | Both |

## Docker

The provided `Dockerfile` is intended to be built from the monorepo root so the
shared `packages/` modules are available in the build context.

Example:

```bash
docker build -f services/user-service/Dockerfile .
```

## Related Documentation

- `ARCHITECTURE.md` for service boundaries, dependencies, and internal design
- `API.md` for protobuf contracts, RPC behavior, and request/response details

## Notes For Contributors

- Keep business logic in `internal/services/`
- Keep transport concerns in `internal/handlers/`
- Keep persistence concerns in `internal/repository/`
- Regenerate and review query code when SQL changes
