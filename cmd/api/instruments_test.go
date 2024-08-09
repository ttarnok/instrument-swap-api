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
	"testing"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/data/mocks"
)

// testData contains some dummy initial data for the mocked database.
var testData = []*data.Instrument{
	{
		ID:              1,
		CreatedAt:       time.Now().UTC(),
		Name:            "M1",
		Manufacturer:    "Korg",
		ManufactureYear: 1990,
		Type:            "synthesizer",
		EstimatedValue:  100000,
		Condition:       "used",
		Description:     "A music workstation manufactured by Korg.",
		FamousOwners:    []string{"The Orb"},
		OwnerUserID:     1,
		Version:         1,
	},
	{
		ID:              2,
		CreatedAt:       time.Now().UTC(),
		Name:            "Wavestation",
		Manufacturer:    "Korg",
		ManufactureYear: 1991,
		Type:            "synthesizer",
		EstimatedValue:  999,
		Condition:       "used",
		Description:     "A vector synthesis synthesizer first produced by Korg.",
		FamousOwners:    []string{"Depeche Mode", "Genesis"},
		OwnerUserID:     1,
		Version:         1,
	},
	{
		ID:              3,
		CreatedAt:       time.Now().UTC(),
		Name:            "DX7",
		Manufacturer:    "Yamaha",
		ManufactureYear: 1985,
		Type:            "synthesizer",
		EstimatedValue:  100000,
		Condition:       "used",
		Description:     "A frequency modulation based synthesizer manufactured by Yamaha.",
		FamousOwners:    []string{"Brian Eno"},
		OwnerUserID:     2,
		Version:         1,
	},
	{
		ID:              4,
		CreatedAt:       time.Now().UTC(),
		Name:            "ESQ-1",
		Manufacturer:    "Ensoniq",
		ManufactureYear: 1987,
		Type:            "synthesizer",
		EstimatedValue:  8000,
		Condition:       "used",
		Description:     "A digital wave morphing synthesizer manufactured by Ensoniq",
		FamousOwners:    []string{"Steve Roach"},
		OwnerUserID:     3,
		Version:         1,
	},
}

var testDataEmpty = []*data.Instrument{}

// TestListInstrumentsHandler unit tests the functionality of listInstrumentsHandler.
func TestListInstrumentsHandler(t *testing.T) {

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: data.Models{Instruments: mocks.NewNonEmptyInstrumentModelMock(testData)},
	}

	app.listInstrumentsHandler(rr, req)

	resp := rr.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(`expected status code %d, got %d`, http.StatusOK, resp.StatusCode)
	}

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

	respAnySlice, ok := respEnvelope["instruments"]
	if !ok {
		t.Fatal(`the response does not contain enveloped instruments`)
	}
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(respAnySlice)
	if err != nil {
		t.Fatal(err)
	}

	var inst []*data.Instrument

	err = json.NewDecoder(buf).Decode(&inst)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(testData, inst) {
		t.Error(`input and output for the handler is not equal`)
	}

}

// TestShowInstrumentHandler unit tests the functionality of showInstrumentHandler.
func TestShowInstrumentHandler(t *testing.T) {

	tests := []struct {
		name               string
		pathParam          string
		expectedStatusCode int
		checkResult        bool
		expectedIndex      int
	}{
		{
			name:               "happy path",
			pathParam:          "1",
			expectedStatusCode: http.StatusOK,
			checkResult:        true,
			expectedIndex:      0,
		},
		{
			name:               "non number path",
			pathParam:          "nonnumber",
			expectedStatusCode: http.StatusNotFound,
			checkResult:        false,
		},
		{
			name:               "non exsistent id",
			pathParam:          "100",
			expectedStatusCode: http.StatusNotFound,
			checkResult:        false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Instruments: mocks.NewNonEmptyInstrumentModelMock(testData)},
			}

			mux := http.NewServeMux()
			mux.HandleFunc("GET /{id}", app.showInstrumentHandler)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tt.pathParam)

			request, err := http.NewRequest("GET", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			res, err := client.Do(request)
			if err != nil {
				log.Fatal(err)
			}

			if res.StatusCode != tt.expectedStatusCode {
				t.Errorf(`expected status code %d, got %d`, tt.expectedStatusCode, res.StatusCode)
			}

			if tt.checkResult {

				bs, err := io.ReadAll(res.Body)
				if err != nil {
					log.Fatal(err)
				}
				defer func() {
					err := res.Body.Close()
					if err != nil {
						t.Fatal(err)
					}
				}()

				var env map[string]*data.Instrument

				err = json.Unmarshal(bs, &env)
				if err != nil {
					t.Fatal(err)
				}
				instrument := env["instrument"]

				if !reflect.DeepEqual(instrument, testData[tt.expectedIndex]) {
					t.Errorf(`expected message body "%#v", got "%#v"`, testData[tt.expectedIndex], instrument)
				}
			}
		})
	}

}

