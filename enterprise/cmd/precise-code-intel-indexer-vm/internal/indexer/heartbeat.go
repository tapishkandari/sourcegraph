package indexer

import (
	"context"
	"time"

	queue "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewHeartbeat(ctx context.Context, queueClient queue.Client, idSet *IDSet, interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.NewHandlerWithErrorMessage("heartbeat", func(ctx context.Context) error {
		return queueClient.Heartbeat(ctx, idSet.Slice())
	}))
}
