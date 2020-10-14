package apiworker

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/command"
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

type QueueStore interface {
	Dequeue(ctx context.Context, payload interface{}) (bool, error)
	SetLogContents(ctx context.Context, indexID int, contents string) error
	Complete(ctx context.Context, indexID int, indexErr error) error
	Heartbeat(ctx context.Context, indexIDs []int) error
}

func NewIndexer(queueStore QueueStore, options Options) goroutine.BackgroundRoutine {
	idSet := newIDSet()
	store := &storeShim{queueStore}
	handler := &Handler{
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
			return queueStore.Heartbeat(ctx, idSet.Slice())
		}),
	)

	return goroutine.Combine(indexer, heartbeat)
}

// storeShim converts a queue client into a workerutil.Store.
type storeShim struct {
	queueStore QueueStore
}

var _ workerutil.Store = &storeShim{}

// Dequeue calls into the inner client.
func (s *storeShim) Dequeue(ctx context.Context, extraArguments interface{}) (workerutil.Record, workerutil.Store, bool, error) {
	var index Index // TODO - make generic
	dequeued, err := s.queueStore.Dequeue(ctx, &index)
	return index, s, dequeued, err
}

// SetLogContents calls into the inner client.
func (s *storeShim) SetLogContents(ctx context.Context, id int, payload string) error {
	return s.queueStore.SetLogContents(ctx, id, payload)
}

// Dequeue MarkComplete into the inner client.
func (s *storeShim) MarkComplete(ctx context.Context, id int) (bool, error) {
	return true, s.queueStore.Complete(ctx, id, nil)
}

// MarkErrored calls into the inner client.
func (s *storeShim) MarkErrored(ctx context.Context, id int, failureMessage string) (bool, error) {
	return true, s.queueStore.Complete(ctx, id, errors.New(failureMessage))
}

// Done is a no-op.
func (s *storeShim) Done(err error) error {
	return err
}
