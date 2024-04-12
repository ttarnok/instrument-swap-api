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

	var input struct {
		Name            string   `json:"name"`
		Manufacturer    string   `json:"manufacturer"`
		ManufactureYear string   `json:"manufacture_year"`
		Type            string   `json:"type"`
		EstimatedValue  int64    `json:"estimated_value"`
		Condition       string   `json:"condition"`
		FamousOwners    []string `json:"famous_owners"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"test": input}, nil)
}
