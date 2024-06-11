package main

import (
	"expvar"
	"net/http"
)

func (app *application) routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/liveliness", app.livelinessHandler)

	mux.HandleFunc("GET /v1/instruments", app.listInstrumentsHandler)
	mux.HandleFunc("GET /v1/instruments/{id}", app.showInstrumentHandler)
	mux.HandleFunc("POST /v1/instruments", app.createInstrumentHandler)
	mux.HandleFunc("PATCH /v1/instruments/{id}", app.updateInstrumentHandler)
	mux.HandleFunc("DELETE /v1/instruments/{id}", app.deleteInstrumentHandler)

	mux.HandleFunc("GET /v1/users", app.listUsersHandler)
	mux.HandleFunc("POST /v1/users", app.registerUserHandler)
	mux.HandleFunc("PUT /v1/users/{id}/password", app.updatePasswordHandler)
	mux.HandleFunc("PUT /v1/users/{id}", app.updateUserHandler)
	mux.HandleFunc("DELETE /v1/users/{id}", app.deleteUserHandler)
	mux.HandleFunc("POST /v1/tokens/authentication", app.createAuthenticationTokenHandler)

	mux.Handle("GET /debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(mux)))))
}
