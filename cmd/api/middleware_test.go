package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
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
