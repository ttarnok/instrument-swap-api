# instrument-swap-api

A backend service which implements the necessary endpoints for a frontend
that provides functionality for musicians to swap each other's instruments.

## Dependencies

- Go version 1.22

## Setup

## Usage

- [X] GET    /v1/users // Show detailed list of users
- [X] POST   /v1/users // Register a new user
- [X] PUT    /v1/users/{id} // Update a user
- [X] DELETE /v1/users/{id} // Delete a user
- [ ] PUT    /v1/users/activated // Activate a new user
- [X] PUT    /v1/users/password // Update the password of the user

- [X] GET    /v1/instruments // Show detailed list of instruments (pagination)
- [X] POST   /v1/instruments // Create a new instrument
- [X] GET    /v1/instruments/{id} // Show the details of a specific instrument
- [X] PATCH  /v1/instruments/{id} // Update a specific instrument
- [X] DELETE /v1/instruments/{id} // Delete a specific instrument

- [X] GET    /v1/swaps // Return the ongoing swap requests
- [ ] GET    /v1/swaps/{id} // Get a specific swap by id
- [ ] POST   /v1/swaps // Initiates a new swap request
- [ ] POST   /v1/swaps/{id}/accept // Accepts a swap request
- [ ] POST   /v1/swaps/{id}/reject // Rejects a swap request
- [ ] DELETE /v1/swaps/{id} // Ends an instrument swap

- [ ] POST   /v1/tokens/authentication // Return a new Access Token
- [ ] POST   /v1/logout // Logs out the current user

- [X] GET    /v1/liveliness
- [ ] GET    /v1/readyness

- [X] GET    /debug/vars // Display apprication metrics
- [X] GET    /debug/pprof // Display debug infos

## Release milestones

### TODO List
handle OPTIONS HTTP method
create custom 405, 404 middleware
consider using Cache-Control header response happypath/errorpath
