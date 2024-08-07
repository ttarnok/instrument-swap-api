package main

import (
	"expvar"
	"net/http"
)

func (app *application) routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/liveliness", app.livelinessHandler)

	mux.HandleFunc("GET /v1/instruments", app.requireActivatedUser(app.listInstrumentsHandler))
	mux.HandleFunc("GET /v1/instruments/{id}", app.requireActivatedUser(app.showInstrumentHandler))
	mux.HandleFunc("POST /v1/instruments", app.requireActivatedUser(app.createInstrumentHandler))
	mux.HandleFunc("PATCH /v1/instruments/{id}", app.requireActivatedUser(app.updateInstrumentHandler))
	mux.HandleFunc("DELETE /v1/instruments/{id}", app.requireActivatedUser(app.deleteInstrumentHandler))

	mux.HandleFunc("GET /v1/users", app.listUsersHandler)
	mux.HandleFunc("POST /v1/users", app.registerUserHandler)
	mux.HandleFunc("PUT /v1/users/{id}/password", app.updatePasswordHandler)
	mux.HandleFunc("PUT /v1/users/{id}", app.updateUserHandler)
	mux.HandleFunc("DELETE /v1/users/{id}", app.deleteUserHandler)

	mux.HandleFunc("POST /v1/token", app.loginHandler)
	mux.HandleFunc("POST /v1/token/refresh", app.refreshHandler)
	mux.HandleFunc("POST /v1/token/blacklist", app.blacklistHandler)

	mux.HandleFunc("GET /v1/swaps", app.requireActivatedUser(app.listSwapsHandler))
	mux.HandleFunc("POST /v1/swaps", app.requireActivatedUser(app.createSwapHandler))
	mux.HandleFunc("GET /v1/swaps/{id}", app.requireActivatedUser(app.showSwapHandler))
	mux.HandleFunc("POST /v1/swaps/{id}/accept", app.requireActivatedUser(app.acceptSwapHandler))
	mux.HandleFunc("POST /v1/swaps/{id}/reject", app.requireActivatedUser(app.rejectSwapHandler))
	mux.HandleFunc("POST /v1/swaps/{id}/end", app.requireActivatedUser(app.endSwapHandler))

	mux.Handle("GET /debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(mux)))))
}
