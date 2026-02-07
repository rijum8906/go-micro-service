set shell := ["zsh", "-c"]
# --- Help ---

# Default: list all available commands [cite: 1]
default:
    @just --list

# --- Service Delegation (The "Router") ---

# Initialize a service (e.g., just setup user-service) [cite: 1]
setup service:
    @just --justfile services/{{service}}/justfile setup

# Run a service in production mode [cite: 1]
run service:
    @just --justfile services/{{service}}/justfile run

# Run a service with hot-reload (Air) [cite: 2]
dev service:
    @just --justfile services/{{service}}/justfile dev

# Apply database migrations for a specific service [cite: 1]
migrate service:
    @just --justfile services/{{service}}/justfile migrate-up

# Create a new migration file (usage: just create-migration user-service add_bio) [cite: 1]
create-migration service name:
    @just --justfile services/{{service}}/justfile create-migration {{name}}

# --- Infrastructure Management ---

# Start local infra for a service (Postgres/Redis via systemd or docker) [cite: 1]
infra-start service:
    @just --justfile services/{{service}}/justfile infra-start

# Stop local infra for a service [cite: 1]
infra-stop service:
    @just --justfile services/{{service}}/justfile infra-stop

# --- Global Monorepo Maintenance ---

# Run all tests across the entire monorepo [cite: 2]
test-all:
    @echo "Running all tests..."
    go test -v ./...

# Synchronize go.work and tidy all go.mod files [cite: 3]
tidy:
    @echo "Tidying workspace..."
    go work sync || true
    find . -name "go.mod" -execdir go mod tidy \;

# Clean all build artifacts and binaries 
clean:
    @echo "Cleaning binaries..."
    rm -rf bin/
    find . -type f -executable -name "*-service" -delete
