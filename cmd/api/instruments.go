package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

func (app *application) showInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	instrument, err := app.models.Instruments.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordnotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorLogResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"instrument": instrument}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
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
		return
	}

}

func (app *application) updateInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	instrument, err := app.models.Instruments.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordnotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name            *string  `json:"name"`
		Manufacturer    *string  `json:"manufacturer"`
		ManufactureYear *int32   `json:"manufacture_year"`
		Type            *string  `json:"type"`
		EstimatedValue  *int64   `json:"estimated_value"`
		Condition       *string  `json:"condition"`
		Description     *string  `json:"description"`
		FamousOwners    []string `json:"famous_owners"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		instrument.Name = *input.Name
	}
	if input.Manufacturer != nil {
		instrument.Manufacturer = *input.Manufacturer
	}
	if input.ManufactureYear != nil {
		instrument.ManufactureYear = *input.ManufactureYear
	}
	if input.Type != nil {
		instrument.Type = *input.Type
	}
	if input.EstimatedValue != nil {
		instrument.EstimatedValue = *input.EstimatedValue
	}
	if input.Condition != nil {
		instrument.Condition = *input.Condition
	}
	if input.Description != nil {
		instrument.Description = *input.Description
	}
	if input.FamousOwners != nil {
		instrument.FamousOwners = input.FamousOwners
	}

	v := validator.New()

	if data.ValidateInstrument(v, instrument); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Instruments.Update(instrument)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"instrument": instrument}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}

func (app *application) deleteInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Instruments.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordnotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "instrument successfully deleted"}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}
