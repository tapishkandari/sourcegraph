package indexer

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// storeShim converts a queue client into a workerutil.Store.
type storeShim struct {
	queueClient queueClient
}

var _ workerutil.Store = &storeShim{}

// Dequeue calls into the inner client.
func (s *storeShim) Dequeue(ctx context.Context, extraArguments interface{}) (workerutil.Record, workerutil.Store, bool, error) {
	var index Index // TODO - make generic
	dequeued, err := s.queueClient.Dequeue(ctx, &index)
	return index, s, dequeued, err
}

// Dequeue MarkComplete into the inner client.
func (s *storeShim) MarkComplete(ctx context.Context, id int) (bool, error) {
	return true, s.queueClient.Complete(ctx, id, nil)
}

// MarkErrored calls into the inner client.
func (s *storeShim) MarkErrored(ctx context.Context, id int, failureMessage string) (bool, error) {
	return true, s.queueClient.Complete(ctx, id, errors.New(failureMessage))
}

// Done is a no-op.
func (s *storeShim) Done(err error) error {
	return err
}

// SetLogContents calls into the inner client.
func (s *storeShim) SetLogContents(ctx context.Context, id int, payload string) error {
	return s.queueClient.SetLogContents(ctx, id, payload)
}
