package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/pascaldekloe/jwt"
	"github.com/ttarnok/instrument-swap-api/internal/auth"
	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/data/mocks"
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

// TestAuthenticate implements unit testes for authenticate middleware.
func TestAuthenticate(t *testing.T) {

	testSecret := "secret"

	var claims jwt.Claims
	claims.Subject = strconv.FormatInt(1, 10)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "instrument-swap.example.example"
	claims.Audiences = []string{"instrument-swap.example.example"}

	validJWTBytes, err := claims.HMACSign(jwt.HS256, []byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}

	claims.Subject = strconv.FormatInt(1, 10)
	claims.Issued = jwt.NewNumericTime(time.Now().Add(-48 * time.Hour))
	claims.NotBefore = jwt.NewNumericTime(time.Now().Add(-48 * time.Hour))
	claims.Expires = jwt.NewNumericTime(time.Now().Add(-48 * time.Hour).Add(24 * time.Hour))
	claims.Issuer = "instrument-swap.example.example"
	claims.Audiences = []string{"instrument-swap.example.example"}
	expiredJWTBytes, err := claims.HMACSign(jwt.HS256, []byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}

	claims.Subject = "non number"
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "instrument-swap.example.example"
	claims.Audiences = []string{"instrument-swap.example.example"}

	nonNumbericSubJWTBytes, err := claims.HMACSign(jwt.HS256, []byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		name               string
		token              string
		expectedStatusCode int
		expectedUser       *data.User
	}

	testCases := []testCase{
		{
			name:               "valid token",
			token:              "Bearer " + string(validJWTBytes),
			expectedStatusCode: http.StatusOK,
			expectedUser:       &data.User{ID: 1, Name: "Test User"},
		},
		{
			name:               "without token",
			token:              "",
			expectedStatusCode: http.StatusOK,
			expectedUser:       data.AnonymousUser,
		},
		{
			name:               "broken token value",
			token:              "asfsgqweg",
			expectedStatusCode: http.StatusUnauthorized,
			expectedUser:       data.AnonymousUser,
		},
		{
			name:               "broken bearer token value",
			token:              "Bearer asfsgqweg",
			expectedStatusCode: http.StatusUnauthorized,
			expectedUser:       data.AnonymousUser,
		},
		{
			name:               "invalid token",
			token:              "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedStatusCode: http.StatusUnauthorized,
			expectedUser:       data.AnonymousUser,
		},
		{
			name:               "expired token",
			token:              "Bearer " + string(expiredJWTBytes),
			expectedStatusCode: http.StatusUnauthorized,
			expectedUser:       data.AnonymousUser,
		},
		{
			name:               "nun numeric subject",
			token:              "Bearer " + string(nonNumbericSubJWTBytes),
			expectedStatusCode: http.StatusInternalServerError,
			expectedUser:       data.AnonymousUser,
		},
		{
			name:               "non existent user",
			token:              "Bearer " + string(validJWTBytes),
			expectedStatusCode: http.StatusUnauthorized,
			expectedUser:       &data.User{ID: 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var app *application
			if tc.expectedUser == data.AnonymousUser {
				app = &application{
					models: data.Models{
						Users: mocks.NewEmptyUserModelMock(),
					},
					auth:   auth.NewAuth(testSecret),
					logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				}
			} else {
				app = &application{
					models: data.Models{
						Users: mocks.NewUserModelMock(map[int64]*data.User{tc.expectedUser.ID: tc.expectedUser}),
					},
					auth:   auth.NewAuth(testSecret),
					logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				}
			}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tc.token != "" {
				req.Header.Add("Authorization", tc.token)
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				if !reflect.DeepEqual(tc.expectedUser, app.contextGetUser(r)) {
					t.Errorf(`expected user %#v, got %#v`, tc.expectedUser, app.contextGetUser(r))
				}

				_, err := w.Write([]byte("OK"))
				if err != nil {
					t.Fatal(err)
				}
			})

			rr := httptest.NewRecorder()

			app.authenticate(next).ServeHTTP(rr, req)
			recRes := rr.Result()

			if tc.expectedStatusCode != recRes.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, recRes.StatusCode)
			}

			if recRes.Header.Get("Vary") != "Authorization" {
				t.Errorf(`response shoud contain "Vary" header with the value of "Authorization", got value: "%s"`, recRes.Header.Get("Vary"))
			}

		})
	}
}

// TestEnableCORS implements unit tests for enableCORS middleware.
func TestEnableCORS(t *testing.T) {

	type testCase struct {
		name                       string
		requestMethod              string
		sendACRMHeader             bool
		expectedStatusCode         int
		shouldContainOptionsHeader bool
	}

	testCases := []testCase{
		{
			name:                       "with options request, with ACMR",
			requestMethod:              http.MethodOptions,
			sendACRMHeader:             true,
			expectedStatusCode:         http.StatusOK,
			shouldContainOptionsHeader: true,
		},
		{
			name:                       "without options request, without ACMR",
			requestMethod:              http.MethodGet,
			sendACRMHeader:             false,
			expectedStatusCode:         http.StatusOK,
			shouldContainOptionsHeader: false,
		},
		{
			name:                       "without options request, with ACMR",
			requestMethod:              http.MethodGet,
			sendACRMHeader:             true,
			expectedStatusCode:         http.StatusOK,
			shouldContainOptionsHeader: false,
		},
		{
			name:                       "with options request, without ACMR",
			requestMethod:              http.MethodOptions,
			sendACRMHeader:             false,
			expectedStatusCode:         http.StatusOK,
			shouldContainOptionsHeader: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &application{}

			req, err := http.NewRequest(tc.requestMethod, "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tc.sendACRMHeader {
				req.Header.Set("Access-Control-Request-Method", "GET")
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write([]byte("OK"))
				if err != nil {
					t.Fatal(err)
				}
			})

			rr := httptest.NewRecorder()

			app.enableCORS(next).ServeHTTP(rr, req)
			recRes := rr.Result()

			if tc.expectedStatusCode != recRes.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, recRes.StatusCode)
			}

			if recRes.Header.Get("Vary") != "Access-Control-Request-Method" {
				t.Errorf(`response shoud contain "Vary" header with the value of "Access-Control-Request-Method", got value: "%s"`, recRes.Header.Get("Vary"))
			}

			if recRes.Header.Get("Access-Control-Allow-Origin") != "*" {
				t.Errorf(`response shoud contain "Access-Control-Allow-Origin" header with the value of "*", got value: "%s"`, recRes.Header.Get("Access-Control-Allow-Origin"))
			}

			if tc.shouldContainOptionsHeader {
				if recRes.Header.Get("Access-Control-Allow-Methods") != "OPTIONS, PUT, PATCH, DELETE" {
					t.Errorf(`response shoud contain "Access-Control-Allow-Methods" header with the value of "OPTIONS, PUT, PATCH, DELETE" got value: "%s"`, recRes.Header.Get("Access-Control-Allow-Methods"))
				}
				if recRes.Header.Get("Access-Control-Allow-Headers") != "Authorization, Content-Type" {
					t.Errorf(`response shoud contain "Access-Control-Allow-Headers" header with the value of "Authorization, Content-Type" got value: "%s"`, recRes.Header.Get("Access-Control-Allow-Headers"))
				}
			}

		})
	}

}
