FROM golang:1.23 AS deps

WORKDIR /instrument-swap-api
ADD *.mod *.sum ./
RUN go mod download

FROM deps AS dev
ADD . .
EXPOSE 8080
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o ./api ./cmd/api
ENTRYPOINT ["/instrument-swap-api/api"]

FROM scratch AS test

WORKDIR /
EXPOSE 8080
COPY --from=dev /instrument-swap-api/api /
ENTRYPOINT ["/api"]
