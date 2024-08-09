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

// TestListSwapsHandler implements unit tests for listSwapsHandler.
func TestListSwapsHandler(t *testing.T) {

	testSwaps := []*data.Swap{
		{
			ID:                    1,
			CreatedAt:             time.Now().UTC(),
			RequesterInstrumentID: 1,
			RecipientInstrumentID: 2,
			Version:               1,
		},
		{
			ID:                    2,
			CreatedAt:             time.Now().UTC(),
			RequesterInstrumentID: 3,
			RecipientInstrumentID: 4,
			Version:               2,
		},
	}

	type testCase struct {
		name               string
		user               *data.User
		shouldCheckBody    bool
		expectedStatusCode int
		expectedTestSwaps  []*data.Swap
	}

	testCases := []testCase{
		{
			name:               "happy path",
			user:               &data.User{ID: 1},
			shouldCheckBody:    true,
			expectedStatusCode: http.StatusOK,
			expectedTestSwaps:  testSwaps,
		},
		{
			name:               "non exitent",
			user:               &data.User{ID: 1},
			shouldCheckBody:    false,
			expectedStatusCode: http.StatusInternalServerError,
			expectedTestSwaps:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			rr := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Swaps: mocks.NewSwapModelMock(tc.expectedTestSwaps)},
			}

			req = app.contextSetUser(req, tc.user)

			app.listSwapsHandler(rr, req)

			resp := rr.Result()

			if tc.expectedStatusCode != resp.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, resp.StatusCode)
			}

			if tc.shouldCheckBody {

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

				respAnySlice, ok := respEnvelope["swaps"]
				if !ok {
					t.Fatal(`the response does not contain enveloped swaps`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var swaps []*data.Swap

				err = json.NewDecoder(buf).Decode(&swaps)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tc.expectedTestSwaps, swaps) {
					t.Errorf("expected swap values\n%#v, got\n%#v", tc.expectedTestSwaps, swaps)
				}
			}

		})
	}

}

// TestShowSwapHandler implements unti test for showSwapHandler.
func TestShowSwapHandler(t *testing.T) {

	testSwaps := []*data.Swap{
		{
			ID:                    1,
			CreatedAt:             time.Now().UTC(),
			RequesterInstrumentID: 1,
			RecipientInstrumentID: 2,
			Version:               1,
		},
	}

	type testCase struct {
		name               string
		pathParam          string
		shouldCheckBody    bool
		baseSwaps          []*data.Swap
		expectedStatusCode int
		expectedTestSwap   *data.Swap
	}

	testCases := []testCase{
		{
			name:               "happy path",
			pathParam:          "1",
			shouldCheckBody:    true,
			baseSwaps:          testSwaps,
			expectedStatusCode: http.StatusOK,
			expectedTestSwap:   testSwaps[0],
		},
		{
			name:               "non numberic path param",
			pathParam:          "non numberic",
			shouldCheckBody:    false,
			baseSwaps:          nil,
			expectedStatusCode: http.StatusNotFound,
			expectedTestSwap:   nil,
		},
		{
			name:               "non existent user id",
			pathParam:          "99",
			shouldCheckBody:    false,
			baseSwaps:          nil,
			expectedStatusCode: http.StatusNotFound,
			expectedTestSwap:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Swaps: mocks.NewSwapModelMock(tc.baseSwaps)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("GET /{id}", app.showSwapHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tc.pathParam)

			req, err := http.NewRequest("GET", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if tc.expectedStatusCode != resp.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, resp.StatusCode)
			}

			if tc.shouldCheckBody {

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

				respAnySlice, ok := respEnvelope["swap"]
				if !ok {
					t.Fatal(`the response does not contain enveloped swaps`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var swap *data.Swap

				err = json.NewDecoder(buf).Decode(&swap)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tc.expectedTestSwap, swap) {
					t.Errorf("expected swap values\n%#v, got\n%#v", tc.expectedTestSwap, swap)
				}
			}

		})
	}
}

