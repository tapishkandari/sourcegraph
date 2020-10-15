package apiserver

import (
	"context"
	"sync"
	"time"

	"github.com/efritz/glock"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiclient"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type handler struct {
	store            MetadataStore
	workerStore      dbworkerstore.Store
	options          Options
	clock            glock.Clock
	indexers         map[string]*indexerMeta
	dequeueSemaphore chan struct{} // tracks available dequeue slots
	m                sync.Mutex    // protects indexers
}

var _ goroutine.Handler = &handler{}

// indexerMeta tracks the last request time of an indexer along with the set of index records which it
// is currently processing.
type indexerMeta struct {
	lastUpdate time.Time
	metas      []indexMeta
}

// indexMeta wraps an index record and the tranaction that is currently locking it for processing.
type indexMeta struct {
	index   workerutil.Record
	tx      dbworkerstore.Store
	started time.Time
}

// TODO - document
type MetadataStore interface {
	// TODO - document
	SetIndexLogContents(ctx context.Context, store basestore.ShareableStore, indexID int, contents string) error
}

func newHandler(store MetadataStore, workerStore dbworkerstore.Store, options Options, clock glock.Clock) *handler {
	dequeueSemaphore := make(chan struct{}, options.MaximumTransactions)
	for i := 0; i < options.MaximumTransactions; i++ {
		dequeueSemaphore <- struct{}{}
	}

	return &handler{
		store:            store,
		workerStore:      workerStore,
		options:          options,
		clock:            clock,
		dequeueSemaphore: dequeueSemaphore,
		indexers:         map[string]*indexerMeta{},
	}
}

// Dequeue pulls an unprocessed index record from the database and assigns the transaction that
// locks that record to the given indexer.
func (m *handler) Dequeue(ctx context.Context, indexerName string) (_ apiclient.Index, dequeued bool, _ error) {
	select {
	case <-m.dequeueSemaphore:
	default:
		return apiclient.Index{}, false, nil
	}
	defer func() {
		if !dequeued {
			// Ensure that if we do not dequeue a record successfully we do not
			// leak from the semaphore. This will happen if the dequeue call fails
			// or if there are no records to process
			m.dequeueSemaphore <- struct{}{}
		}
	}()

	record, tx, dequeued, err := m.workerStore.DequeueWithIndependentTransactionContext(ctx, nil)
	if err != nil {
		return apiclient.Index{}, false, err
	}
	if !dequeued {
		return apiclient.Index{}, false, nil
	}

	index, err := m.options.ToIndex(record)
	if err != nil {
		return apiclient.Index{}, false, tx.Done(err)
	}

	now := m.clock.Now()
	m.addMeta(indexerName, indexMeta{index: record, tx: tx, started: now})
	return index, true, nil
}

// addMeta removes the given index to the given indexer. This method also updates the last
// updated time of the indexer.
func (m *handler) addMeta(indexerName string, meta indexMeta) {
	m.m.Lock()
	defer m.m.Unlock()

	indexer, ok := m.indexers[indexerName]
	if !ok {
		indexer = &indexerMeta{}
		m.indexers[indexerName] = indexer
	}

	now := m.clock.Now()
	indexer.metas = append(indexer.metas, meta)
	indexer.lastUpdate = now
}

// SetLogContents updates a currently processing index record with the given log contents.
func (m *handler) SetLogContents(ctx context.Context, indexerName string, indexID int, contents string) error {
	index, ok := m.findMeta(indexerName, indexID, false)
	if !ok {
		return nil
	}

	// We're holding the index in a transaction, so if we want to modify that record we
	// need to do it in the same transaction. Here, we call the SetIndexLogContents method
	// on the codeintel store using the transaction attached to the processing index record.
	if err := m.store.SetIndexLogContents(ctx, index.tx, indexID, contents); err != nil {
		return err
	}

	return nil
}

// Complete marks the target index record as complete or errored depending on the existence of
// an error message, then finalizes the transaction that locks that record.
func (m *handler) Complete(ctx context.Context, indexerName string, indexID int, errorMessage string) (bool, error) {
	index, ok := m.findMeta(indexerName, indexID, true)
	if !ok {
		return false, nil
	}

	if err := m.completeIndex(ctx, index, errorMessage); err != nil {
		return false, err
	}

	return true, nil
}

// findMeta finds and returns an index meta value matching the given index identifier. If remove is
// true and the meta value is found, it is removed from the manager.
func (m *handler) findMeta(indexerName string, indexID int, remove bool) (indexMeta, bool) {
	m.m.Lock()
	defer m.m.Unlock()

	indexer, ok := m.indexers[indexerName]
	if !ok {
		return indexMeta{}, false
	}

	for i, meta := range indexer.metas {
		if meta.index.RecordID() == indexID {
			if remove {
				l := len(indexer.metas) - 1
				indexer.metas[i] = indexer.metas[l]
				indexer.metas = indexer.metas[:l]
			}

			return meta, true
		}
	}

	return indexMeta{}, false
}

// completeIndex marks the target index record as complete or errored depending on the existence
// of an error message, then finalizes the transaction that locks that record.
func (m *handler) completeIndex(ctx context.Context, meta indexMeta, errorMessage string) (err error) {
	defer func() { m.dequeueSemaphore <- struct{}{} }()

	if errorMessage == "" {
		_, err = meta.tx.MarkComplete(ctx, meta.index.RecordID())
	} else {
		_, err = meta.tx.MarkErrored(ctx, meta.index.RecordID(), errorMessage)
	}

	return meta.tx.Done(err)
}
