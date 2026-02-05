# Go Microservices Justfile
set shell := ["zsh", "-c"]

# Default: list all commands
default:
    @just --list

# --- Infrastructure ---

# Start core services (Postgres, Redis)
infra:
    sudo systemctl start postgresql redis

# Stop core services
infra-stop:
    sudo systemctl stop postgresql redis

# --- User Service ---

# Full setup for the user service
user-setup:
    cd services/user-service && go mod download && swag init
    # Running migrations if you have them
    # migrate -path services/user-service/migrations -database ${DB_URL} up

# Run the user service with hot reload (requires 'air')
user-dev:
    cd services/user-service && air

# --- Global Commands ---

# Run all tests in the monorepo
test-all:
    go test -v ./...

# Tidy all go modules
tidy:
    go work sync || true
    find . -name "go.mod" -execdir go mod tidy \;

# Cleanup binaries
clean:
    rm -rf bin/
    find . -type f -name "user-service" -delete
