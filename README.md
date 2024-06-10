# instrument-swap-api

A backend service which implements the necessary endpoints for a frontend
that provides functionality for musicians to swap each other's instruments.

## Dependencies

- Go version 1.22

## Setup

## Usage

[ ] GET    /v1/users // Show detailed list of users
[ ] POST   /v1/users // Register a new user
[ ] PUT    /v1/users/{id} // Update a user
[ ] DELETE /v1/users/{id} // Delete a user
[ ] PUT    /v1/users/activated // Activate a new user
[ ] PUT    /v1/users/password // Update the password of the user

[ ] GET    /v1/instruments // Show detailed list of instruments (pagination)
[ ] POST   /v1/instruments // Create a new instrument
[ ] GET    /v1/instruments/{id} // Show the details of a specific instrument
[ ] PATCH  /v1/instruments/{id} // Update a specific instrument
[ ] DELETE /v1/instruments/{id} // Delete a specific instrument

[ ] GET    /v1/swaps // Return the ongoing swap requests
[ ] POST   /v1/swaps // Initiates a new swap request
[ ] POST   /v1/swaps/{id}/accept // Accepts a swap request
[ ] POST   /v1/swaps/{id}/reject // Rejects a swap request
[ ] DELETE /v1/swaps/{id} // Ends an instrument swap

[ ] POST   /v1/tokens/authentication // Return a new Access Token
[ ] POST   /v1/logout // Logs out the current user

[ ] GET    /v1/liveliness
[ ] GET    /v1/readyness

[ ] GET    /debug/vars // Display apprication metrics
[ ] GET    /debug/pprof // Display debug infos

## Release milestones

### TODO List
handle OPTIONS HTTP method
create custom 405, 404 middleware
consider using Cache-Control header response happypath/errorpath
