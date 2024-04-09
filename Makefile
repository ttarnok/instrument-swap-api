.PHONY: format
format:
	go fmt ./...

.PHONY: run
run: format
	go run ./cmd/api
