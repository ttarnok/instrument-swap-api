
include .envrc
# ============================================================================ #
# HELPERS
# ============================================================================ #

## help: prints makefile targets and their usage
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## code/fmt: formats the source code
.PHONY: code/fmt
code/fmt:
	@echo 'Formatting source code...'
	@go fmt ./...

# ============================================================================ #
# DEVELOPMENT
# ============================================================================ #

## build/api: builds the api binary from the source code
.PHONY: build/api
build/api: code/fmt
	@echo 'Building the api binary...'
	@go build -ldflags="-s -w" -o=./bin/api ./cmd/api

## run/api: runs the cmd/api application
.PHONY: run/api
run/api: build/api
	@echo 'Running the application...'
	@./bin/api -port=4000 -db-dsn=$(INSTRUMENT_SWAP_DB_DSN) -jwt-secret=$(JWT_SECRET)

## docker/compose/up: runs docker compose up for the local dev db
.PHONY: docker/compose/up
docker/compose/up:
	docker-compose up --build -d

## docker/compose/down: runs docker compose down for the local dev db
.PHONY: docker/compose/down
docker/compose/down:
	docker-compose down

## db/migrations/new name=fizz: creates a pair of new migration files with the name of fizz
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all available database migrations
.PHONY: db/migrations/up
db/migrations/up:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) up

## db/migrations/down: downgrade all available database migrations
.PHONY: db/migrations/down
db/migrations/down:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) down

## db/migrations/version: lists the actual applied db migrations version
.PHONY: db/migrations/version
db/migrations/version:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) version

##db/migrations/force version=number: forcefully downmigrates to the specified version
.PHONY: db/migrations/force
db/migrations/force:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) force ${version}



# ============================================================================ #
# QUALITY CONTROL
# ============================================================================ #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	@echo 'staticcheck...'
	staticcheck ./...
	@echo 'golangci-lint...'
	golangci-lint run
	@echo 'Running tests...'
	go test -v -count=1 -race -vet=off ./...

## cover: generate test coverage report
.PHONY: cover
cover:
	go test -covermode=count -coverprofile=profile.out ./...
	go tool cover -html=profile.out
