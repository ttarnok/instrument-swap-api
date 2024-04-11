package main

import "net/http"

func (a application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Im alive"))
}
