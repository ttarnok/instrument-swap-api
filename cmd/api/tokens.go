package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

// loginHandler implements a handler that respond with auth and refresh tokens.
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

// refreshHandler handles new access token generation.
func (app *application) refreshHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		AccessToken  string `json:"access"`
		RefreshToken string `json:"refresh"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Parse the access token and check the fields.
	accessClaims, err := app.auth.AccessToken.ParseClaims([]byte(input.AccessToken))
	if err != nil {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Check access token validity.
	if app.auth.AccessToken.IsValid([]byte(input.AccessToken)) {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Check if the access token is blacklisted.
	isBlacklisted, err := app.auth.BlacklistToken.IsTokenBlacklisted(accessClaims.ID)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}
	if isBlacklisted {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Parse the refresh token and check the fields.
	refreshClaims, err := app.auth.RefreshToken.ParseClaims([]byte(input.RefreshToken))
	if err != nil {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Check refresh token validity.
	if !app.auth.RefreshToken.IsValid([]byte(input.RefreshToken)) {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Check if the access token is blacklisted.
	isBlacklisted, err = app.auth.BlacklistToken.IsTokenBlacklisted(refreshClaims.ID)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}
	if isBlacklisted {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Check whether the two tokens belong to the same user.
	if accessClaims.Subject != refreshClaims.Subject {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Check whether the user is a valid user in the app.
	userID, err := strconv.ParseInt(refreshClaims.Subject, 10, 64)
	if err != nil {
		app.invalidCredentialsResponse(w, r)
		return
	}

	user, err := app.models.Users.GetByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	// Blacklist the refresh token.
	err = app.auth.BlacklistToken.BlacklistToken(refreshClaims.ID)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	// Generate new access token.
	jwtBytesAccess, err := app.auth.AccessToken.NewToken(user.ID)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"access": string(jwtBytesAccess)}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}