// TestCreateInstrumentHandler implememts unit tests for createInstrumentHandler.
func TestCreateInstrumentHandler(t *testing.T) {

	type inputInstrument struct {
		Name            string   `json:"name"`
		Manufacturer    string   `json:"manufacturer"`
		ManufactureYear int32    `json:"manufacture_year"`
		Type            string   `json:"type"`
		EstimatedValue  int64    `json:"estimated_value"`
		Condition       string   `json:"condition"`
		Description     string   `json:"description"`
		FamousOwners    []string `json:"famous_owners"`
	}

	type testCase struct {
		name                   string
		input                  inputInstrument
		ownerUser              data.User
		expectedStatusCode     int
		expectedLocationHeader string
		checkResult            bool
		expectedResult         *data.Instrument
		expectedErrResult      map[string]string
	}

	testCases := []testCase{
		{
			name: "happy path",
			input: inputInstrument{
				Name:            "M1",
				Manufacturer:    "Korg",
				ManufactureYear: 1990,
				Type:            "synthesizer",
				EstimatedValue:  100000,
				Condition:       "used",
				Description:     "A music workstation manufactured by Korg.",
				FamousOwners:    []string{"The Orb"},
				// OwnerUserID:     1,
			},
			ownerUser:              data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			expectedStatusCode:     http.StatusCreated,
			expectedLocationHeader: "/v1/instrumets/0",
			checkResult:            true,
			expectedResult: &data.Instrument{
				ID:              0,
				CreatedAt:       time.Time{},
				Name:            "M1",
				Manufacturer:    "Korg",
				ManufactureYear: 1990,
				Type:            "synthesizer",
				EstimatedValue:  100000,
				Condition:       "used",
				Description:     "A music workstation manufactured by Korg.",
				FamousOwners:    []string{"The Orb"},
				OwnerUserID:     1,
				Version:         0,
			},
			expectedErrResult: nil,
		},
		{
			name: "empty instrument name",
			input: inputInstrument{
				Name:            "",
				Manufacturer:    "Korg",
				ManufactureYear: 1990,
				Type:            "synthesizer",
				EstimatedValue:  100000,
				Condition:       "used",
				Description:     "A music workstation manufactured by Korg.",
				FamousOwners:    []string{"The Orb"},
				// OwnerUserID:     1,
			},
			ownerUser:              data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			expectedStatusCode:     http.StatusUnprocessableEntity,
			expectedLocationHeader: "",
			checkResult:            true,
			expectedResult:         nil,
			expectedErrResult:      map[string]string{"name": "must be provided"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &application{
				models: data.Models{
					Instruments: mocks.NewNonEmptyInstrumentModelMock(testDataEmpty),
				},
			}

			setUser := func(next http.HandlerFunc) http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					r = app.contextSetUser(r, &tc.ownerUser)

					next.ServeHTTP(w, r)
				})
			}

			ts := httptest.NewServer(setUser(http.HandlerFunc(app.createInstrumentHandler)))
			defer ts.Close()

			bsInput, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			request, err := http.NewRequest("POST", ts.URL, bytes.NewBuffer(bsInput))
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			res, err := client.Do(request)
			if err != nil {
				log.Fatal(err)
			}

			if res.StatusCode != tc.expectedStatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, res.StatusCode)
			}

			if tc.expectedLocationHeader != res.Header.Get("Location") {
				t.Errorf(`Expected Location header value "%s", got "%s"`, tc.expectedLocationHeader, res.Header.Get("Location"))
			}

			if tc.checkResult {
				bs, err := io.ReadAll(res.Body)
				if err != nil {
					log.Fatal(err)
				}
				defer func() {
					err := res.Body.Close()
					if err != nil {
						t.Fatal(err)
					}
				}()

				if tc.expectedResult != nil {

					var env map[string]*data.Instrument

					err = json.Unmarshal(bs, &env)
					if err != nil {
						t.Fatal(err)
					}
					instrument := env["instrument"]

					if !reflect.DeepEqual(tc.expectedResult, instrument) {
						t.Errorf("expected result:\n%#v, got:\n%#v", tc.expectedResult, instrument)
					}

				}

				if tc.expectedErrResult != nil {

					var env map[string]map[string]string

					err = json.Unmarshal(bs, &env)
					if err != nil {
						t.Fatal(err)
					}
					errResult := env["error"]

					if !reflect.DeepEqual(tc.expectedErrResult, errResult) {
						t.Errorf("expected result:\n%#v, got:\n%#v", tc.expectedErrResult, errResult)
					}
				}

			}

		})
	}

}

