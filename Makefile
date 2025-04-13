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
	@docker-compose up -d $(SERVICE_NAME) swagger-ui
	@echo "$(SERVICE_NAME) is started."

## postgres/ready: Wait until PostgreSQL is ready
.PHONY: postgres/ready
postgres/ready: postgres/start
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker-compose exec -T pvz-postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB); do \
		echo "Waiting for PostgreSQL database $(POSTGRES_DB)..."; \
		sleep 2; \
	done
	@echo "PostgreSQL database $(POSTGRES_DB) is ready!"
	
## postgres/start: Start the PostgreSQL container if not running
.PHONY: postgres/start
postgres/start:
	@docker-compose up -d pvz-postgres
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

## gen/api: generate api
.PHONY: gen/api
gen/api: 
	oapi-codegen -config gen/oapi-codegen.yaml gen/swagger.yaml

## gen/proto: generate protos
.PHONY: gen/proto
gen/proto: 
	protoc --proto_path=protos \
		--go_out=./protos/gen --go_opt=paths=source_relative \
		--go-grpc_out=./protos/gen --go-grpc_opt=paths=source_relative \
		protos/proto/pvz/pvz.proto

## gen/mock: generate tests mocks
.PHONY: gen/mock
gen/mock: 
	mockgen -source=internal/pvz/pg_repository.go -destination=internal/pvz/mock/pg_repository_mock.go

## coverage: check tests coverage
.PHONY: coverage
coverage:
	go test -coverprofile=./gen/test/coverage.out ./...
	go tool cover -func=./gen/test/coverage.out 
	go tool cover -html=./gen/test/coverage.out

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify

	@echo 'Running linter...'
	golangci-lint run --config .golangci.yml
	
	@echo 'Running tests...'
	go test -race -vet=off ./...