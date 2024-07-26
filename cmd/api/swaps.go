package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

// listSwapsHandler handles listing all swaps for the user within the context.
func (app *application) listSwapsHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	swaps, err := app.models.Swaps.GetAllForUser(user.ID)
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

// showSwapHandler handles the retrieval of a swap with the given id.
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

// createSwapHandler handles the creation of a new swap with the given instruments.
func (app *application) createSwapHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		RequesterInstrumentID int64 `json:"requester_instrument_id"`
		RecipientInstrumentID int64 `json:"recipient_instrument_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	swap := &data.Swap{
		RequesterInstrumentID: input.RequesterInstrumentID,
		RecipientInstrumentID: input.RecipientInstrumentID,
	}

	v := validator.New()

	if data.ValidateSwap(v, swap); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Check the instrument id-s are real.
	_, err = app.models.Instruments.Get(input.RecipientInstrumentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, errors.New("recipient instrument not found"))
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}
	_, err = app.models.Instruments.Get(input.RequesterInstrumentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, errors.New("requester instrument not found"))
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	// Check the instrument are not swapped
	recSwap, err := app.models.Swaps.GetByInstrumentID(input.RecipientInstrumentID)
	if err != nil && !errors.Is(err, data.ErrRecordNotFound) {
		app.serverErrorLogResponse(w, r, err)
		return
	}
	if recSwap != nil {
		app.badRequestResponse(w, r, errors.New("recipient instrument already in a swap"))
		return
	}
	reqSwap, err := app.models.Swaps.GetByInstrumentID(input.RequesterInstrumentID)
	if err != nil && !errors.Is(err, data.ErrRecordNotFound) {
		app.serverErrorLogResponse(w, r, err)
		return
	}
	if reqSwap != nil {
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

// acceptSwapHandler handles the acception of the given swap.
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

// rejectSwapHandler handles the rejection of the given swap.
func (app *application) rejectSwapHandler(w http.ResponseWriter, r *http.Request) {
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
		app.badRequestResponse(w, r, errors.New("swap is not rejectable"))
		return
	}

	swap.IsRejected = true
	swap.IsEnded = true
	now := time.Now()
	swap.RejectedAt = &now
	swap.EndedAt = &now

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

// endSwapHandler handles the end of a swap withthe given id.
func (app *application) endSwapHandler(w http.ResponseWriter, r *http.Request) {
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

	swap.IsEnded = true
	now := time.Now()
	swap.EndedAt = &now

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
