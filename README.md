# instrument-swap-api

A backend service which implements the necessary endpoints for a frontend
that provides functionality for musicians to swap each other's instruments.\
This is a personal project to learn Go by building a hands-on example application.

## Dependencies

1. [Go version 1.23.0](https://go.dev/dl/)
2. [Docker Desktop 4.33.0](https://www.docker.com/products/docker-desktop/)
3. [staticcheck 2024.1.1 (0.5.1)](https://staticcheck.dev)
4. [golangci-lint 1.60.2](https://golangci-lint.run)
5. [golang-migrate v4.17.0](https://github.com/golang-migrate/migrate)

## Setup

Development expects to run in a Unix-like system with support for Makefile and Git.\
To set up a development environment, please go through the following steps:

1. Make sure you have a Unix-like system with makefile support.
2. Make sure that Git is installed and configured properly.
3. Make sure to install all the listed dependencies in the proper order.
4. Clone the project code from GitHub.
5. To see development options, run ```make help``` from the project's root directory.

## Usage

Running the application with the ```--help``` or ```-h``` flags will display the possible command-line options and exit.

Command line options for the application (If a value is not set explicitly, the default value will be set, if there is any.):
- **help:** discard the other given flags, display the possible command-line options and exit.
- **version:** discard the other given flags, display the version of the application and exit.
- **environment:** the running environment with the possible values of dev, test or prod (default value is dev)
- **port:** API server port (default value is 4000)
- **db-dsn:** dsn for the used Postgres database in the format of "postgres://username:password@host/database?params" (there is no default value)
- **db-max-open-conns:** maximum number of open database connections for the Postgres database. (default value is 25)
- **db-max-idle-conns:** maximum number of idle database connections for the Postgres database. (default value is 25)
- **db-max-idle-time:** maximum connection idle timein nanoseconds for the Postgres database. (default value is 15 minutes)
- **limiter-rps:** maximum requests per second for the rate limiter (default value is 2)
- **limiter-burst:** maximum burst for rate limiter (default value is 4)
- **limiter-enabled:** a boolean value to enable or disable the limiter. (default value is true)
- **jwt-secret:** the secret for jwt token creation and signature verification.
- **redis-address:** the address of the used redis database in the form of host:port (there is no default value)
- **redis-password:** the password for redis (the default value is empty string)
- **redis-db:** the number of the used redis database (default value is 0)

For a convenient development experience you can use a ```.env``` file in the process root folder to set the following environment variables (makefile expects these variables to be set):
- **INSTRUMENT_SWAP_API_PORT**
- **INSTRUMENT_SWAP_DB_DSN**
- **JWT_SECRET**
- **REDIS_ADDR**
- **REDIS_PASSWORD**
- **REDIS_DB**

### API endpoints
- GET    /v1/users // Show detailed list of users
- POST   /v1/users // Register a new user
- PUT    /v1/users/{id} // Update a user
- DELETE /v1/users/{id} // Delete a user
- PUT    /v1/users/password // Update the password of the user

- GET    /v1/instruments // Show detailed list of instruments (pagination)
- POST   /v1/instruments // Create a new instrument
- GET    /v1/instruments/{id} // Show the details of a specific instrument
- PATCH  /v1/instruments/{id} // Update a specific instrument
- DELETE /v1/instruments/{id} // Delete a specific instrument

- GET    /v1/swaps // Return the ongoing swap requests
- GET    /v1/swaps/{id} // Get a specific swap by id
- POST   /v1/swaps // Initiates a new swap request
- POST   /v1/swaps/{id}/accept // Accepts a swap request
- POST   /v1/swaps/{id}/reject // Rejects a swap request
- DELETE /v1/swaps/{id} // Ends an instrument swap

- POST   /v1/token // Return a new Access Token + Refresh Token
- POST   /v1/token/refresh // Return a new Access Token
- POST   /v1/token/blacklist // Blacklists a refresh token
- POST   /v1/token/logout // Blacklists the given access and refresh tokens

- GET    /v1/liveliness

- GET    /debug/vars // Display apprication metrics
- GET    /debug/pprof // Display debug infos

## Release milestones

### TODO List
- Implement user activation.
- Replace pq database driver to [pgx](https://github.com/jackc/pgx).
- Refactor the whole application from the [fat service pattern](https://www.alexedwards.net/blog/the-fat-service-pattern) into more decoupled parts.
  - Every decoupled part should depend only on the code that they really use. (Interface segregation principle: "Clients should not be forced to depend upon interfaces that they do not use.")
  - Consider making dependencies abstract (dependency injection via depending on interfaces), so the code will become more modular, unit tests will become more clear, unit tests won't depend on the behaviour of the dependencies.
  - Refactor unit tests:
    - Use assertions, mocks and test suites from [testify](https://github.com/stretchr/testify).
    - Refactor all redundant test related functionality into helper functions.
- Implement better support for user roles. Consider storing roles in the jwt access tokens.
- Consider using [viper](https://github.com/spf13/viper) for better configuration support.
- Make sure every part of the application logs when it should log:
  - Create informative logs to inform about the state changes of the application. (startup/shutdown etc.)
  - Create a trace id generator middleware that generates a traceid for every incoming request and put this trace id into the request context.
  - Create a logger middleware that logs every request start/completion with trace id.
  - Make sure every time an error is handled by code, it is logged (and logged exactly once).
  - Make sure every panic handling is logged with the current stack trace information.
- Consider using [Elasticsearch](https://www.elastic.co/elasticsearch) [Logstash](https://www.elastic.co/logstash) [Kibana](https://www.elastic.co/kibana) stack for logs
- Implement tracing via [OpenTelemetry](https://opentelemetry.io/docs/languages/go/) and [Zipkin](https://zipkin.io)
- Make sure the api is not leaking implementation details. Make sure the error responses does not leaking implementation details:
  - Only return standars uniform errors. (safe errors)
  - In case of nonstandard errors return internal server error. (non safe errors)
- Implement better metrics gathering. [prometheus](https://prometheus.io) or [datadog](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/go/). For local development, here is a terminal based monitoring solution: [exvarmon](https://github.com/divan/expvarmon).
- Implement a basic CI/CD pipeline via [GitHub Actions](https://docs.github.com/en/actions) or [Jenkins](https://www.jenkins.io), and deploy it to aws.
- Use [golang-jwt](github.com/golang-jwt/jwt) to handle jwt-s.
- Support to use multiple secrets for jwt-s. Consider using a keystore to load the secrets from.
  - Implement the key id claim, so we will know which key was used to sign the token.
- Refactor the application to use [Kubernetes](https://kubernetes.io).
  - Consider using [Kind](https://kind.sigs.k8s.io) or [Minikube](https://minikube.sigs.k8s.io/docs/) to run a local development cluster.
  - Consider using [kustomize](https://github.com/kubernetes-sigs/kustomize) to simplify yaml files.
- Consider refactoring into Microservices architecture.
