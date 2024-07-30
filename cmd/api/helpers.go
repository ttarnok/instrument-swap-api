package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

// type envelope is used to envelope json responses.
type envelope map[string]any

// extractIDParam is used to extract id parameters from request paths.
func (app *application) extractIDParam(r *http.Request) (int64, error) {
	// Read and Validate params
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	if id < 1 {
		return 0, errors.New("invalid id value")
	}
	return id, nil
}

// writeJSON is used to send a JSON response to the clinet.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}
	return nil
}

// readJSON is used to extract JSON data from a client request.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	// 1Mb
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	// Now an error will thrown if the client sends unexpected fields (unknown field) in the body.
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {

		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {

		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Make sure the body contains only one JSON, not a stream.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// readQParamString is used to extract string typed query string from requests.
func (app *application) readQParamString(qs url.Values, key string, defaultValue string) string {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

// readQParamCSV is used to extract csv typed query string from requests.
func (app *application) readQParamCSV(qs url.Values, key string, defaultValue []string) []string {

	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// readQParamInt is used to extract integer typed query string from requests.
func (app *application) readQParamInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {

	sv := qs.Get(key)
	if sv == "" {
		return defaultValue
	}

	iv, err := strconv.Atoi(sv)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return iv

}
