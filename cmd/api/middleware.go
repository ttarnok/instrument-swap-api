package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorLogResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		handler.ServeHTTP(w, r)
	})
}
