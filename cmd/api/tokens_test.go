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
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pascaldekloe/jwt"
	"github.com/ttarnok/instrument-swap-api/internal/auth"
	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/data/mocks"
)

// TestLoginHandler implements unit tests for loginHandler.
func TestLoginHandler(t *testing.T) {

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
			name:               "blacklisted refresh token",
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
			t.Parallel()

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(tc.users)},
				auth:   auth.NewAuth("secretsecretsecret", mocks.NewBlacklistServiceMock()),
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /", app.loginHandler)

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

				respAccessToken, ok := respEnvelope["access"]
				if !ok {
					t.Fatal(`the response does not contain an enveloped access token`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAccessToken)
				if err != nil {
					t.Fatal(err)
				}

				var accessToken string

				err = json.NewDecoder(buf).Decode(&accessToken)
				if err != nil {
					t.Fatal(err)
				}

				if len(accessToken) == 0 {
					t.Error(`should respont a non 0 length access token`)
				}

				_, err = app.auth.AccessToken.ParseClaims([]byte(accessToken))
				if err != nil {
					t.Errorf(`access token do not parse due to: "%s"`, err.Error())
				}

				if !app.auth.AccessToken.IsValid([]byte(accessToken)) {
					t.Error(`access token shoul be valid`)
				}

				respRefreshToken, ok := respEnvelope["refresh"]
				if !ok {
					t.Fatal(`the response does not contain an enveloped refresh token`)
				}
				buf = new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respRefreshToken)
				if err != nil {
					t.Fatal(err)
				}

				var refreshToken string

				err = json.NewDecoder(buf).Decode(&refreshToken)
				if err != nil {
					t.Fatal(err)
				}

				if len(refreshToken) == 0 {
					t.Error(`should respont a non 0 length refresh token`)
				}

				_, err = app.auth.RefreshToken.ParseClaims([]byte(refreshToken))
				if err != nil {
					t.Errorf(`refresh token do not parse due to: "%s"`, err.Error())
				}

				if !app.auth.RefreshToken.IsValid([]byte(refreshToken)) {
					t.Error("refresh token shoul be valid")
				}
			}

		})
	}

}

// GenerateTestToken generates a jwt token for testing purposes.
func GenerateTestToken(secret string, id string, subject int64, issued time.Time, notBefore time.Time, expires time.Time, issuer string, tokenType string) []byte {

	var claims jwt.Claims
	claims.ID = id
	claims.Subject = strconv.FormatInt(subject, 10)
	claims.Issued = jwt.NewNumericTime(issued)
	claims.NotBefore = jwt.NewNumericTime(notBefore)
	claims.Expires = jwt.NewNumericTime(expires)
	claims.Issuer = issuer
	claims.Audiences = []string{"instrument-swap.example.example"}
	claims.Set = map[string]interface{}{"token_type": tokenType}

	JWTBytes, err := claims.HMACSign(jwt.HS256, []byte(secret))
	if err != nil {
		panic("test token generation error")
	}

	return JWTBytes
}

// TestRefreshHandler unit tests refreshHandler.
func TestRefreshHandler(t *testing.T) {

	testUser := &data.User{
		ID:    1,
		Name:  "Dummy Username",
		Email: "test@example.com",
	}

	testSecret := "secretsecretsecret"

	testTokenID1 := uuid.NewString()
	testTokenID2 := uuid.NewString()
	// testTokenID3 := uuid.NewString()
	// testTokenID4 := uuid.NewString()

	type testCase struct {
		name                 string
		users                []*data.User
		accessToken          []byte
		refreshToken         []byte
		expectedStatusCode   int
		shouldValidateResult bool
		expectedUserID       int64
		blacklist            []string
	}

	testCases := []testCase{
		{
			name:                 "happy path",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusCreated,
			shouldValidateResult: true,
			expectedUserID:       testUser.ID,
			blacklist:            nil,
		},
		{
			name:                 "blacklisted refresh token",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            []string{testTokenID2},
		},
		{
			name:                 "blacklisted access token",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            []string{testTokenID1},
		},
		{
			name:                 "non existent user id",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, 2, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, 2, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            nil,
		},
		{
			name:                 "wrong secret",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken("wring secret", testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:         GenerateTestToken("wrong secret", testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            nil,
		},
		{
			name:                 "not expired access token",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now(), time.Now(), time.Now().Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            nil,
		},
		{
			name:                 "expired refresh token",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            nil,
		},
		{
			name:                 "wrong access token type",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "accessxxx"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            nil,
		},
		{
			name:                 "no access token",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "refresh"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            nil,
		},
		{
			name:                 "not matching user ids",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, 2, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            nil,
		},
		{
			name:                 "wrong issuer",
			users:                []*data.User{testUser},
			accessToken:          GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(5*time.Minute), "xx", "access"),
			refreshToken:         GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode:   http.StatusUnauthorized,
			shouldValidateResult: false,
			expectedUserID:       0,
			blacklist:            nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(tc.users)},
				auth:   auth.NewAuth(testSecret, mocks.NewBlacklistServiceMockWithData(tc.blacklist)),
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /", app.refreshHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/", ts.URL)

			type input struct {
				AccessToken  string `json:"access"`
				RefreshToken string `json:"refresh"`
			}

			reqBody, err := json.Marshal(input{AccessToken: string(tc.accessToken), RefreshToken: string(tc.refreshToken)})
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

			if tc.shouldValidateResult {

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

				respAccessToken, ok := respEnvelope["access"]
				if !ok {
					t.Fatal(`the response does not contain an enveloped access token`)
				}
				buf := new(bytes.Buffer)
				err = json.NewEncoder(buf).Encode(respAccessToken)
				if err != nil {
					t.Fatal(err)
				}

				var accessToken string

				err = json.NewDecoder(buf).Decode(&accessToken)
				if err != nil {
					t.Fatal(err)
				}

				if len(accessToken) == 0 {
					t.Error(`should respont a non 0 length access token`)
				}

				claims, err := app.auth.AccessToken.ParseClaims([]byte(accessToken))
				if err != nil {
					t.Errorf(`access token do not parse due to: "%s"`, err.Error())
				}

				if !app.auth.AccessToken.IsValid([]byte(accessToken)) {
					t.Error(`access token shoul be valid`)
				}

				userID, err := strconv.ParseInt(claims.Subject, 10, 64)
				if err != nil {
					t.Fatal(err)
				}
				if tc.expectedUserID != userID {
					t.Errorf(`Expected user id claim %d, got %d`, tc.expectedUserID, userID)
				}

			}
		})
	}
}

