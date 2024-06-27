package main

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/data"
)

// TestContext tests the happy path of contextSetUser.
func TestContext(t *testing.T) {
	app := &application{}

	now := time.Now()
	inUser := &data.User{
		ID:        1,
		CreatedAt: now,
		Name:      "Test Usern",
		Email:     "test@example.com",
		Activated: true,
		Version:   1,
	}
	err := inUser.Password.Set("Welcome1")
	if err != nil {
		t.Fatalf("can not set password for test user")
	}

	req := &http.Request{}

	req = app.contextSetUser(req, inUser)
	outUser := app.contextGetUser(req)

	if !reflect.DeepEqual(inUser, outUser) {
		t.Errorf("wanted to get the same user from context, got different")
	}
}

// TestContextPanic tests the panic path of contextSetUser,
// when there is no User information in the provided context.
func TestContextPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic without user information from the context")
		}
	}()

	app := &application{}

	req := &http.Request{}

	_ = app.contextGetUser(req)
}
