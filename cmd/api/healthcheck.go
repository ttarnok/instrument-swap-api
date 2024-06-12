package main

import (
	"net/http"
)

func (app *application) livelinessHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"liveliness": data}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
	}
}
