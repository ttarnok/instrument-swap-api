package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

type application struct {
	config config
	logger *slog.Logger
}

func main() {

	// ----------------------------–----------------------------------------------
	// Init config

	var cfg config = config{4000, "dev"}

	// ----------------------------–----------------------------------------------
	// Init logger

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// ----------------------------–----------------------------------------------
	// Init and Stratup Server

	app := application{config: cfg, logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/liveliness", app.livelinesskHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env, "version", version)
	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

func (a application) livelinesskHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Im alive"))
}