// TestCreateSwapHandler impolements unit tests for createSwapHandler.
func TestCreateSwapHandler(t *testing.T) {

	type inputSwap struct {
		RequesterInstrumentID int64 `json:"requester_instrument_id"`
		RecipientInstrumentID int64 `json:"recipient_instrument_id"`
	}

	type testCase struct {
		name               string
		input              inputSwap
		reqUser            data.User
		instruments        []*data.Instrument
		expectedStatusCode int
		shouldCheckBody    bool
	}

	testInstruments := []*data.Instrument{
		{
			ID:              1,
			CreatedAt:       time.Now().UTC(),
			Name:            "TB303",
			Manufacturer:    "Roland",
			ManufactureYear: 1990,
			Type:            "synthesizer",
			EstimatedValue:  100000,
			Condition:       "used",
			Description:     "A bass synth manufactured by Roland.",
			FamousOwners:    []string{"Carbon Based Lifeforms"},
			OwnerUserID:     1,
			Version:         1,
		},
		{
			ID:              2,
			CreatedAt:       time.Now().UTC(),
			Name:            "TR909",
			Manufacturer:    "Roland",
			ManufactureYear: 1990,
			Type:            "synthesizer",
			EstimatedValue:  100000,
			Condition:       "used",
			Description:     "A drum machine manufactured by Roland.",
			FamousOwners:    []string{"The Orb"},
			OwnerUserID:     1,
			Version:         1,
		},
	}

	testCases := []testCase{
		{
			name: "happy path",
			input: inputSwap{
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 2,
			},
			reqUser:            data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			instruments:        testInstruments,
			expectedStatusCode: http.StatusCreated,
			shouldCheckBody:    true,
		},
		{
			name: "create a swap for another user",
			input: inputSwap{
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 2,
			},
			reqUser:            data.User{ID: 3, Name: "Test User", Email: "testuser@example.com"},
			instruments:        testInstruments,
			expectedStatusCode: http.StatusForbidden,
			shouldCheckBody:    false,
		},
		{
			name: "non existent RequesterInstrumentID",
			input: inputSwap{
				RequesterInstrumentID: 10,
				RecipientInstrumentID: 2,
			},
			reqUser:            data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			instruments:        testInstruments,
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckBody:    false,
		},
		{
			name: "non existent RecipientInstrumentID",
			input: inputSwap{
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 20,
			},
			reqUser:            data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			instruments:        testInstruments,
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckBody:    false,
		},
		{
			name: "invalid RecipientInstrumentID",
			input: inputSwap{
				RequesterInstrumentID: 1,
				RecipientInstrumentID: -2,
			},
			reqUser:            data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			instruments:        testInstruments,
			expectedStatusCode: http.StatusUnprocessableEntity,
			shouldCheckBody:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{
					Swaps:       mocks.NewSwapModelMock(nil),
					Instruments: mocks.NewNonEmptyInstrumentModelMock(tc.instruments),
				},
			}

			setUser := func(next http.HandlerFunc) http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					r = app.contextSetUser(r, &tc.reqUser)

					next.ServeHTTP(w, r)
				})
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /", setUser(app.createSwapHandler))

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/", ts.URL)

			bs, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("POST", path, bytes.NewBuffer(bs))
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if tc.expectedStatusCode != resp.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, resp.StatusCode)
			}

			if tc.shouldCheckBody {
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

				respAnySlice, ok := respEnvelope["swap"]
				if !ok {
					t.Fatal(`the response does not contain enveloped swaps`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var swap *data.Swap

				err = json.NewDecoder(buf).Decode(&swap)
				if err != nil {
					t.Fatal(err)
				}

				if tc.input.RecipientInstrumentID != swap.RecipientInstrumentID {
					t.Errorf(`Expected RecipientInstrumentID %d, got %d`, tc.input.RecipientInstrumentID, swap.RecipientInstrumentID)
				}

				if tc.input.RequesterInstrumentID != swap.RequesterInstrumentID {
					t.Errorf(`Expected RequesterInstrumentID %d, got %d`, tc.input.RequesterInstrumentID, swap.RequesterInstrumentID)
				}

			}
		})
	}

}

