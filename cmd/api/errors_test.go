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
	w := httptest.NewRecorder()

	app.errorResponse(w, r, http.StatusInternalServerError, expextedErrorMsg)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

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

	w := httptest.NewRecorder()

	app.serverErrorLogResponse(w, r, testErr)

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

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

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

	w := httptest.NewRecorder()

	app.notFoundResponse(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

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

	w := httptest.NewRecorder()

	app.failedValidationResponse(w, r, validationErrors)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

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
