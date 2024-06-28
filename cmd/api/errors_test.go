package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// TestErrorRespnses tests simple error-response handlers related helper functions.
func TestErrorRespnses(t *testing.T) {
	app := &application{}

	tests := []struct {
		name              string
		f                 func(http.ResponseWriter, *http.Request)
		expectedMsg       string
		expectedStausCode int
		expectedHeaders   map[string]string
	}{
		{
			name:              "notFoundResponse",
			f:                 app.notFoundResponse,
			expectedMsg:       "the requested resource could not be found",
			expectedStausCode: http.StatusNotFound,
		},
		{
			name:              "editConflictResponse",
			f:                 app.editConflictResponse,
			expectedMsg:       "unable to update the record due to an edit conflict, please try again",
			expectedStausCode: http.StatusConflict,
		},
		{
			name:              "rateLimitExcededResponse",
			f:                 app.rateLimitExcededResponse,
			expectedMsg:       "rate limit exceeded",
			expectedStausCode: http.StatusTooManyRequests,
		},
		{
			name:              "invalidCredentialsResponse",
			f:                 app.invalidCredentialsResponse,
			expectedMsg:       "invalid authentication credentials",
			expectedStausCode: http.StatusUnauthorized,
		},
		{
			name:              "invalidAuthenticationTokenResponse",
			f:                 app.invalidAuthenticationTokenResponse,
			expectedMsg:       "invalid or missing authentication token",
			expectedStausCode: http.StatusUnauthorized,
			expectedHeaders:   map[string]string{"WWW-Authenticate": "Bearer"},
		},
		{
			name:              "authenticationRequiredResponse",
			f:                 app.authenticationRequiredResponse,
			expectedMsg:       "you must be authenticated to access this resource",
			expectedStausCode: http.StatusUnauthorized,
		},
		{
			name:              "inactiveAccountResponse",
			f:                 app.inactiveAccountResponse,
			expectedMsg:       "your user account must be activated to access this resource",
			expectedStausCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			url := "https://www.example.com/path"

			r, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal("cannot set up request for testing")
			}

			rr := httptest.NewRecorder()

			tt.f(rr, r)
			resp := rr.Result()
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

			var jErrRes struct {
				Error string `json:"error"`
			}

			err = json.Unmarshal(body, &jErrRes)
			if err != nil {
				t.Fatal("cannot unmarshal json for text")
			}

			if jErrRes.Error != tt.expectedMsg {
				t.Errorf(`expected response body %#v, got %#v`, tt.expectedMsg, jErrRes.Error)
			}

			if resp.StatusCode != tt.expectedStausCode {
				t.Errorf(`expected status code %d, goit %d`, tt.expectedStausCode, resp.StatusCode)
			}

			if tt.expectedHeaders != nil {
				for hk, hv := range tt.expectedHeaders {
					if resp.Header.Get(hk) != hv {
						t.Errorf(`expected header walue "%v", got "%v"`, hv, resp.Header.Get(hk))
					}
				}
			}
		})
	}
}

// TestLogError tests the happy path of logError.
func TestLogError(t *testing.T) {

	testErr := errors.New("test error")

	expectedLevel := "ERROR"
	expextedMsg := testErr.Error()
	exPectedMethod := "GET"
	expectedURI := "/path"

	url := fmt.Sprintf("https://www.example.com%s", expectedURI)

	buf := &bytes.Buffer{}

	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))
	app := &application{logger: logger}

	r, err := http.NewRequest(exPectedMethod, url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	app.logError(r, testErr)

	var jRes struct {
		Time   time.Time `json:"time"`
		Level  string    `json:"level"`
		Msg    string    `json:"msg"`
		Method string    `json:"method"`
		URI    string    `json:"uri"`
	}

	err = json.Unmarshal(buf.Bytes(), &jRes)
	if err != nil {
		t.Fatal("cannot marshall data for testing")
	}

	if jRes.Level != expectedLevel {
		t.Errorf(`expected "%s", got "%s"`, expectedLevel, jRes.Level)
	}

	if jRes.Msg != expextedMsg {
		t.Errorf(`expected "%s", got "%s"`, expextedMsg, jRes.Msg)
	}

	if jRes.Method != exPectedMethod {
		t.Errorf(`expected "%s", got "%s"`, exPectedMethod, jRes.Method)
	}

	if jRes.URI != expectedURI {
		t.Errorf(`expected "%s", got "%s"`, expectedURI, jRes.URI)
	}

}

