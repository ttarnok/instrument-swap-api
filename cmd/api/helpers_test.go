package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
