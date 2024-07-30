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
	"strconv"
	"testing"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/data/mocks"
)

// TestListUsersHandler implements unit tests for listUsersHandler.
func TestListUsersHandler(t *testing.T) {

	testUsers := []*data.User{
		{
			ID:    1,
			Name:  "Dummy Username",
			Email: "test@example.com",
		},
		{
			ID:    2,
			Name:  "Other Temp User",
			Email: "temp@example.com",
		},
	}

	type testCase struct {
		name               string
		users              []*data.User
		expectedStatusCode int
		shouldCheckBody    bool
	}

	testCases := []testCase{
		{
			name:               "happy path",
			users:              testUsers,
			expectedStatusCode: http.StatusOK,
			shouldCheckBody:    true,
		},
		{
			name:               "empty users",
			users:              []*data.User{},
			expectedStatusCode: http.StatusOK,
			shouldCheckBody:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(tc.users)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("GET /", app.listUsersHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/", ts.URL)

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

				respAnySlice, ok := respEnvelope["users"]
				if !ok {
					t.Fatal(`the response does not contain enveloped users`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var users []*data.User

				err = json.NewDecoder(buf).Decode(&users)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tc.users, users) {
					t.Errorf(`expected users \n%#v, got \n%#v`, tc.users, users)
				}

			}
		})
	}

}

// TestDeleteUserHandler implements unit tests for deleteUserHandler.
func TestDeleteUserHandler(t *testing.T) {

	testUsers := []*data.User{
		{
			ID:    1,
			Name:  "Dummy Username",
			Email: "test@example.com",
		},
		{
			ID:    2,
			Name:  "Other Temp User",
			Email: "temp@example.com",
		},
	}

	type testCase struct {
		name               string
		users              []*data.User
		pathParam          int
		expectedStatusCode int
	}

	testCases := []testCase{
		{
			name:               "happy path",
			users:              testUsers,
			pathParam:          1,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "non existent user path",
			users:              testUsers,
			pathParam:          22,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "no users",
			users:              []*data.User{},
			pathParam:          22,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "invalid param",
			users:              testUsers,
			pathParam:          0,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(tc.users)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /{id}", app.deleteUserHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, strconv.Itoa(tc.pathParam))

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

			allUsers, err := app.models.Users.GetAll()
			if err != nil {
				t.Fatal(err)
			}

			for _, user := range allUsers {
				if user.ID == int64(tc.pathParam) {
					t.Errorf(`the user with id(T%d) should be deleted from the model`, tc.pathParam)
				}
			}

		})
	}
}