// TestUpdateInstrumentHandler implements unit tests for updateInstrumentHandler.
func TestUpdateInstrumentHandler(t *testing.T) {

	type inputInstrument struct {
		Name            string   `json:"name"`
		Manufacturer    string   `json:"manufacturer"`
		ManufactureYear int32    `json:"manufacture_year"`
		Type            string   `json:"type"`
		EstimatedValue  int64    `json:"estimated_value"`
		Condition       string   `json:"condition"`
		Description     string   `json:"description"`
		FamousOwners    []string `json:"famous_owners"`
	}

	type testCase struct {
		name               string
		input              inputInstrument
		ownerUser          data.User
		pathParam          string
		expectedIndex      int
		expectedStatusCode int
		expectedResult     *data.Instrument
		expectedErrResult  map[string]string
	}

	testCases := []testCase{
		{
			name: "happy path - partial update",
			input: inputInstrument{
				Name:         "TB303",
				Manufacturer: "Roland",
			},
			ownerUser:          data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "1",
			expectedIndex:      0,
			expectedStatusCode: http.StatusOK,
			expectedResult: &data.Instrument{
				ID:              1,
				CreatedAt:       testData[0].CreatedAt,
				Name:            "TB303",
				Manufacturer:    "Roland",
				ManufactureYear: 1990,
				Type:            "synthesizer",
				EstimatedValue:  100000,
				Condition:       "used",
				Description:     "A music workstation manufactured by Korg.",
				FamousOwners:    []string{"The Orb"},
				OwnerUserID:     1,
				Version:         1,
			},
			expectedErrResult: nil,
		},
		{
			name: "happy path - full update",
			input: inputInstrument{
				Name:            "E1",
				Manufacturer:    "Moog",
				ManufactureYear: 2000,
				Type:            "guitar",
				EstimatedValue:  2,
				Condition:       "poor",
				Description:     "This is a dummy test description",
				FamousOwners:    []string{"Apple", "Banana", "Cherry"},
			},
			ownerUser:          data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "1",
			expectedIndex:      0,
			expectedStatusCode: http.StatusOK,
			expectedResult: &data.Instrument{
				ID:              1,
				CreatedAt:       testData[0].CreatedAt,
				Name:            "E1",
				Manufacturer:    "Moog",
				ManufactureYear: 2000,
				Type:            "guitar",
				EstimatedValue:  2,
				Condition:       "poor",
				Description:     "This is a dummy test description",
				FamousOwners:    []string{"Apple", "Banana", "Cherry"},
				OwnerUserID:     1,
				Version:         1,
			},
			expectedErrResult: nil,
		},
		{
			name:               "not valid path param",
			input:              inputInstrument{},
			ownerUser:          data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "no",
			expectedIndex:      -1,
			expectedStatusCode: http.StatusNotFound,
			expectedResult:     nil,
			expectedErrResult:  map[string]string{"error": "the requested resource could not be found"},
		},
		{
			name:               "not existing path param",
			input:              inputInstrument{},
			ownerUser:          data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "10000",
			expectedIndex:      -1,
			expectedStatusCode: http.StatusNotFound,
			expectedResult:     nil,
			expectedErrResult:  map[string]string{"error": "the requested resource could not be found"},
		},
		{
			name: "attempt to update instrument for a different user",
			input: inputInstrument{
				Name:         "TB303",
				Manufacturer: "Roland",
			},
			ownerUser:          data.User{ID: 2, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "1",
			expectedIndex:      0,
			expectedStatusCode: http.StatusForbidden,
			expectedResult:     nil,
			expectedErrResult:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Instruments: mocks.NewNonEmptyInstrumentModelMock(testData)},
			}

			setUser := func(next http.HandlerFunc) http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					r = app.contextSetUser(r, &tc.ownerUser)

					next.ServeHTTP(w, r)
				})
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /{id}", setUser(app.updateInstrumentHandler))

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tc.pathParam)

			bs, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			request, err := http.NewRequest("POST", path, bytes.NewBuffer(bs))
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			res, err := client.Do(request)
			if err != nil {
				log.Fatal(err)
			}

			if tc.expectedStatusCode != res.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, res.StatusCode)
			}

			if tc.expectedResult != nil {

				bs, err := io.ReadAll(res.Body)
				if err != nil {
					log.Fatal(err)
				}
				defer func() {
					err := res.Body.Close()
					if err != nil {
						t.Fatal(err)
					}
				}()

				var env map[string]*data.Instrument

				err = json.Unmarshal(bs, &env)
				if err != nil {
					t.Fatal(err)
				}
				instrument := env["instrument"]

				if !reflect.DeepEqual(instrument, tc.expectedResult) {
					t.Errorf("expected message body\n%#v, got\n%#v", tc.expectedResult, instrument)
				}
			}

			if tc.expectedErrResult != nil {

				bs, err := io.ReadAll(res.Body)
				if err != nil {
					log.Fatal(err)
				}
				defer func() {
					err := res.Body.Close()
					if err != nil {
						t.Fatal(err)
					}
				}()

				var errEnv map[string]string

				err = json.Unmarshal(bs, &errEnv)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tc.expectedErrResult, errEnv) {
					t.Errorf("expected message body\n%#v, got\n%#v", tc.expectedErrResult, errEnv)
				}
			}
		})
	}
}

