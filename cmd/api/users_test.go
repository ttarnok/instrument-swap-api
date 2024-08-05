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

// compareUsers compares two users, if the fields are identical returns true.
// compareUsers does not consider the Password field.
func compareUsers(u1 *data.User, u2 *data.User) bool {

	if u1.ID != u2.ID {
		return false
	}

	if u1.Name != u2.Name {
		return false
	}

	if u1.Email != u2.Email {
		return false
	}

	if u1.Version != u2.Version {
		return false
	}

	if u1.CreatedAt != u2.CreatedAt {
		return false
	}

	if u1.Activated != u2.Activated {
		return false
	}

	return true
}

// TestUpdateUserHandler implement unti tests for updateUserHandler.
func TestUpdateUserHandler(t *testing.T) {

	testUser := data.User{
		ID:    1,
		Name:  "Dummy Username",
		Email: "test@example.com",
	}

	err := testUser.Password.Set("asd123asd123")
	if err != nil {
		t.Fatal(err)
	}

	type inputBody struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type testCase struct {
		name               string
		users              []*data.User
		pathParam          int
		input              inputBody
		expectedStatusCode int
		expectedUser       *data.User
	}

	testCases := []testCase{
		{
			name:               "happy path",
			users:              []*data.User{&testUser},
			pathParam:          1,
			input:              inputBody{Name: testUser.Name, Email: testUser.Email},
			expectedStatusCode: http.StatusOK,
			expectedUser:       &testUser,
		},
		{
			name:               "invalid path param",
			users:              []*data.User{&testUser},
			pathParam:          0,
			input:              inputBody{Name: testUser.Name, Email: testUser.Email},
			expectedStatusCode: http.StatusNotFound,
			expectedUser:       nil,
		},
		{
			name:               "non existent user",
			users:              []*data.User{&testUser},
			pathParam:          11,
			input:              inputBody{Name: testUser.Name, Email: testUser.Email},
			expectedStatusCode: http.StatusNotFound,
			expectedUser:       nil,
		},
		{
			name:               "non valid name",
			users:              []*data.User{&testUser},
			pathParam:          1,
			input:              inputBody{Name: "", Email: testUser.Email},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedUser:       nil,
		},
		{
			name:               "non valid email",
			users:              []*data.User{&testUser},
			pathParam:          1,
			input:              inputBody{Name: testUser.Name, Email: ""},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedUser:       nil,
		},
		{
			name:               "perform update",
			users:              []*data.User{&testUser},
			pathParam:          1,
			input:              inputBody{Name: "NewName", Email: "NewEmail@example.com"},
			expectedStatusCode: http.StatusOK,
			expectedUser: &data.User{
				ID:    1,
				Name:  "NewName",
				Email: "NewEmail@example.com",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(tc.users)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /{id}", app.updateUserHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, strconv.Itoa(tc.pathParam))

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

			if tc.expectedUser != nil {

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

				respAnySlice, ok := respEnvelope["user"]
				if !ok {
					t.Fatal(`the response does not contain an enveloped user`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var user *data.User

				err = json.NewDecoder(buf).Decode(&user)
				if err != nil {
					t.Fatal(err)
				}

				if !compareUsers(tc.expectedUser, user) {
					t.Errorf("expected user \n%#v, got \n%#v", tc.expectedUser, user)
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

// TestRegisterUserHandler unit tests registerUserHandler.
func TestRegisterUserHandler(t *testing.T) {

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

	type inputType struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type testCase struct {
		name               string
		input              inputType
		expectedStatusCode int
		shouldCheckBody    bool
	}

	testCases := []testCase{
		{
			name: "happy path",
			input: inputType{
				Name:     "Test User",
				Email:    "test@email.com",
				Password: "asd123asd123",
			},
			expectedStatusCode: http.StatusCreated,
			shouldCheckBody:    true,
		},
		{
			name: "empty password",
			input: inputType{
				Name:     "Test User",
				Email:    "test@email.com",
				Password: "",
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			shouldCheckBody:    false,
		},
		{
			name: "empty email",
			input: inputType{
				Name:     "Test User",
				Email:    "",
				Password: "asd123asd123",
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			shouldCheckBody:    false,
		},
		{
			name: "empty username",
			input: inputType{
				Name:     "",
				Email:    "test@email.com",
				Password: "asd123asd123",
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			shouldCheckBody:    false,
		},
		{
			name:               "empty input",
			input:              inputType{},
			expectedStatusCode: http.StatusUnprocessableEntity,
			shouldCheckBody:    false,
		},
		{
			name: "already existing email",
			input: inputType{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "asd123asd123",
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			shouldCheckBody:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(testUsers)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /test", app.registerUserHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/test", ts.URL)

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

				respAnySlice, ok := respEnvelope["user"]
				if !ok {
					t.Fatal(`the response does not contain an enveloped user`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAnySlice)
				if err != nil {
					t.Fatal(err)
				}

				var user *data.User

				err = json.NewDecoder(buf).Decode(&user)
				if err != nil {
					t.Fatal(err)
				}

				if tc.input.Name != user.Name {
					t.Errorf(`Expected user name "%s", got "%s"`, tc.input.Name, user.Name)
				}

				if tc.input.Email != user.Email {
					t.Errorf(`Expected user email "%s", got "%s"`, tc.input.Email, user.Email)
				}

				user, err = app.models.Users.GetByEmail(tc.input.Email)
				if err != nil {
					t.Fatal(err)
				}

				isPassMatch, err := user.Password.Matches(tc.input.Password)
				if err != nil {
					t.Fatal(err)
				}

				if !isPassMatch {
					t.Errorf(`expected matching passwords`)
				}

			}

		})
	}

}

// TestUpdatePasswordHandler implements unit tests for updatePasswordHandler.
func TestUpdatePasswordHandler(t *testing.T) {

	testUser := data.User{
		ID:    1,
		Name:  "Dummy Username",
		Email: "test@example.com",
	}

	err := testUser.Password.Set("asd123asd123")
	if err != nil {
		t.Fatal(err)
	}

	type inputType struct {
		Password    string `json:"password"`
		NewPassword string `json:"new_password"`
	}

	type testCase struct {
		name               string
		inputPath          string
		input              inputType
		user               data.User
		expectedStatusCode int
		shouldCheckModel   bool
	}

	testCases := []testCase{
		{
			name:      "happy path",
			inputPath: "1",
			input: inputType{
				Password:    "asd123asd123",
				NewPassword: "123qwe123qwe",
			},
			user:               testUser,
			expectedStatusCode: http.StatusOK,
			shouldCheckModel:   true,
		},
		{
			name:      "happy path2",
			inputPath: "1",
			input: inputType{
				Password:    "asd123asd123",
				NewPassword: "newpass1111",
			},
			user:               testUser,
			expectedStatusCode: http.StatusOK,
			shouldCheckModel:   true,
		},
		{
			name:      "invalid url path",
			inputPath: "Nan",
			input: inputType{
				Password:    "asd123asd123",
				NewPassword: "newpass1111",
			},
			user:               testUser,
			expectedStatusCode: http.StatusNotFound,
			shouldCheckModel:   false,
		},
		{
			name:      "non exestent url path",
			inputPath: "111",
			input: inputType{
				Password:    "asd123asd123",
				NewPassword: "newpass1111",
			},
			user:               testUser,
			expectedStatusCode: http.StatusNotFound,
			shouldCheckModel:   false,
		},
		{
			name:      "non matching pass",
			inputPath: "1",
			input: inputType{
				Password:    "asd123as23232323",
				NewPassword: "newpass1111",
			},
			user:               testUser,
			expectedStatusCode: http.StatusBadRequest,
			shouldCheckModel:   false,
		},
		{
			name:      "not valid new pass",
			inputPath: "1",
			input: inputType{
				Password:    "asd123asd123",
				NewPassword: "new",
			},
			user:               testUser,
			expectedStatusCode: http.StatusUnprocessableEntity,
			shouldCheckModel:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock([]*data.User{&tc.user})},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /{id}", app.updatePasswordHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tc.inputPath)

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

			if tc.shouldCheckModel {
				id, err := strconv.Atoi(tc.inputPath)
				if err != nil {
					t.Fatal(err)
				}

				user, err := app.models.Users.GetByID(int64(id))
				if err != nil {
					t.Fatal(err)
				}

				isPassMatch, err := user.Password.Matches(tc.input.NewPassword)
				if err != nil {
					t.Fatal(err)
				}

				if !isPassMatch {
					t.Error(`Updated password should match the NewPassword`)
				}
			}
		})
	}
}
