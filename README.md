# instrument-swap-api

A backend service which implements the necessary endpoints for a frontend
that provides functionality for musicians to swap each other's instruments.

This is a personal project to learn Go by building a hands-on example application.

The application uses a Postgres database to store data and a Redis database to help JWT token support.

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

To run the application in a local development environment make sure the necessary environment variables are set then ```make docker/compose/up``` and then ```make docker/logs``` is a good starting point.

There is a `/postman` folder in the project root which contains an exported [Postman](https://www.postman.com) collection where you can try the provided API endpoints in action. Before you run the collection, please make sure the baseURL Collection variable is set properly according to the current application URL.

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

For a convenient development experience you can use a ```.env``` file in the process root folder to set the following environment variables (makefile expects these variables to be set). (For demonstration purposes only, I provided a .env file with basic dummy values that works for the development environment):

- **REDIS_ADDR:** address for the redis instance that serves the application (used by the application to connect to Redis); format: database_name:port
- **REDIS_PASSWORD:** password for the redis instance that serves the application (used by the application to connect to Redis) (for development, empty password is fine)
- **REDIS_DB:** database index for the redis instance that serves the application (used by the application to connect to Redis) (for development 0 is fine)
- **POSTGRES_USER:** database user for the postgres database that serves the application (used by Docker when creating the database)
- **POSTGRES_PASSWORD:** password for the database user for the postgres database that serves the application (used by Docker when creating the database)
- **INSTRUMENT_SWAP_API_PORT:** the port for the application endpoint (used by the application)
- **INSTRUMENT_SWAP_DB_DSN:** the dsn for the postgres database that serves the application (used by the application to connect to the Postgres database)
- **JWT_SECRET:** secret for jwt support (used by the application)

## API endpoints

### List users
GET `/v1/users`

Returns a detailed list of the registered users.

### Register a new user
POST `/v1/users`

Allows you to restister a new user.

The request body needs to be in JSON format and includes the following properties:
 - `name` - string - Required
 - `email` - string - Required
 - `password` - string - Required

Example
```
POST /v1/users

{
  "name": "John Doe",
  "email": "johndoe@example.com",
  "password": "secret"
}
```
The response body will contain the user details of the registered user.

### Update an existing user
PATCH `/v1/users/{id}`

Allows you to update an existing user. Requires authentication, the given user id in the url path should match the user id specified in the JTW Access Token claim.

The request body needs to be in JSON format and includes the user properties that need to be modified:
 - `name` - string
 - `email` - string

Example
```
PATCH /v1/users/1
Authorization: Bearer <YOUR ACCESS TOKEN>

{
  "name": "John Smith"
}
```
The response body will contain the user details of the updated user.

### Update the password of a user
PUT `/v1/users/{id}/password`

Allows you to update the password of an existing user. Requires authentication, the given user id in the url path should match the user id specified in the JTW Access Token claim.

The request body needs to be in JSON format and includes the old and the new password for the user:
 - `password` - string - Required
 - `new_password` - string - Required

Example
```
PUT /v1/users/{id}/password
Authorization: Bearer <YOUR ACCESS TOKEN>

{
  "password": "oldpassword",
  "new_password": "newpassword"
}
```

### Delete an existing user
DELETE `/v1/users/{id}`

Allows you to delete an user. Requires authentication, the given user id in the url path should match the user id specified in the JTW Access Token claim.

Example
```
DELETE /v1/users/1
Authorization: Bearer <YOUR ACCESS TOKEN>
```

### Show the list of instruments
GET `/v1/instruments`

Returns a detailed list of the instrumets. Requires authentication.

Optional query parameters:
- `name` - to query instruments with a specific name
- `manufacturer` - to query instruments with a specific manufacturer
- `famous_owners` - to query instruments with a specific famous owner
- `owner_user_id` - to query instruments with a specific owner user id
- `page` - to get the nth page of the result
- `page_size` - to specify how many instruments should be on a result page
- `sort` - to specify an attribute that we want to base the ordering of the result on
  - Possinble values: `id`, `name`, `manufacturer`, `type`, `manufacture_year`, `estimated_value`, `owner_user_id`, `-id`, `-name`, `-manufacturer`, `-type`, `-manufacture_year`, `-estimated_value`, `-owner_user_id`
  - Values starting with hyphen represents descending order, otherwise the ordering will be ascending

Example 1
```
GET /v1/instruments?sort=-name
Authorization: Bearer <YOUR ACCESS TOKEN>
```

Example 2
```
GET /v1/instruments?page_size=2&page=2
Authorization: Bearer <YOUR ACCESS TOKEN>
```

Example 3
```
GET /v1/instruments?page_size=1&page=10&sort=manufacturer
Authorization: Bearer <YOUR ACCESS TOKEN>
```

The response body will contain the list of the queried instruments and pagination related metadata information.

The returned metadata information contains:
- `current_page` - the current page index
- `page_size` - the page size
- `first_page` - the index of the first page
- `last_page` - the index of the last page
- `total_records` - the total number of instruments on all the pages

### Get the attributes of the specified instrument
GET `/v1/instruments/{id}`

Allows you to view an existing intsrument. Requires authentication.

Example
```
GET /v1/instruments/1
Authorization: Bearer <YOUR ACCESS TOKEN>
```

The response body will contain the details of the instrument with the given instrument id.

### Create a new instrument
POST `/v1/instruments`

Allows you the create a new instrument. Requires authentication.

The request body needs to be in JSON format and includes the user properties that need to be modified:
 - `name` - string - Required
 - `manufacturer` - string - Required
 - `manufacture_year` - int - Required
 - `type` - string - Required
   - accepted values: `synthesizer`, `guitar`
 - `estimated_value` - int - Required
 - `condition` - string - Required
 - `description` - string - Required
 - `famous_owners` - []string - Required

Example
```
POST /v1/instruments
Authorization: Bearer <YOUR ACCESS TOKEN>

{
  "name": "Instrument name",
  "manufacturer": "Famous company",
  "manufacture_year": 1990,
  "type": "guitar",
  "estimated_value": 10000,
  "condition": "outstanding",
  "description": "Here comes the description...",
  "famous_owners": ["Band name 1", "Band name 2"]
}
```

The response body will contain the details of the newly created instrument.

### Update an existing instrument
PATCH `/v1/instruments/{id}`

Allows you to update an existing instrument. Requires authentication, the given instrument id in the url path should match to an instrument with an owner user id specified in the JTW Access Token claim.

The request body needs to be in JSON format and includes the instrument properties that need to be modified:
- `name` - string
- `manufacturer` - string
- `manufacture_year` - int
- `type` - string
  - accepted values: `synthesizer`, `guitar`
- `estimated_value` - int
- `condition` - string
- `description` - string
- `famous_owners` - []string

Example
```
PATCH /v1/instruments/1
Authorization: Bearer <YOUR ACCESS TOKEN>

{
  "name": "Updated Instrument name",
  "condition": "good"
}
```
The response body will contain the details of the newly updated instrument.

### Delete an instrument
DELETE `/v1/instruments/{id}`

Deletes the instrument with the specified instrument id. Requires authentication, the given instrument id in the url path should match to an instrument with an owner user id specified in the JTW Access Token claim. Instrument with an ongoing swap cannotbe deleted.

Example
```
DELETE /v1/instruments/1
Authorization: Bearer <YOUR ACCESS TOKEN>
```

### Get the ongoing swaps
GET `/v1/swaps`

Returns the ongoing swaps of the authenticated user. Requires authentication.

Example
```
GET /v1/swaps
Authorization: Bearer <YOUR ACCESS TOKEN>
```
The response body will contain a list of the requested swaps.

### Get a specific swap
GET `/v1/swaps/{id}`

Returns the details of the given swap. Requires authentication. The given swap id should belong to the authenticated user.

Example
```
GET /v1/swaps/1
Authorization: Bearer <YOUR ACCESS TOKEN>
```
The response body will contain the details of the requested swap.

### Create a new swap request
POST `/v1/swaps`

Creates a new swap request. Requires authentication.

The request body needs to be in JSON format. The requester_instrument_id must belong to the authenticated user. You can use the following properties:
 - `requester_instrument_id` - int - Required
 - `recipient_instrument_id` - int - Required

Example
```
POST /v1/swaps
Authorization: Bearer <YOUR ACCESS TOKEN>

{
  "requester_instrument_id": 1210,
  "recipient_instrument_id": 4
}
```
The response body will contain the details of the newly created swap.

### Modify the state of a swap
PATCH `/v1/swaps/{id}`

Modifies the state of the given swap swap. Requires authentication.

The request body needs to be in JSON format and should contain can use the desired state for the swap with the following property:
- `status` - string - Required
  - possibel values: `accepted`, `rejected`, `ended`

Possible state changes:
  - A newly created swap can be accepted or rejected. Only the recipient user can accept or reject a swap.
  - An accepted swap can be ended. Both the requester user and the recipient user can end a swap.

Example
```
PATCH /v1/swaps/1
Authorization: Bearer <YOUR ACCESS TOKEN>

{
  "status": "ended"
}
```

The response body will contain the details of the updated swap.

### Log in the user, create a new Access and Refresh JWT Token pair
POST `/v1/token`

Logs in the user and returns the Access and Refresh tokens for the authentication of the logged in user.

The request body needs to be in JSON format and should contain the following user cretentials to log in:
- `email` - string - Required
- `password` - string - Required

Example
```
POST /v1/token

{
  "email": "johnsmith@example.com",
  "password": "mysecretpassword123"
}
```
The response body will contain the Access and Refresh Tokens.
- Access Tokens expires in 5 minutes.
- Refresh Tokens expires in 24 hours.

### Return a new Access Token
POST `/v1/token/refresh`

If we have an expired Access Token and a non expired Refresh Token, we can get a new Access Token and a new Refresh Token.

The request body needs to be in JSON format and should contain the following properties:
- `access` - string - Required
- `refresh` - string - Required

Example
```
POST /v1/token

{
  "access": "asdsadasfwft43r2wrffwf",
  "refresh": "sd,afkucghqwelf,kuabw.fKHJBDFLauzvf"
}
```
The response body will contain the newly created Access and Refresh Tokens. The old Access and Refresh Tokens can not used anymore.

### Invalidate a refresh token
POST `/v1/token/blacklist`

Invalidates the given Refresh Token, which is cannot be used anymore.

The request body needs to be in JSON format and should contain the Refresh Token to invalidate:
- `refresh` - string - Required

Example
```
POST /v1/token/blacklist

{
  "refresh": "sd,afkucghqwelf,kuabw.fKHJBDFLauzvf"
}
```

- POST   /v1/token/logout // Blacklists the given access and refresh tokens
### Logout the user, with token invalidation
POST `/v1/token/logout`

Logs out the  user with the given Access and Refresh Tokens. Invalidates both tokens.

The request body needs to be in JSON format and should contain the following properties:
- `access` - string - Required
- `refresh` - string - Required

Example
```
POST /v1/token/logout

{
  "access": "asdsadasfwft43r2wrffwf",
  "refresh": "sd,afkucghqwelf,kuabw.fKHJBDFLauzvf"
}
```

### Get Application status
GET `/v1/liveliness`

Returns the status of the application. This is not a readiness endpoint.

Example
```
GET  /v1/liveliness
```

The response body will contain the state, the environment and the version of the application.

### Display apprication metrics
GET `/debug/vars`

Displays apprication metrics. In production applications this endpoint should be protected.

Example
```
GET  /debug/vars
```

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
- Implement better metrics gathering. Consider [prometheus](https://prometheus.io) or [datadog](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/go/). For local development, here is a terminal based monitoring solution: [exvarmon](https://github.com/divan/expvarmon).
- Implement a basic CI/CD pipeline via [GitHub Actions](https://docs.github.com/en/actions) or [Jenkins](https://www.jenkins.io), and deploy the application to aws.
- Use [golang-jwt](github.com/golang-jwt/jwt) to handle jwt-s.
- Support to use multiple secrets for jwt-s. Consider using a keystore to load the secrets from.
  - Implement the key id claim, so we will know which key was used to sign the token.
- Refactor the application to use [Kubernetes](https://kubernetes.io).
  - Consider using [Kind](https://kind.sigs.k8s.io) or [Minikube](https://minikube.sigs.k8s.io/docs/) to run a local development cluster.
  - Consider using [kustomize](https://github.com/kubernetes-sigs/kustomize) to simplify Kubernetes related yaml files.
- Consider refactoring into Microservices architecture.
