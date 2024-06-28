package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestLivelinessHandler test the livelinessHandler handler.
func TestLivelinessHandler(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	app := &application{}

	app.livelinessHandler(rr, r)

	rs := rr.Result()

	if rs.StatusCode != http.StatusOK {
		t.Errorf(`expected status code %d, got %d`, http.StatusOK, rs.StatusCode)
	}

	defer func() {
		err := rs.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	if !strings.Contains(string(body), `"status":"available"`) {
		t.Errorf(`Response should containe "status":"available"`)
	}

}
