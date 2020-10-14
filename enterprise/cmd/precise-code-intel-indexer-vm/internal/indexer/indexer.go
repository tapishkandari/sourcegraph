package indexer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/indexer/command"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type Options struct {
	NumHandlers           int
	PollInterval          time.Duration
	HeartbeatInterval     time.Duration
	WorkerMetrics         workerutil.WorkerMetrics
	FrontendURL           string
	FrontendURLFromDocker string
	AuthToken             string
	FirecrackerImage      string
	UseFirecracker        bool
	FirecrackerNumCPUs    int
	FirecrackerMemory     string
	FirecrackerDiskSpace  string
	ImageArchivePath      string
}

func NewIndexer(queueClient queueClient, options Options) goroutine.BackgroundRoutine {
	idSet := newIDSet()
	store := &storeShim{queueClient}
	handler := &Handler{
		queueClient:   queueClient,
		idSet:         idSet,
		commandRunner: command.DefaultRunner,
		options:       options,
		uuidGenerator: uuid.NewRandom,
	}

	indexer := workerutil.NewWorker(context.Background(), store, workerutil.WorkerOptions{
		Handler:     handler,
		NumHandlers: options.NumHandlers,
		Interval:    options.PollInterval,
		Metrics:     options.WorkerMetrics,
	})

	heartbeat := goroutine.NewPeriodicGoroutine(
		context.Background(),
		options.HeartbeatInterval,
		goroutine.NewHandlerWithErrorMessage("heartbeat", func(ctx context.Context) error {
			return queueClient.Heartbeat(ctx, idSet.Slice())
		}),
	)

	return goroutine.Combine(indexer, heartbeat)
}
