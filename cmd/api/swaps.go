package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

const (
	StatusSwapAccepted = "accepted" // StatusSwapAccepted
	StatusSwapRejected = "rejected" // StatusSwapRejected
	StatusSwapEnded    = "ended"    // StatusSwapEnded
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

// showSwapHandler handles the retrieval of a swap with the given id for the user within the context.
func (app *application) showSwapHandler(w http.ResponseWriter, r *http.Request) {

	authUser := app.contextGetUser(r)

	id, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	swapsForUser, err := app.models.Swaps.GetAllForUser(authUser.ID)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

	var foundSwap *data.Swap
	for _, swap := range swapsForUser {
		if swap.ID == id {
			foundSwap = swap
			break
		}
	}

	if foundSwap == nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"swap": foundSwap}, nil)
	if err != nil {
		app.serverErrorLogResponse(w, r, err)
		return
	}

}

// createSwapHandler handles the creation of a new swap with the given instruments.
func (app *application) createSwapHandler(w http.ResponseWriter, r *http.Request) {

	ownerUser := app.contextGetUser(r)

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

	requesterInstrument, err := app.models.Instruments.Get(input.RequesterInstrumentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, errors.New("requester instrument not found"))
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	// Only the authenticated user can be the requester.
	if requesterInstrument.OwnerUserID != ownerUser.ID {
		app.forbiddenResponse(w, r)
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

// isValidInputSwapStatus checks whether the given status string is a uniform value that is accepted by the application.
func isValidInputSwapStatus(status string) bool {
	if status != StatusSwapAccepted && status != StatusSwapRejected && status != StatusSwapEnded {
		return false
	}
	return true
}

// validateSwapStatusTransition checks whether the requested state transition is possible from the current state.
// In case of any problem returns a corresponding error, otherwise returns nil.
func validateSwapStatusTransition(currentIsAccepted, currentIsRejected, currentIsEnded bool, requestedSwapStatus string) error {
	if requestedSwapStatus == StatusSwapRejected && (currentIsAccepted || currentIsRejected || currentIsEnded) {
		return errors.New("swap is not rejectable")
	}
	if requestedSwapStatus == StatusSwapAccepted && (currentIsAccepted || currentIsRejected || currentIsEnded) {
		return errors.New("swap is not acceptable")
	}
	return nil
}

// isValidSwapStatusTransitionUser validate whether the authorized user is permitted to perform the requested status transition.
func isValidSwapStatusTransitionUser(requestedSwapStatus string, requesterUserID int64, recipientUserID int64, authUserID int64) bool {
	if requestedSwapStatus == StatusSwapEnded {
		if requesterUserID == authUserID || recipientUserID == authUserID {
			return true
		}
	}
	if requestedSwapStatus == StatusSwapAccepted || requestedSwapStatus == StatusSwapRejected {
		if recipientUserID == authUserID {
			return true
		}
	}
	return false
}

// doSwapStatusTransition performs the swap status transition.
// Due to pointer semantics, the function mutates the given swap.
func doSwapStatusTransition(swap *data.Swap, requestedSwapStatus string) {
	now := time.Now()
	if requestedSwapStatus == StatusSwapAccepted {
		swap.IsAccepted = true
		swap.AcceptedAt = &now
	} else if requestedSwapStatus == StatusSwapRejected {
		swap.IsRejected = true
		swap.IsEnded = true
		swap.RejectedAt = &now
		swap.EndedAt = &now
	} else if requestedSwapStatus == StatusSwapEnded {
		swap.IsEnded = true
		swap.EndedAt = &now
	}
}

// updateSwapStatusHandler handles the possible status changes of the swaps.
func (app *application) updateSwapStatusHandler(w http.ResponseWriter, r *http.Request) {
	authUser := app.contextGetUser(r)

	var input struct {
		Status string `json:"status"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !isValidInputSwapStatus(input.Status) {
		app.errorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("not valid status value: %q", input.Status))
		return
	}

	swapID, err := app.extractIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	swap, err := app.models.Swaps.Get(swapID)
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

	if err := validateSwapStatusTransition(swap.IsAccepted, swap.IsRejected, swap.IsEnded, input.Status); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	recipientInstrument, err := app.models.Instruments.Get(swap.RecipientInstrumentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, errors.New("recipient instrument not found"))
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	requesterInstrument, err := app.models.Instruments.Get(swap.RequesterInstrumentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, errors.New("recipient instrument not found"))
		default:
			app.serverErrorLogResponse(w, r, err)
		}
		return
	}

	if !isValidSwapStatusTransitionUser(input.Status, requesterInstrument.OwnerUserID, recipientInstrument.OwnerUserID, authUser.ID) {
		app.forbiddenResponse(w, r)
		return
	}

	doSwapStatusTransition(swap, input.Status)

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
