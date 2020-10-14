package apiserver

import (
	"context"

	"github.com/efritz/glock"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewServer(store MetadataStore, workerStore dbworkerstore.Store, options Options) goroutine.BackgroundRoutine {
	handler := newHandler(store, workerStore, options, glock.NewRealClock())

	return goroutine.Combine(
		httpserver.New(options.Port, handler.setupRoutes),
		goroutine.NewPeriodicGoroutine(context.Background(), options.CleanupInterval, handler),
	)
}
