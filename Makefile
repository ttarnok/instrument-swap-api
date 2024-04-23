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
	./bin/api -port=4000 -db-dsn=$(INSTRUMENT_SWAP_DB_DSN)

.PHONY: dcup
dcup:
	docker-compose up --build -d

.PHONY: dcdown
dcdown:
	docker-compose down
