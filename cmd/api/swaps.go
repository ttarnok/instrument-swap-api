package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

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

func (app *application) showSwapHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	swap, err := app.models.Swaps.Get(id)
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

	err = app.writeJSON(w, http.StatusOK, envelope{"swap": swap}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}

func (app *application) createSwapHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		RequesterInstrumentId int64 `json:"requester_instrument_id"`
		RecipientInstrumentId int64 `json:"recipient_instrument_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	swap := &data.Swap{
		RequesterInstrumentId: input.RequesterInstrumentId,
		RecipientInstrumentId: input.RecipientInstrumentId,
	}

	v := validator.New()

	if data.ValidateSwap(v, swap); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Check the instrument id-s are real.
	_, err = app.models.Instruments.Get(input.RecipientInstrumentId)
	if err != nil {
		fmt.Println(err)
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, errors.New("recipient instrument not found"))
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}
	_, err = app.models.Instruments.Get(input.RequesterInstrumentId)
	if err != nil {
		fmt.Println(err)
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, errors.New("requester instrument not found"))
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	// Check the instrument are not swapped
	rec_swap, err := app.models.Swaps.GetByInstrumentId(input.RecipientInstrumentId)
	if err != nil && !errors.Is(err, data.ErrRecordNotFound) {
		app.serverErrorLogResponse(w, r, err)
		return
	}
	if rec_swap != nil {
		app.badRequestResponse(w, r, errors.New("recipient instrument already in a swap"))
		return
	}
	req_swap, err := app.models.Swaps.GetByInstrumentId(input.RequesterInstrumentId)
	if err != nil && !errors.Is(err, data.ErrRecordNotFound) {
		app.serverErrorLogResponse(w, r, err)
		return
	}
	if req_swap != nil {
		app.badRequestResponse(w, r, errors.New("requester instrument already in a swap"))
		return
	}

	// Create the swap
	err = app.models.Swaps.Insert(swap)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/swaps/%d", swap.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"swap": swap}, headers)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}

func (app *application) acceptSwapHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	swap, err := app.models.Swaps.Get(id)
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

	if swap.IsAccepted || swap.IsRejected || swap.IsEnded {
		app.badRequestResponse(w, r, errors.New("swap is not acceptable"))
		return
	}

	swap.IsAccepted = true
	now := time.Now()
	swap.AcceptedAt = &now

	err = app.models.Swaps.Update(swap)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"swap": swap}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}
}
