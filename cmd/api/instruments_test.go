package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/data/mocks"
)

// testData contains some dummy initial data for the mocked database.
var testData = []*data.Instrument{
	{
		ID:              1,
		CreatedAt:       time.Now().UTC(),
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
		CreatedAt:       time.Now().UTC(),
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
		CreatedAt:       time.Now().UTC(),
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
		CreatedAt:       time.Now().UTC(),
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

// TestListInstrumentsHandler unit tests the functionality of listInstrumentsHandler.
func TestListInstrumentsHandler(t *testing.T) {

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

// TestShowInstrumentHandler unit tests the functionality of showInstrumentHandler.
func TestShowInstrumentHandler(t *testing.T) {

	tests := []struct {
		name               string
		pathParam          string
		expectedStatusCode int
		checkResult        bool
		expectedIndex      int
	}{
		{
			name:               "happy path",
			pathParam:          "1",
			expectedStatusCode: http.StatusOK,
			checkResult:        true,
			expectedIndex:      0,
		},
		{
			name:               "non number path",
			pathParam:          "nonnumber",
			expectedStatusCode: http.StatusNotFound,
			checkResult:        false,
		},
		{
			name:               "non exsistent id",
			pathParam:          "100",
			expectedStatusCode: http.StatusNotFound,
			checkResult:        false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Instruments: mocks.NewNonEmptyInstrumentModelMock(testData)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("GET /{id}", app.showInstrumentHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tt.pathParam)

			request, err := http.NewRequest("GET", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			res, err := client.Do(request)
			if err != nil {
				log.Fatal(err)
			}

			if res.StatusCode != tt.expectedStatusCode {
				t.Errorf(`expected status code %d, got %d`, tt.expectedStatusCode, res.StatusCode)
			}

			if tt.checkResult {
				bs, err := io.ReadAll(res.Body)
				if err != nil {
					log.Fatal(err)
				}
				defer func() {
					err := res.Body.Close()
					if err != nil {
						t.Fatal(err)
					}
				}()

				var env map[string]*data.Instrument

				err = json.Unmarshal(bs, &env)
				if err != nil {
					t.Fatal(err)
				}
				instrument := env["instrument"]

				if !reflect.DeepEqual(instrument, testData[tt.expectedIndex]) {
					t.Errorf(`expected message body "%#v", got "%#v"`, testData[tt.expectedIndex], instrument)
				}
			}
		})
	}

}
