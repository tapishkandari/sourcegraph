package indexer

import (
	"context"
	"time"

	"github.com/google/uuid"
	queue "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type IndexerOptions struct {
	NumIndexers    int
	Interval       time.Duration
	Metrics        IndexerMetrics
	HandlerOptions HandlerOptions
}

func NewIndexer(ctx context.Context, queueClient queue.Client, idSet *IDSet, options IndexerOptions) *workerutil.Worker {
	handler := &Handler{
		queueClient:   queueClient,
		idSet:         idSet,
		newCommander:  NewDefaultCommander,
		options:       options.HandlerOptions,
		uuidGenerator: uuid.NewRandom,
	}

	workerMetrics := workerutil.WorkerMetrics{
		HandleOperation: options.Metrics.ProcessOperation,
	}

	return workerutil.NewWorker(ctx, &storeShim{queueClient}, workerutil.WorkerOptions{
		Handler:     handler,
		NumHandlers: options.NumIndexers,
		Interval:    options.Interval,
		Metrics:     workerMetrics,
	})
}
