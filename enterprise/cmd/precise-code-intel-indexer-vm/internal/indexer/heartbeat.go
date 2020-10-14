package indexer

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewHeartbeat(ctx context.Context, queueClient queueClient, idSet *IDSet, interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.NewHandlerWithErrorMessage("heartbeat", func(ctx context.Context) error {
		return queueClient.Heartbeat(ctx, idSet.Slice())
	}))
}
