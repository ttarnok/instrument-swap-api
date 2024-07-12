package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestRecoverPanic implement unit tests for recoverPanic middleware.
func TestRecoverPanic(t *testing.T) {

	var (
		expextedSatuscode        = http.StatusInternalServerError
		expectedConnectionHeader = "close"
	)

	app := application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	app.recoverPanic(next).ServeHTTP(rr, req)
	recRes := rr.Result()

	if expextedSatuscode != recRes.StatusCode {
		t.Errorf(`expected status code %d, got %d`, expextedSatuscode, recRes.StatusCode)
	}

	if expectedConnectionHeader != recRes.Header.Get("Connection") {
		t.Errorf(`expected Connection header value %s, got %s`, expectedConnectionHeader, recRes.Header.Get("Connection"))
	}

}

// TestRateLimit tests the functionality of rateLimit middleware.
// The test cases run parallel. Has long runtime.
func TestRateLimit(t *testing.T) {

	type testCase struct {
		name                       string
		limiterCfgEnabled          bool
		limiterCfgRequestPerSecond float64
		limiterCfgBurst            int
		expectedStatusCode         int
		requestCount               int
		requestSleepDuration       time.Duration
	}

	testCases := []testCase{
		{
			name:                       "limiter disabled",
			limiterCfgEnabled:          false,
			limiterCfgBurst:            0,
			limiterCfgRequestPerSecond: 0,
			expectedStatusCode:         http.StatusOK,
			requestCount:               1,
			requestSleepDuration:       0,
		},
		{
			name:                       "limiter enabled",
			limiterCfgEnabled:          true,
			limiterCfgBurst:            4,
			limiterCfgRequestPerSecond: 2,
			expectedStatusCode:         http.StatusOK,
			requestCount:               1,
			requestSleepDuration:       0,
		},
		{
			name:                       "limiter enabled - multiple requests#1",
			limiterCfgEnabled:          true,
			limiterCfgBurst:            10,
			limiterCfgRequestPerSecond: 2,
			expectedStatusCode:         http.StatusOK,
			requestCount:               10,
			requestSleepDuration:       0,
		},
		{
			name:                       "limiter enabled - multiple requests#2",
			limiterCfgEnabled:          true,
			limiterCfgBurst:            1,
			limiterCfgRequestPerSecond: 2,
			expectedStatusCode:         http.StatusOK,
			requestCount:               2,
			requestSleepDuration:       500 * time.Millisecond,
		},
		{
			name:                       "limiter enabled - multiple requests#3",
			limiterCfgEnabled:          true,
			limiterCfgBurst:            4,
			limiterCfgRequestPerSecond: 1,
			expectedStatusCode:         http.StatusOK,
			requestCount:               5,
			requestSleepDuration:       250 * time.Millisecond,
		},
		{
			name:                       "limiter enabled - multiple too many requests#1",
			limiterCfgEnabled:          true,
			limiterCfgBurst:            2,
			limiterCfgRequestPerSecond: 1,
			expectedStatusCode:         http.StatusTooManyRequests,
			requestCount:               10,
			requestSleepDuration:       time.Millisecond,
		},
		{
			name:                       "limiter enabled - multiple too many requests#2",
			limiterCfgEnabled:          true,
			limiterCfgBurst:            8,
			limiterCfgRequestPerSecond: 1,
			expectedStatusCode:         http.StatusTooManyRequests,
			requestCount:               10,
			requestSleepDuration:       time.Millisecond,
		},
		{
			name:                       "limiter enabled - multiple too many requests#3",
			limiterCfgEnabled:          true,
			limiterCfgBurst:            2,
			limiterCfgRequestPerSecond: 1,
			expectedStatusCode:         http.StatusTooManyRequests,
			requestCount:               10,
			requestSleepDuration:       time.Millisecond,
		},
		{
			name:                       "limiter enabled - multiple too many requests#4",
			limiterCfgEnabled:          true,
			limiterCfgBurst:            4,
			limiterCfgRequestPerSecond: 2,
			expectedStatusCode:         http.StatusTooManyRequests,
			requestCount:               10,
			requestSleepDuration:       250 * time.Millisecond,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cfg config

			cfg.limiter.enabled = tc.limiterCfgEnabled
			cfg.limiter.burst = tc.limiterCfgBurst
			cfg.limiter.requestPerSecond = tc.limiterCfgRequestPerSecond

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				config: cfg,
			}

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.RemoteAddr = "123.123.123.123:1234"

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write([]byte("OK"))
				if err != nil {
					t.Fatal(err)
				}
			})

			testInst := app.rateLimit(next)

			var recRes *http.Response

			for range tc.requestCount {
				rr := httptest.NewRecorder()

				testInst.ServeHTTP(rr, req)
				recRes = rr.Result()
				time.Sleep(time.Duration(tc.requestSleepDuration))
			}

			if tc.expectedStatusCode != recRes.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, recRes.StatusCode)
			}

		})
	}

}
