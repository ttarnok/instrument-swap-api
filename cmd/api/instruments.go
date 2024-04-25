package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
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
		ManufactureYear: 1980,
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
		ManufactureYear int32    `json:"manufacture_year"`
		Type            string   `json:"type"`
		EstimatedValue  int64    `json:"estimated_value"`
		Condition       string   `json:"condition"`
		Description     string   `json:"description"`
		FamousOwners    []string `json:"famous_owners"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	instrument := &data.Instrument{
		Name:            input.Name,
		Manufacturer:    input.Manufacturer,
		ManufactureYear: input.ManufactureYear,
		Type:            input.Type,
		EstimatedValue:  input.EstimatedValue,
		Condition:       input.Condition,
		FamousOwners:    input.FamousOwners,
	}

	v := validator.New()

	if data.ValidateInstrument(v, instrument); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Instruments.Insert(instrument)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	// create a location header for the client, with the location of the newly created resource
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/instrumets/%d", instrument.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"instrument": instrument}, headers)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
	}

}
