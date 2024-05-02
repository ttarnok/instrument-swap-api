INSTRUMENT_SWAP_DB_DSN := "postgres://instrumentswap:s3cr3t@localhost/instrumentswap?sslmode=disable"

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: run
run: fmt
	@go run ./cmd/api

.PHONY: br
br: fmt

	@go build -o=./bin/api ./cmd/api
	./bin/api -port=4000 -db-dsn=$(INSTRUMENT_SWAP_DB_DSN) -limiter-enabled=false

.PHONY: dcup
dcup:
	docker-compose up --build -d

.PHONY: dcdown
dcdown:
	docker-compose down

.PHONY: mup
mup:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) up

.PHONY: mdown
mdown:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) down

.PHONY: mver
mver:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) version

mforce1:
	migrate -path=./migrations -database=$(INSTRUMENT_SWAP_DB_DSN) force 1
