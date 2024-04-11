.PHONY: format
format:
	@go fmt ./...

.PHONY: run
run: format
	@go run ./cmd/api

.PHONY: br
br: format
	@go build -o=./bin/api ./cmd/api
	./bin/api -port=4000
