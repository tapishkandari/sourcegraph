package main

import (
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
)

func setupServer() []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{httpserver.New(port, noRoutes)}
}

func noRoutes(router *mux.Router) {}
