.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: run
run: fmt
	@go run ./cmd/api

.PHONY: br
br: fmt
	@go build -o=./bin/api ./cmd/api
	./bin/api -port=4000
