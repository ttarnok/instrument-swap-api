package main

import (
	"context"
	"net/http"

	"github.com/ttarnok/instrument-swap-api/internal/data"
)

type contextKey string

// userContextKey is a key for user values within the context.
const userContextKey = contextKey("user")

// contextSetUser injects user information into the context within the request.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser retrieves user information from the context within the request.
// Panics if no user data found inside the context.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user

}