// TestAcceptSwapHandler implements unit tests for acceptSwapHandler.
func TestAcceptSwapHandler(t *testing.T) {

	type testCase struct {
		name               string
		pathParam          string
		swap               []*data.Swap
		expectedStatusCode int
		shouldCheckBody    bool
	}

	testCases := []testCase{
		{
			name:      "happy path",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusOK,
			shouldCheckBody:    true,
		},
		{
			name:      "non exisitent swap id",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    2,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusNotFound,
			shouldCheckBody:    false,
		},
		{
			name:      "swap with accepted state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsAccepted:            true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckBody:    false,
		},
		{
			name:      "swap with rejected state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsRejected:            true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckBody:    false,
		},
		{
			name:      "swap with ended state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsEnded:               true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckBody:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Swaps: mocks.NewSwapModelMock(tc.swap)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /{id}", app.acceptSwapHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tc.pathParam)

			req, err := http.NewRequest("POST", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if tc.expectedStatusCode != resp.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, resp.StatusCode)
			}

			if tc.shouldCheckBody {

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

				respAnySlice, ok := respEnvelope["swap"]
				if !ok {
					t.Fatal(`the response does not contain enveloped swaps`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var swap *data.Swap

				err = json.NewDecoder(buf).Decode(&swap)
				if err != nil {
					t.Fatal(err)
				}

				if !swap.IsAccepted {
					t.Error("output swap should be accepted")
				}
			}

		})
	}

}

// TestRejectSwapHandler implements unit tests for rejectSwapHandler.
func TestRejectSwapHandler(t *testing.T) {

	type testCase struct {
		name               string
		pathParam          string
		swap               []*data.Swap
		expectedStatusCode int
		shouldCheckBody    bool
	}

	testCases := []testCase{
		{
			name:      "happy path",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusOK,
			shouldCheckBody:    true,
		},
		{
			name:      "non exisitent swap id",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    2,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusNotFound,
			shouldCheckBody:    false,
		},
		{
			name:      "swap with accepted state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsAccepted:            true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckBody:    false,
		},
		{
			name:      "swap with rejected state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsRejected:            true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckBody:    false,
		},
		{
			name:      "swap with ended state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsEnded:               true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckBody:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Swaps: mocks.NewSwapModelMock(tc.swap)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /{id}", app.rejectSwapHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tc.pathParam)

			req, err := http.NewRequest("POST", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if tc.expectedStatusCode != resp.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, resp.StatusCode)
			}

			if tc.shouldCheckBody {

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

				respAnySlice, ok := respEnvelope["swap"]
				if !ok {
					t.Fatal(`the response does not contain enveloped swaps`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var swap *data.Swap

				err = json.NewDecoder(buf).Decode(&swap)
				if err != nil {
					t.Fatal(err)
				}

				if !swap.IsRejected {
					t.Error("output swap should be rejected")
				}
				if !swap.IsEnded {
					t.Error("output swap should be ended")
				}
			}

		})
	}
}

// TestEndSwapHandler implements unit tests for endSwapHandler.
func TestEndSwapHandler(t *testing.T) {

	type testCase struct {
		name               string
		pathParam          string
		swap               []*data.Swap
		expectedStatusCode int
		shouldCheckBody    bool
	}

	testCases := []testCase{
		{
			name:      "happy path",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusOK,
			shouldCheckBody:    true,
		},
		{
			name:      "non exisitent swap id",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    2,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusNotFound,
			shouldCheckBody:    false,
		},
		{
			name:      "happy path - swap with accepted state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsAccepted:            true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusOK,
			shouldCheckBody:    false,
		},
		{
			name:      "happy path - swap with rejected state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsRejected:            true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusOK,
			shouldCheckBody:    false,
		},
		{
			name:      "happy path - swap with ended state",
			pathParam: "1",
			swap: []*data.Swap{{
				ID:                    1,
				CreatedAt:             time.Now().UTC(),
				RequesterInstrumentID: 1,
				IsEnded:               true,
				RecipientInstrumentID: 2,
				Version:               1,
			}},
			expectedStatusCode: http.StatusOK,
			shouldCheckBody:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Swaps: mocks.NewSwapModelMock(tc.swap)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /{id}", app.endSwapHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tc.pathParam)

			req, err := http.NewRequest("POST", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if tc.expectedStatusCode != resp.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, resp.StatusCode)
			}

			if tc.shouldCheckBody {

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

				respAnySlice, ok := respEnvelope["swap"]
				if !ok {
					t.Fatal(`the response does not contain enveloped swaps`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var swap *data.Swap

				err = json.NewDecoder(buf).Decode(&swap)
				if err != nil {
					t.Fatal(err)
				}

				if !swap.IsEnded {
					t.Error("output swap should be ended")
				}
			}

		})
	}
}
