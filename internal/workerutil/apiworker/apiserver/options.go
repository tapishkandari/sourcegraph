package apiserver

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiclient"
)

type Options struct {
	// TODO - document
	Port int

	// MaximumTransactions is the maximum number of active records that can be given out to indexers. The
	// manager dequeue method will stop returning records while the number of outstanding transactions is
	// at this threshold.
	MaximumTransactions int

	// RequeueDelay controls how far into the future to make an indexer's records visible to another
	// agent once it becomes unresponsive.
	RequeueDelay time.Duration

	// UnreportedMaxAge is the maximum time between an index record being dequeued and it appearing in
	// the indexer's heartbeat requests before it being considered lost.
	UnreportedIndexMaxAge time.Duration

	// DeathThreshold is the minimum time since the last indexerheartbeat before the indexer can be
	// considered as unresponsive. This should be configured to be longer than the indexer's heartbeat
	// interval.
	DeathThreshold time.Duration

	// TODO - document
	CleanupInterval time.Duration

	// TODO - document
	ToIndex func(record workerutil.Record) (apiclient.Index, error)
}
