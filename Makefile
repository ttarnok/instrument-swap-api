
-include .env
# ============================================================================ #
# HELPERS
# ============================================================================ #

## help: prints makefile targets and their usage
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ============================================================================ #
# DEVELOPMENT - LOCAL DEV
# ============================================================================ #

## code/fmt: formats the source code
.PHONY: code/fmt
code/fmt:
	@echo 'Formatting source code...'
	@go fmt ./...

## build/api: formats and builds the api binary locally
.PHONY: build/api
build/api: code/fmt
	@echo 'Building the api binary...'
	@go build -ldflags="-s -w" -o=./bin/api ./cmd/api

## run/api: formats, builds and runs the cmd/api application locally
## : depends on the following environment variables:
## : - INSTRUMENT_SWAP_DB_DSN -> dsn of the used Postgres database
## : - JWT_SECRET -> secret for the JWTs
## : - REDIS_ADDR -> dsn for Redis
## : - REDIS_PASSWORD -> Redis password
## : - REDIS_DB -> Redis database index number
.PHONY: run/api
run/api: build/api
	@echo 'Running the application...'
	@./bin/api -db-dsn=$(INSTRUMENT_SWAP_DB_DSN) -jwt-secret=$(JWT_SECRET) -redis-address=${REDIS_ADDR} -redis-password=${REDIS_PASSWORD} -redis-db=${REDIS_DB}

# ============================================================================ #
# DEVELOPMENT - DOCKER
# ============================================================================ #

## docker/compose/up: runs docker compose up for the local dev environment
## : (Postgres database with migrations, Redis, application binary)
.PHONY: docker/compose/up
docker/compose/up:
	docker-compose up --build -d

## docker/compose/down: runs docker compose down for the local dev environment
.PHONY: docker/compose/down
docker/compose/down:
	docker-compose down

## docker/build/app: builds the docker image of the dockerized application
.PHONY: docker/build/app
docker/build/app:
	docker build -t instrument-swap-api:test .

## docker/run/app: runs the built docker image of the dockerized application
## : depends on the following environment variables:
## : - INSTRUMENT_SWAP_DB_DSN -> dsn of the used Postgres database
## : - JWT_SECRET -> secret for the JWTs
## : - REDIS_ADDR -> dsn for Redis
## : - REDIS_PASSWORD -> Redis password
## : - REDIS_DB -> Redis database index number
.PHONY: docker/run/app
docker/run/app:
	docker run instrument-swap-api:test \
           -db-dsn=${INSTRUMENT_SWAP_DB_DSN} \
           -jwt-secret=${JWT_SECRET} \
           -redis-address=${REDIS_ADDR} \
           -redis-password=${REDIS_PASSWORD} \
           -redis-db=${REDIS_DB}

## docker/logs: Fethes the application related logs from docker
.PHONY: docker/logs
docker/logs:
	docker logs -f instrument-swap-api

# ============================================================================ #
# DEVELOPMENT - DATABASE MIGRATION
# ============================================================================ #

## db/migrations/new name=fizz: creates a pair of new migration files with the name of fizz
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all available database migrations
## : depends on the following environment variables:
## : - INSTRUMENT_SWAP_DB_DSN -> dsn of the used Postgres database
.PHONY: db/migrations/up
db/migrations/up:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) up

## db/migrations/down: downgrade all available database migrations
## : depends on the following environment variables:
## : - INSTRUMENT_SWAP_DB_DSN -> dsn of the used Postgres database
.PHONY: db/migrations/down
db/migrations/down:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) down

## db/migrations/version: lists the actual applied db migrations version
## : depends on the following environment variables:
## : - INSTRUMENT_SWAP_DB_DSN -> dsn of the used Postgres database
.PHONY: db/migrations/version
db/migrations/version:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) version

##db/migrations/force version=number: forcefully downmigrates to the specified version
## : depends on the following environment variables:
## : - INSTRUMENT_SWAP_DB_DSN -> dsn of the used Postgres database
.PHONY: db/migrations/force
db/migrations/force:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) force ${version}

# ============================================================================ #
# QUALITY CONTROL
# ============================================================================ #

## audit: tidy dependencies, formats, vets, lints and tests all code
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
