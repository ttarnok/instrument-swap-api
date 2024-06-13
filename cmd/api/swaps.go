package main

import "net/http"

func (app *application) listSwapsHandler(w http.ResponseWriter, r *http.Request) {

	swaps, err := app.models.Swaps.GetAll()
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"swaps": swaps}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}
}