// TestErrorResponse is testing the happy path of errorResponse.
func TestErrorResponse(t *testing.T) {

	expextedErrorMsg := "test error message"
	expectedContentType := "application/json"

	app := &application{}

	r, err := http.NewRequest("GET", "https://www.example.com/path", nil)
	if err != nil {
		t.Fatal("cannot initiate a request for the test")
	}
	rr := httptest.NewRecorder()

	app.errorResponse(rr, r, http.StatusInternalServerError, expextedErrorMsg)

	resp := rr.Result()
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

	var jRes struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(body, &jRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf(`expected status code %d, got %d`, http.StatusInternalServerError, resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != expectedContentType {
		t.Errorf(`expected content type "%s", got "%s"`, expectedContentType, resp.Header.Get("Content-Type"))
	}

	if jRes.Error != expextedErrorMsg {
		t.Errorf(`expected errors message "%s", got "%s"`, expextedErrorMsg, jRes.Error)
	}

}

// TestServerErrorLogResponse tests the happy path of serverErrorLogResponse.
func TestServerErrorLogResponse(t *testing.T) {
	testErr := errors.New("test error")

	expectedLevel := "ERROR"
	expextedMsg := testErr.Error()
	exPectedMethod := "GET"
	expectedURI := "/path"

	expextedErrorMsg := "the server encountered a problem and could not process your request"
	expectedContentType := "application/json"

	url := fmt.Sprintf("https://www.example.com%s", expectedURI)

	buf := &bytes.Buffer{}

	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))
	app := &application{logger: logger}

	r, err := http.NewRequest(exPectedMethod, url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.serverErrorLogResponse(rr, r, testErr)

	var jRes struct {
		Time   time.Time `json:"time"`
		Level  string    `json:"level"`
		Msg    string    `json:"msg"`
		Method string    `json:"method"`
		URI    string    `json:"uri"`
	}

	err = json.Unmarshal(buf.Bytes(), &jRes)
	if err != nil {
		t.Fatal("cannot marshall data for testing")
	}

	if jRes.Level != expectedLevel {
		t.Errorf(`expected "%s", got "%s"`, expectedLevel, jRes.Level)
	}

	if jRes.Msg != expextedMsg {
		t.Errorf(`expected "%s", got "%s"`, expextedMsg, jRes.Msg)
	}

	if jRes.Method != exPectedMethod {
		t.Errorf(`expected "%s", got "%s"`, exPectedMethod, jRes.Method)
	}

	if jRes.URI != expectedURI {
		t.Errorf(`expected "%s", got "%s"`, expectedURI, jRes.URI)
	}

	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf(`expected status code %d, got %d`, http.StatusInternalServerError, resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != expectedContentType {
		t.Errorf(`expected content type "%s", got "%s"`, expectedContentType, resp.Header.Get("Content-Type"))
	}

	if jErrRes.Error != expextedErrorMsg {
		t.Errorf(`expected errors message "%s", got "%s"`, expextedErrorMsg, jErrRes.Error)
	}

}

// TestNotFoundResponse tests the happy path of notFoundResponse.
func TestNotFoundResponse(t *testing.T) {
	expectedMsg := "the requested resource could not be found"
	expectedStatusCode := http.StatusNotFound

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.notFoundResponse(rr, r)

	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if jErrRes.Error != expectedMsg {
		t.Errorf(`expected response body "%s", got "%s"`, expectedMsg, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, got %d`, expectedStatusCode, resp.StatusCode)
	}

}

// TestFailedValidationResponse tests the happy path of failedValidationResponse.
func TestFailedValidationResponse(t *testing.T) {

	expectedStatusCode := http.StatusUnprocessableEntity

	validationErrors := make(map[string]string)

	validationErrors["field1"] = "field1 has error"
	validationErrors["field2"] = "field2 has error"

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.failedValidationResponse(rr, r, validationErrors)

	resp := rr.Result()
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

	var jErrRes struct {
		Error map[string]string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if !reflect.DeepEqual(jErrRes.Error, validationErrors) {
		t.Errorf(`expected response body "%#v", got "%#v"`, validationErrors, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, goit %d`, expectedStatusCode, resp.StatusCode)
	}

}

// TestBadRequestResponse tests the happy path of badRequestResponse.
func TestBadRequestResponse(t *testing.T) {
	expectedStatusCode := http.StatusBadRequest
	expectesErrorMsg := "test error"

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.badRequestResponse(rr, r, errors.New(expectesErrorMsg))

	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if jErrRes.Error != expectesErrorMsg {
		t.Errorf(`expected response body "%#v", got "%#v"`, expectesErrorMsg, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, goit %d`, expectedStatusCode, resp.StatusCode)
	}
}

// TestEditConflictResponse tests the happy path for editConflictResponse.
func TestEditConflictResponse(t *testing.T) {
	expectesErrorMsg := "unable to update the record due to an edit conflict, please try again"
	expectedStatusCode := http.StatusConflict

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.editConflictResponse(rr, r)
	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if jErrRes.Error != expectesErrorMsg {
		t.Errorf(`expected response body "%#v", got "%#v"`, expectesErrorMsg, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, goit %d`, expectedStatusCode, resp.StatusCode)
	}
}

// TestRateLimitExcededResponse tests the happy path for rateLimitExcededResponse.
func TestRateLimitExcededResponse(t *testing.T) {
	expectesErrorMsg := "rate limit exceeded"
	expectedStatusCode := http.StatusTooManyRequests

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.rateLimitExcededResponse(rr, r)
	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if jErrRes.Error != expectesErrorMsg {
		t.Errorf(`expected response body "%#v", got "%#v"`, expectesErrorMsg, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, goit %d`, expectedStatusCode, resp.StatusCode)
	}
}

// TestInvalidCredentialsResponse tests the happy path for invalidCredentialsResponse.
func TestInvalidCredentialsResponse(t *testing.T) {
	expectesErrorMsg := "invalid authentication credentials"
	expectedStatusCode := http.StatusUnauthorized

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.invalidCredentialsResponse(rr, r)
	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if jErrRes.Error != expectesErrorMsg {
		t.Errorf(`expected response body "%#v", got "%#v"`, expectesErrorMsg, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, goit %d`, expectedStatusCode, resp.StatusCode)
	}
}

// TestInvalidAuthenticationTokenResponse tests the happy path for invalidAuthenticationTokenResponse.
func TestInvalidAuthenticationTokenResponse(t *testing.T) {
	expectesErrorMsg := "invalid or missing authentication token"
	expectedStatusCode := http.StatusUnauthorized
	expectedWWWAuthenticate := "Bearer"

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.invalidAuthenticationTokenResponse(rr, r)
	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if jErrRes.Error != expectesErrorMsg {
		t.Errorf(`expected response body "%#v", got "%#v"`, expectesErrorMsg, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, goit %d`, expectedStatusCode, resp.StatusCode)
	}

	if resp.Header.Get("WWW-Authenticate") != expectedWWWAuthenticate {
		t.Errorf(`expected WWW-Authenticate header "%s", got "%s"`, expectedWWWAuthenticate, resp.Header.Get("WWW-Authenticate"))
	}
}

// TestAuthenticationRequiredResponse tests the happy path for AuthenticationRequiredResponse.
func TestAuthenticationRequiredResponse(t *testing.T) {
	expectesErrorMsg := "you must be authenticated to access this resource"
	expectedStatusCode := http.StatusUnauthorized

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.authenticationRequiredResponse(rr, r)
	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if jErrRes.Error != expectesErrorMsg {
		t.Errorf(`expected response body "%#v", got "%#v"`, expectesErrorMsg, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, goit %d`, expectedStatusCode, resp.StatusCode)
	}
}

// TestInactiveAccountResponse tests the happy path for inactiveAccountResponse.
func TestInactiveAccountResponse(t *testing.T) {
	expectesErrorMsg := "your user account must be activated to access this resource"
	expectedStatusCode := http.StatusForbidden

	app := &application{}

	url := "https://www.example.com/path"

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("cannot set up request for testing")
	}

	rr := httptest.NewRecorder()

	app.inactiveAccountResponse(rr, r)
	resp := rr.Result()
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

	var jErrRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &jErrRes)
	if err != nil {
		t.Fatal("cannot unmarshal json for text")
	}

	if jErrRes.Error != expectesErrorMsg {
		t.Errorf(`expected response body "%#v", got "%#v"`, expectesErrorMsg, jErrRes.Error)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf(`expected status code %d, goit %d`, expectedStatusCode, resp.StatusCode)
	}
}
