package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

// listInstrumentsHandler is responsible for listing instruments.
func (app *application) listInstrumentsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name         string
		Manufacturer string
		Type         string
		FamousOwners []string
		OwnerUserID  int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readQParamString(qs, "name", "")
	input.Manufacturer = app.readQParamString(qs, "manufacturer", "")
	input.Type = app.readQParamString(qs, "type", "")
	input.FamousOwners = app.readQParamCSV(qs, "famous_owners", []string{})

	input.OwnerUserID = int64(app.readQParamInt(qs, "owner_user_id", 0, v))

	input.Page = app.readQParamInt(qs, "page", 1, v)
	input.PageSize = app.readQParamInt(qs, "page_size", 20, v)

	input.Sort = app.readQParamString(qs, "sort", "id")
	input.SortSafeList = []string{"id", "name", "manufacturer", "type", "manufacture_year", "estimated_value", "owner_user_id",
		"-id", "-name", "-manufacturer", "-type", "-manufacture_year", "-estimated_value", "-owner_user_id"}

	data.ValidateFilters(v, input.Filters)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	instruments, metadata, err := app.models.Instruments.GetAll(input.Name, input.Manufacturer, input.Type, input.FamousOwners, input.OwnerUserID, input.Filters)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"instruments": instruments, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}
}

// showInstrumentHandler shows a specific instrument.
func (app *application) showInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	instrument, err := app.models.Instruments.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
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

// createInstrumentHandler creates a new instrument.
func (app *application) createInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	var input struct {
		Name            string   `json:"name"`
		Manufacturer    string   `json:"manufacturer"`
		ManufactureYear int32    `json:"manufacture_year"`
		Type            string   `json:"type"`
		EstimatedValue  int64    `json:"estimated_value"`
		Condition       string   `json:"condition"`
		Description     string   `json:"description"`
		FamousOwners    []string `json:"famous_owners"`
		// OwnerUserID     int64    `json:"owner_user_id"`
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
		Description:     input.Description,
		FamousOwners:    input.FamousOwners,
		OwnerUserID:     user.ID,
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

// updateInstrumentHandler updates an instrument.
// JSON items with null values will be ignored and will remain unchanged.
func (app *application) updateInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	ownerUser := app.contextGetUser(r)

	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	instrument, err := app.models.Instruments.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	if ownerUser.ID != instrument.OwnerUserID {
		app.forbiddenResponse(w, r)
		return
	}

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

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != "" {
		instrument.Name = input.Name
	}
	if input.Manufacturer != "" {
		instrument.Manufacturer = input.Manufacturer
	}

	if input.ManufactureYear != 0 {
		instrument.ManufactureYear = input.ManufactureYear
	}
	if input.Type != "" {
		instrument.Type = input.Type
	}
	if input.EstimatedValue != 0 {
		instrument.EstimatedValue = input.EstimatedValue
	}
	if input.Condition != "" {
		instrument.Condition = input.Condition
	}
	if input.Description != "" {
		instrument.Description = input.Description
	}
	if len(input.FamousOwners) != 0 {
		instrument.FamousOwners = input.FamousOwners
	}

	v := validator.New()

	if data.ValidateInstrument(v, instrument); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Instruments.Update(instrument)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"instrument": instrument}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}

// deleteInstrumentHandler deletes an instrument.
func (app *application) deleteInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Instruments.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrConflict):
			app.badRequestResponse(w, r, errors.New("can not perform the operation on a swapped instrument"))
		case errors.Is(err, data.ErrRecordNotFound):
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
