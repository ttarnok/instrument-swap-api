package main

import (
	"net/http"
	"strconv"

	"github.com/ttarnok/instrument-swap-api/internal/data"
)

func (app *application) showInstrumentHandler(w http.ResponseWriter, r *http.Request) {
	// Read and Validate params
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	instrument := data.Instrument{
		Name:            "MS-20",
		Manufacturer:    "Korg",
		ManufactureYear: "1980",
		Type:            "Synthesiser",
		EstimatedValue:  100000,
		Condition:       "Excellent",
		FamousOwners:    []string{"Cher", "Don", "Eye"},
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"instrument": instrument}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
	}

}

func (app *application) createInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Creating a new instrument"))

}
