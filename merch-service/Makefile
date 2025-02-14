include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DOCKER
# ==================================================================================== #

## up: Build and start the service container along with dependencies
.PHONY: up
up: postgres/ready $(SERVICE_NAME)/start
	@echo "$(SERVICE_NAME) is up and running."

## $(SERVICE_NAME)/start: Start the service container
.PHONY: $(SERVICE_NAME)/start
$(SERVICE_NAME)/start:
	@docker-compose up -d $(SERVICE_NAME)
	@echo "$(SERVICE_NAME) is started."

## postgres/ready: Wait until PostgreSQL is ready
.PHONY: postgres/ready
postgres/ready: postgres/start
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker-compose exec postgres pg_isready -U $$POSTGRES_USER -d $$POSTGRESQL_DB; do \
		echo "Waiting for PostgreSQL database $(POSTGRES_DB)..."; \
		sleep 2; \
	done
	@echo "PostgreSQL database $(POSTGRES_DB) is ready!"
	
## postgres/start: Start the PostgreSQL container if not running
.PHONY: postgres/start
postgres/start:
	@docker-compose up -d postgres
	@echo "Starting PostgreSQL container..."

# ==================================================================================== #
# MIGRATION TASKS
# ==================================================================================== #

## migrate/up: Apply all migrations to the database
.PHONY: migrate/up
migrate/up:
	@echo "Applying migrations for $(SERVICE_NAME)..."
	@migrate -path ./migrations -database postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE) up
	@echo "Migrations applied for $(SERVICE_NAME)."

## migrate/down: Rollback the last migration
.PHONY: migrate/down
migrate/down:
	@echo "Rolling back the last migration for $(SERVICE_NAME)..."
	@migrate -path ./migrations -database postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE) down
	@echo "Last migration rolled back for $(SERVICE_NAME)."

## migrate/new name=$(name): Create a new migration
.PHONY: migrate/new
migrate/new:
	@echo "Creating new migration file for $(name)..."
	@migrate create -seq -ext sql -dir ./migrations $(name)
	@echo "New migration file created in ./migrations/"

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## gen: generates tests mocks
.PHONY: gen 
gen: 
	mockgen -source=internal/merch/pg_repository.go -destination=internal/merch/mock/pg_repository_mock.go
	mockgen -source=internal/auth/pg_repository.go -destination=internal/auth/mock/pg_repository_mock.go

## coverage: check tests coverage
.PHONY: coverage
coverage:
	go test -coverprofile=./docs/coverage.out ./...
	go tool cover -func=./docs/coverage.out
	go tool cover -html=./docs/coverage.out

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify

	@echo 'Formatting code...'
	go fmt ./...
	swag fmt ./...

	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...

	@echo 'Running tests...'
	go test -race -vet=off ./...