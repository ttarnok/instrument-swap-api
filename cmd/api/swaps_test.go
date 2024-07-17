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
