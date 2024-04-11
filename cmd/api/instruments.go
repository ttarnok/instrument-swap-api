package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) showInstrumentHandler(w http.ResponseWriter, r *http.Request) {
	// Read and Validate params
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "Param value: %d\n", id)

}

func (app *application) createInstrumentHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Creating a new instrument"))

}
