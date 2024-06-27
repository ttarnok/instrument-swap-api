package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestLogError(t *testing.T) {

	testErr := errors.New("test error")

	expectedLevel := "ERROR"
	expextedMsg := testErr.Error()
	exPectedMethod := "GET"
	expectedURI := "/path"

	url := fmt.Sprintf("https://www.example.com%s", expectedURI)
	fmt.Println(url)

	buf := &bytes.Buffer{}

	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))
	app := &application{logger: logger}

	r, err := http.NewRequest(exPectedMethod, url, nil)
	if err != nil {
		t.Fatal("can not set up request for testing")
	}

	app.logError(r, testErr)

	var jRes struct {
		Time   time.Time `json:"time"`
		Level  string    `json:"level"`
		Msg    string    `json:"msg"`
		Method string    `json:"method"`
		URI    string    `json:"uri"`
	}

	err = json.Unmarshal(buf.Bytes(), &jRes)
	fmt.Println(jRes)
	if err != nil {
		t.Fatal("can not marshall data for testing")
	}

	if jRes.Level != expectedLevel {
		t.Errorf(`expected "%s", got "%s"`, expectedLevel, jRes.Level)
	}

	if jRes.Msg != expextedMsg {
		t.Errorf(`expected "%s", got "%s"`, expextedMsg, jRes.Msg)
	}

	if jRes.Method != exPectedMethod {
		t.Errorf(`expected "%s", got "%s"`, exPectedMethod, jRes.Method)
	}

	if jRes.URI != expectedURI {
		t.Errorf(`expected "%s", got "%s"`, expectedURI, jRes.URI)
	}

}
