package main

import (
	"errors"
	"net/http"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

func (app *application) listUsersHandler(w http.ResponseWriter, r *http.Request) {

	users, err := app.models.Users.GetAll()
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: true, // TODO: implement user activation
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}
