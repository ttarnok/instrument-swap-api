package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
)

// TestExtractIDParam implements unit tests to test extractIDParam.
func TestExtractIDParam(t *testing.T) {

	tests := []struct {
		name            string
		testPath        string
		expectedIDParam int64
		expectedError   bool
	}{
		{
			name:            "happy path",
			testPath:        "13",
			expectedIDParam: 13,
			expectedError:   false,
		},
		{
			name:            "non numeric ID value",
			testPath:        "nonnum",
			expectedIDParam: 0,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.Handle("GET /{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				app := &application{}

				idParam, err := app.extractIDParam(r)
				if err != nil && !tt.expectedError {
					t.Fatal("not expected error", err)
				}
				if err == nil && tt.expectedError {
					t.Errorf("expected error, got nill, %d", idParam)
				}

				if idParam != tt.expectedIDParam {
					t.Errorf("expected IDParam %d, got %d", tt.expectedIDParam, idParam)
				}

			}))

			ts := httptest.NewServer(mux)
			defer ts.Close()

			testURL := fmt.Sprintf("%s/%s", ts.URL, tt.testPath)

			req, err := http.NewRequest(http.MethodGet, testURL, nil)
			if err != nil {
				t.Fatal(err)
			}

			_, err = http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
		})

	}

}

// TestWriteJSON tests the functionality of writeJSON.
func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		data          envelope
		headers       http.Header
		expectedError bool
	}{
		{
			name:          "happy path",
			statusCode:    http.StatusOK,
			data:          envelope{"data": "message"},
			headers:       map[string][]string{"Vary": {"Value1", "Value2"}},
			expectedError: false,
		},
		{
			name:       "status not found",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rr := httptest.NewRecorder()

			app := application{}

			err := app.writeJSON(rr, tt.statusCode, tt.data, tt.headers)
			// Error check.
			if err != nil && !tt.expectedError {
				t.Errorf(`not expected error, got %#v`, err)
			}
			if err == nil && tt.expectedError {
				t.Error("expected error")
			}

			resp := rr.Result()

			// Status Code check.
			if resp.StatusCode != tt.statusCode {
				t.Errorf(`expected status code %d, got %d`, tt.statusCode, resp.StatusCode)
			}

			// Body check.
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

			mapBody := make(envelope)
			err = json.Unmarshal(body, &mapBody)
			if err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(tt.data, mapBody) {
				t.Errorf(`expected message body %#v, got %#v`, tt.data, mapBody)
			}

			// Headers check.
			for k, v := range tt.headers {
				header := resp.Header.Values(k)
				if !slices.Equal(header, v) {
					t.Errorf(`expected header %#v, got %#v`, v, header)
				}

			}
			// Must contain: "Content-Type", "application/json".
			if resp.Header.Get("Content-Type") != "application/json" {
				t.Error(`the response must contain "Content-Type": "application/json" header`)
			}
		})
	}
}

// TestReadJSON implements unit tests for readJSON.
func TestReadJSON(t *testing.T) {

	tests := []struct {
		name   string
		source map[string]string
		target map[string]string
	}{
		{
			name:   "happy path",
			source: map[string]string{"key1": "value1", "key2": "value2"},
			target: make(map[string]string),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rr := httptest.NewRecorder()

			bs, err := json.Marshal(tt.source)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("GET", "/", bytes.NewBuffer(bs))

			app := application{}

			err = app.readJSON(rr, req, &tt.target)
			if err != nil {
				t.Error(err)
			}

			if !maps.Equal(tt.source, tt.target) {
				t.Errorf(`expected value %#v, got %#v`, tt.source, tt.target)
			}

		})
	}
}

// TestReadQParamString tests the functionality of readQParamString.
func TestReadQParamString(t *testing.T) {

	tests := []struct {
		name         string
		setKey       string
		retrieveKey  string
		value        string
		defaultValue string
	}{
		{
			name:         "happy path",
			setKey:       "name",
			retrieveKey:  "name",
			value:        "Ava",
			defaultValue: "N/A",
		},
		{
			name:         "default value",
			setKey:       "name",
			retrieveKey:  "name2",
			value:        "Ava",
			defaultValue: "N/A",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			v := url.Values{}
			v.Set(tt.setKey, "Ava")

			app := &application{}

			paramValue := app.readQParamString(v, tt.retrieveKey, tt.defaultValue)

			if tt.setKey == tt.retrieveKey && paramValue != tt.value {
				t.Errorf(`expected value "%s", got "%s"`, tt.value, paramValue)
			}
			if tt.setKey != tt.retrieveKey && paramValue != tt.defaultValue {
				t.Errorf(`expected default value "%s", got "%s"`, tt.defaultValue, paramValue)
			}

		})

	}
}
