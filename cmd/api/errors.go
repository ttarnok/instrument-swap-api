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
		w.WriteHeader(500)
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
