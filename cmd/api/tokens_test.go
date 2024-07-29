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
	"testing"

	"github.com/ttarnok/instrument-swap-api/internal/auth"
	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/data/mocks"
)

// TestCreateAuthenticationTokenHandler implements unit tests for createAuthenticationTokenHandler.
func TestCreateAuthenticationTokenHandler(t *testing.T) {

	testUser := &data.User{
		ID:    1,
		Name:  "Dummy Username",
		Email: "test@example.com",
	}

	err := testUser.Password.Set("123qwe123qwe")
	if err != nil {
		t.Fatal(err)
	}

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type testCase struct {
		name               string
		users              []*data.User
		rb                 requestBody
		expectedStatusCode int
		shouldCheckBody    bool
	}

	testCases := []testCase{
		{
			name:               "happy path",
			users:              []*data.User{testUser},
			rb:                 requestBody{Email: "test@example.com", Password: "123qwe123qwe"},
			expectedStatusCode: http.StatusCreated,
			shouldCheckBody:    true,
		},
		{
			name:               "non valid password",
			users:              []*data.User{testUser},
			rb:                 requestBody{Email: "test@example.com", Password: "123"},
			expectedStatusCode: http.StatusUnprocessableEntity,
			shouldCheckBody:    false,
		},
		{
			name:               "non existent",
			users:              []*data.User{testUser},
			rb:                 requestBody{Email: "aaa@example.com", Password: "123qwe123qwe"},
			expectedStatusCode: http.StatusUnauthorized,
			shouldCheckBody:    false,
		},
		{
			name:               "not matching password",
			users:              []*data.User{testUser},
			rb:                 requestBody{Email: "test@example.com", Password: "123qwe123qwexxx"},
			expectedStatusCode: http.StatusUnauthorized,
			shouldCheckBody:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(tc.users)},
				auth:   auth.NewAuth("secretsecretsecret"),
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /", app.createAuthenticationTokenHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/", ts.URL)

			reqBody, err := json.Marshal(tc.rb)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("POST", path, bytes.NewBuffer(reqBody))
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

				respAnySlice, ok := respEnvelope["authentication_token"]
				if !ok {
					t.Fatal(`the response does not contain an enveloped authentication_token`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var authenticationToken string

				err = json.NewDecoder(buf).Decode(&authenticationToken)
				if err != nil {
					t.Fatal(err)
				}

				if len(authenticationToken) == 0 {
					t.Error(`should respont a non 0 length auth token`)
				}
			}

		})
	}

}
