package testhelpers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// DoTestAPICall helper function calls the given test endpoint, and returns the response.
// DoTestAPICall handles potential errors via the given *testint.T value.
func DoTestAPICall(t *testing.T, method string, url string, body io.Reader) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return resp

}

// GetResponseBody genric helper function accepts a respinse object, and returns the response body.
// The type of the response type should be given as a type a paramether during the function call.
// GetResponseBody handles potential errors via the given *testint.T value.
func GetResponseBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
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

	var respEnvelope T

	err = json.Unmarshal(body, &respEnvelope)
	if err != nil {
		t.Fatal(err)
	}

	return respEnvelope
}
