.PHONY: build run provision clean test migrate-up migrate-down migrate-create migrate-version migrate-force

GO := go

# Database configuration (can be overridden via environment variables)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= password
DB_NAME ?= welltaxpro
DB_SSLMODE ?= disable

# Migration configuration
MIGRATE_IMAGE := migrate/migrate:v4.17.0
MIGRATION_DIR := $(shell pwd)/migrations
DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Build the server binary
build:
	$(GO) build -o bin/welltaxpro ./src/cmd/main

# Build the provisioner
build-provisioner:
	$(GO) build -o bin/provisioner ./src/cmd/provisioner

# Run the server
run: build
	./bin/welltaxpro --config config/environment/dev-config.yaml

# Run database migrations
provision: build-provisioner
	./bin/provisioner --config config/environment/dev-config.yaml

# Clean build artifact welltaxpro
clean-welltaxpro:
	rm -f bin/welltaxpro

# Run tests
test:
	$(GO) test ./...

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Development: run with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Format code
fmt:
	$(GO) fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# ============================================================================
# Database Migration Targets (using go-migrate via Docker)
# ============================================================================

# Apply all pending migrations
migrate-up:
	@echo "Creating database if it doesn't exist..."
	@docker run --rm --network host postgres:14 \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=$(DB_SSLMODE)" \
		-c "SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || \
		docker run --rm --network host postgres:14 \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=$(DB_SSLMODE)" \
		-c "CREATE DATABASE $(DB_NAME)"
	@echo "Running migrations..."
	docker run --rm -v $(MIGRATION_DIR):/migrations --network host \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$(DATABASE_URL)" \
		up

# Rollback migrations (usage: make migrate-down STEPS=1)
migrate-down:
	docker run --rm -v $(MIGRATION_DIR):/migrations --network host \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$(DATABASE_URL)" \
		down $(if $(STEPS),$(STEPS),1)

# Create a new migration file (usage: make migrate-create NAME=create_users_table)
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME parameter required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	docker run --rm -v $(MIGRATION_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		create -ext sql -dir /migrations -seq $(NAME)

# Show current migration version
migrate-version:
	docker run --rm -v $(MIGRATION_DIR):/migrations --network host \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$(DATABASE_URL)" \
		version

# Force migration version (usage: make migrate-force VERSION=20231215120000)
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION parameter required. Usage: make migrate-force VERSION=version_number"; \
		exit 1; \
	fi
	docker run --rm -v $(MIGRATION_DIR):/migrations --network host \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$(DATABASE_URL)" \
		force $(VERSION)
