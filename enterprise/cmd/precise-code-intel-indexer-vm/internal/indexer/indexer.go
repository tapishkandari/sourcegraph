package indexer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type IndexerOptions struct {
	NumIndexers    int
	Interval       time.Duration
	Metrics        IndexerMetrics
	HandlerOptions HandlerOptions
}

func NewIndexer(ctx context.Context, queueClient queueClient, idSet *IDSet, options IndexerOptions) *workerutil.Worker {
	handler := &Handler{
		queueClient:   queueClient,
		idSet:         idSet,
		commander:     DefaultCommander,
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
