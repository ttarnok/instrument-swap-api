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

	expextedErrorMessage := "test error message"
	expectedContentType := "application/json"

	app := &application{}

	r, err := http.NewRequest("GET", "https://www.example.com/path", nil)
	if err != nil {
		t.Fatal("cannot initiate a request for the test")
	}
	w := httptest.NewRecorder()

	app.errorResponse(w, r, http.StatusInternalServerError, expextedErrorMessage)

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

	if jRes.Error != expextedErrorMessage {
		t.Errorf(`expected errors message "%s", got "%s"`, expextedErrorMessage, jRes.Error)
	}

}