func TestDeleteInstrumentHandler(t *testing.T) {

	type testCase struct {
		name               string
		dabatase           []*data.Instrument
		ownerUser          data.User
		pathParam          string
		expectedStatusCode int
		expectedResult     map[string]string
	}

	testCases := []testCase{
		{
			name: "happy path",
			dabatase: []*data.Instrument{{
				ID:              1,
				CreatedAt:       time.Now().UTC(),
				Name:            "M1",
				Manufacturer:    "Korg",
				ManufactureYear: 1990,
				Type:            "synthesizer",
				EstimatedValue:  100000,
				Condition:       "used",
				Description:     "A music workstation manufactured by Korg.",
				FamousOwners:    []string{"The Orb"},
				OwnerUserID:     1,
				Version:         1,
			}},
			ownerUser:          data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "1",
			expectedStatusCode: http.StatusOK,
			expectedResult:     map[string]string{"message": "instrument successfully deleted"},
		},
		{
			name: "non existent id",
			dabatase: []*data.Instrument{{
				ID:              1,
				CreatedAt:       time.Now().UTC(),
				Name:            "M1",
				Manufacturer:    "Korg",
				ManufactureYear: 1990,
				Type:            "synthesizer",
				EstimatedValue:  100000,
				Condition:       "used",
				Description:     "A music workstation manufactured by Korg.",
				FamousOwners:    []string{"The Orb"},
				OwnerUserID:     1,
				Version:         1,
			}},
			ownerUser:          data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "1000",
			expectedStatusCode: http.StatusNotFound,
			expectedResult:     map[string]string{"error": "the requested resource could not be found"},
		},
		{
			name: "delete swapped instrument",
			dabatase: []*data.Instrument{{
				ID:              999,
				CreatedAt:       time.Now().UTC(),
				Name:            "M1",
				Manufacturer:    "Korg",
				ManufactureYear: 1990,
				Type:            "synthesizer",
				EstimatedValue:  100000,
				Condition:       "used",
				Description:     "A music workstation manufactured by Korg.",
				FamousOwners:    []string{"The Orb"},
				OwnerUserID:     1,
				Version:         1,
			}},
			ownerUser:          data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "999",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult:     map[string]string{"error": "can not perform the operation on a swapped instrument"},
		},
		{
			name: "delete an instrument owned by an other user",
			dabatase: []*data.Instrument{{
				ID:              1,
				CreatedAt:       time.Now().UTC(),
				Name:            "M1",
				Manufacturer:    "Korg",
				ManufactureYear: 1990,
				Type:            "synthesizer",
				EstimatedValue:  100000,
				Condition:       "used",
				Description:     "A music workstation manufactured by Korg.",
				FamousOwners:    []string{"The Orb"},
				OwnerUserID:     2,
				Version:         1,
			}},
			ownerUser:          data.User{ID: 1, Name: "Test User", Email: "testuser@example.com"},
			pathParam:          "1",
			expectedStatusCode: http.StatusForbidden,
			expectedResult:     map[string]string{"error": "Forbidden"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				models: data.Models{Instruments: mocks.NewNonEmptyInstrumentModelMock(tc.dabatase)},
			}

			setUser := func(next http.HandlerFunc) http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					r = app.contextSetUser(r, &tc.ownerUser)

					next.ServeHTTP(w, r)
				})
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /{id}", setUser(app.deleteInstrumentHandler))

			ts := httptest.NewServer(mux)
			defer ts.Close()

			path := fmt.Sprintf("%s/%s", ts.URL, tc.pathParam)

			request, err := http.NewRequest("POST", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			client := &http.Client{}
			res, err := client.Do(request)
			if err != nil {
				log.Fatal(err)
			}

			if tc.expectedStatusCode != res.StatusCode {
				t.Errorf(`expected status code %d, got %d`, tc.expectedStatusCode, res.StatusCode)
			}

			bs, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
			}()

			var env map[string]string

			err = json.Unmarshal(bs, &env)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expectedResult, env) {
				t.Errorf("expected message body\n%#v, got\n%#v", tc.expectedResult, env)
			}
		})
	}

}