// TestBlacklistHandler unit tests blacklistHandler.
func TestBlacklistHandler(t *testing.T) {

	testUser := &data.User{
		ID:    1,
		Name:  "Dummy Username",
		Email: "test@example.com",
	}

	testSecret := "secretsecretsecret"

	type testCase struct {
		name               string
		users              []*data.User
		refreshToken       []byte
		expectedStatusCode int
	}

	testTokenID1 := uuid.NewString()

	testCases := []testCase{
		{
			name:               "happy path",
			users:              []*data.User{testUser},
			refreshToken:       GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "happy path2",
			users:              []*data.User{testUser},
			refreshToken:       GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "different user id",
			users:              []*data.User{testUser},
			refreshToken:       GenerateTestToken(testSecret, testTokenID1, 2, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "different user id",
			users:              []*data.User{testUser},
			refreshToken:       GenerateTestToken(testSecret, testTokenID1, 2, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "expired token",
			users:              []*data.User{testUser},
			refreshToken:       GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour), time.Now().Add(-48.*time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "wrong issuer",
			users:              []*data.User{testUser},
			refreshToken:       GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "xx", "refresh"),
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "happy path2",
			users:              []*data.User{testUser},
			refreshToken:       GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "wrongtokentype"),
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(tc.users)},
				auth:   auth.NewAuth(testSecret, mocks.NewBlacklistServiceMock()),
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /", app.blacklistHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/", ts.URL)

			type input struct {
				RefreshToken string `json:"refresh"`
			}

			reqBody, err := json.Marshal(input{RefreshToken: string(tc.refreshToken)})
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

		})
	}
}

// TestLogoutHandler unit tests logoutHandler.
func TestLogoutHandler(t *testing.T) {

	testUser := &data.User{
		ID:    1,
		Name:  "Dummy Username",
		Email: "test@example.com",
	}

	testSecret := "secretsecretsecret"

	testTokenID1 := uuid.NewString()
	testTokenID2 := uuid.NewString()
	// testTokenID3 := uuid.NewString()
	// testTokenID4 := uuid.NewString()

	type testCase struct {
		name               string
		users              []*data.User
		accessToken        []byte
		refreshToken       []byte
		blacklist          []string
		expectedStatusCode int
	}

	testCases := []testCase{
		{
			name:               "happy path",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "blacklisted access token",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          []string{testTokenID1},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "blacklisted refresh token",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          []string{testTokenID2},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "blacklisted refresh and access token",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          []string{testTokenID2, testTokenID1},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "expired access token",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-10*time.Hour), time.Now().Add(-10*time.Hour), time.Now().Add(-10*time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "expired refresh token",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "before notberore access token",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(time.Hour), time.Now().Add(time.Hour), time.Now().Add(time.Hour).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "before notberore refresh token",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(time.Hour), time.Now().Add(time.Hour), time.Now().Add(time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "not valid access token issuer",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "xx", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "not valid refresh token issuer",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "xx", "refresh"),
			blacklist:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "not valid access token type",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "xx"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "refresh"),
			blacklist:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "not valid refresh token type",
			users:              []*data.User{testUser},
			accessToken:        GenerateTestToken(testSecret, testTokenID1, testUser.ID, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute), time.Now().Add(-time.Minute).Add(5*time.Minute), "instrument-swap.example.example", "access"),
			refreshToken:       GenerateTestToken(testSecret, testTokenID2, testUser.ID, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour).Add(24*time.Hour), "instrument-swap.example.example", "xx"),
			blacklist:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Users: mocks.NewUserModelMock(tc.users)},
				auth:   auth.NewAuth(testSecret, mocks.NewBlacklistServiceMockWithData(tc.blacklist)),
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /", app.logoutHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/", ts.URL)

			type input struct {
				AccessToken  string `json:"access"`
				RefreshToken string `json:"refresh"`
			}

			reqBody, err := json.Marshal(input{AccessToken: string(tc.accessToken), RefreshToken: string(tc.refreshToken)})
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
		})
	}

}
