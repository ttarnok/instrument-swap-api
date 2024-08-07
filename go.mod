module github.com/ttarnok/instrument-swap-api

go 1.22.5

require github.com/lib/pq v1.10.9

require golang.org/x/time v0.5.0

require golang.org/x/crypto v0.22.0

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.6.0
	github.com/pascaldekloe/jwt v1.12.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)
