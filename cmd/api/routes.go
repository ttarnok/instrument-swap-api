package main

import "net/http"

func (app *application) routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)

	mux.HandleFunc("GET /v1/instruments", app.listInstrumentsHandler)
	mux.HandleFunc("GET /v1/instruments/{id}", app.showInstrumentHandler)
	mux.HandleFunc("POST /v1/instruments", app.createInstrumentHandler)
	mux.HandleFunc("PATCH /v1/instruments/{id}", app.updateInstrumentHandler)
	mux.HandleFunc("DELETE /v1/instruments/{id}", app.deleteInstrumentHandler)

	return app.recoverPanic(mux)
}
