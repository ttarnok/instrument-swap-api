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
