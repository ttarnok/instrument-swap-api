package main

import (
	"errors"
	"net/http"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

// createAuthenticationTokenHandler implements a handler that respond with auth tokens.
func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	jwtBytesAccess, err := app.auth.AccessToken.NewToken(user.ID)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	jwtBytesRefresh, err := app.auth.RefreshToken.NewToken(user.ID)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"access": string(jwtBytesAccess), "refresh": string(jwtBytesRefresh)}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}
