package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/data/mocks"
)

// TestListInstrumentsHandler unit tests the functionality of listInstrumentsHandler.
func TestListInstrumentsHandler(t *testing.T) {

	now := time.Now().UTC()

	testData := []*data.Instrument{
		{
			ID:              1,
			CreatedAt:       now,
			Name:            "M1",
			Manufacturer:    "Korg",
			ManufactureYear: 1990,
			Type:            "synthesizer",
			EstimatedValue:  100000,
			Condition:       "used",
			Description:     "A music workstation manufactured by Korg.",
			FamousOwners:    []string{"The Orb"},
			OwnerUserID:     1,
			Version:         1,
		},
		{
			ID:              2,
			CreatedAt:       now,
			Name:            "Wavestation",
			Manufacturer:    "Korg",
			ManufactureYear: 1991,
			Type:            "synthesizer",
			EstimatedValue:  999,
			Condition:       "used",
			Description:     "A vector synthesis synthesizer first produced by Korg.",
			FamousOwners:    []string{"Depeche Mode", "Genesis"},
			OwnerUserID:     1,
			Version:         1,
		},
		{
			ID:              3,
			CreatedAt:       now,
			Name:            "DX7",
			Manufacturer:    "Yamaha",
			ManufactureYear: 1985,
			Type:            "synthesizer",
			EstimatedValue:  100000,
			Condition:       "used",
			Description:     "A frequency modulation based synthesizer manufactured by Yamaha.",
			FamousOwners:    []string{"Brian Eno"},
			OwnerUserID:     2,
			Version:         1,
		},
		{
			ID:              4,
			CreatedAt:       now,
			Name:            "ESQ-1",
			Manufacturer:    "Ensoniq",
			ManufactureYear: 1987,
			Type:            "synthesizer",
			EstimatedValue:  8000,
			Condition:       "used",
			Description:     "A digital wave morphing synthesizer manufactured by Ensoniq",
			FamousOwners:    []string{"Steve Roach"},
			OwnerUserID:     3,
			Version:         1,
		},
	}

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: data.Models{Instruments: mocks.NewNonEmptyInstrumentModelMock(testData)},
	}

	app.listInstrumentsHandler(rr, req)

	resp := rr.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(`expected status code %d, got %d`, http.StatusOK, resp.StatusCode)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	respEnvelope := make(envelope)

	err = json.Unmarshal(body, &respEnvelope)
	if err != nil {
		t.Fatal(err)
	}

	respAnySlice, ok := respEnvelope["instruments"]
	if !ok {
		t.Fatal(`the response does not contain enveloped instruments`)
	}
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(respAnySlice)
	if err != nil {
		t.Fatal(err)
	}

	var inst []*data.Instrument

	err = json.NewDecoder(buf).Decode(&inst)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(testData, inst) {
		t.Error(`input and output for the handler is not equal`)
	}

}
