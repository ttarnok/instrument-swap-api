package main

import (
	"net/http"
)

// Helper for logging errors in handlers.
// TODO: consider to use error handling middleware.
// TODO: consider to log trace id too.
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// Helper for sending JSON formatted error responses.
// TODO: consider to use error handloing middleware.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {

	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Helper to respond internal server error to the client.
// TODO: consider using error handling middleware.
func (app *application) serverErrorLogResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)

}

// Helper to respond json validation errors.
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err)
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

func (app *application) rateLimitExcededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded333"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}
