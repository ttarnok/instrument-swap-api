package main

import (
	"net/http"
)

// logError logs and error message, besides request related informations.
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// errorResponse sends well formatted error response to the client.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {

	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// serverErrorLogResponse sends InternalServerError response to the client and logs the error that caused it.
func (app *application) serverErrorLogResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)

}

// notFoundResponse send NotFound response to the client.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// failedValidationResponse send UnprocessableEntity error to the client.
// Also sends a fields/errors map that contains the failed entries during user input validation.
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// badRequestResponse sends BadRequest response to the client and an error that contains the problem.
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// editConflictResponse sends ConflictResponse response to the client.
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// swappedInstrumentResponse...
func (app *application) swappedInstrumentResponse(w http.ResponseWriter, r *http.Request) {
	message := "can not perform the operation on a swapped instrument"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// rateLimitExcededResponse sends TooManyRequests to the client.
func (app *application) rateLimitExcededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

// invalidCredentialsResponse sends Unauthorized response  to the client.
// Indicates invalid authentication credentials.
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// invalidAuthenticationTokenResponse sends Unauthorized response to the client.
// Indicates invalid or missing authentication token.
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// authenticationRequiredResponse sends Unauthorized response to the client.
// Indicates that the user is trying to request a resource that requires an authenticated user.
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// inactiveAccountResponse Forbidden response to the client.
// Indicates that the user that is initiating the request is not activated yet.
func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}
