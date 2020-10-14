package apiserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
)

var shutdownErr = errors.New("service shutting down")

// Handle requeues every locked index record assigned to indexers which have not been updated
// for longer than the death threshold.
func (m *handler) Handle(ctx context.Context) error {
	return m.requeueIndexes(ctx, m.pruneIndexers())
}

func (m *handler) HandleError(err error) {
	log15.Error("Failed to requeue indexes", "err", err)
}

func (m *handler) OnShutdown() {
	m.m.Lock()
	defer m.m.Unlock()

	for _, indexer := range m.indexers {
		for _, meta := range indexer.metas {
			if err := meta.tx.Done(shutdownErr); err != nil && err != shutdownErr {
				log15.Error(fmt.Sprintf("Failed to close transaction holding index %d", meta.index.RecordID()), "err", err)
			}
		}
	}
}

// Heartbeat bumps the last updated time of the indexer and closes any transactions locking
// records whose identifiers were not supplied.
func (m *handler) Heartbeat(ctx context.Context, indexerName string, indexIDs []int) error {
	return m.requeueIndexes(ctx, m.pruneIndexes(indexerName, indexIDs))
}

// pruneIndexes removes the indexes whose identifier is not in the given list from the given indexer.
// This method returns the index meta values which were removed. Index meta values which were created
// very recently will be counted as live to account for the time between when the record is dequeued
// in this service and when it is added to the heartbeat requests from the indexer. This method also
// updates the last updated time of the indexer.
func (m *handler) pruneIndexes(indexerName string, ids []int) (dead []indexMeta) {
	now := m.clock.Now()

	idMap := map[int]struct{}{}
	for _, id := range ids {
		idMap[id] = struct{}{}
	}

	m.m.Lock()
	defer m.m.Unlock()

	indexer, ok := m.indexers[indexerName]
	if !ok {
		indexer = &indexerMeta{}
		m.indexers[indexerName] = indexer
	}

	var live []indexMeta
	for _, meta := range indexer.metas {
		if _, ok := idMap[meta.index.RecordID()]; ok || now.Sub(meta.started) < m.options.UnreportedIndexMaxAge {
			live = append(live, meta)
		} else {
			dead = append(dead, meta)
		}
	}

	indexer.metas = live
	indexer.lastUpdate = now
	return dead
}

// requeueIndexes requeues the given index records.
func (m *handler) requeueIndexes(ctx context.Context, metas []indexMeta) (errs error) {
	for _, meta := range metas {
		if err := m.requeueIndex(ctx, meta); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

// requeueIndex requeues the given index record , then finalizes the transaction that locks that record.
func (m *handler) requeueIndex(ctx context.Context, meta indexMeta) error {
	defer func() { m.dequeueSemaphore <- struct{}{} }()

	err := meta.tx.Requeue(ctx, meta.index.RecordID(), m.clock.Now().Add(m.options.RequeueDelay))
	return meta.tx.Done(err)
}

// pruneIndexers removes the data associated with indexers which have not been updated for longer
// than the death threshold and returns all index meta values assigned to removed indexers.
func (m *handler) pruneIndexers() (metas []indexMeta) {
	m.m.Lock()
	defer m.m.Unlock()

	for name, indexer := range m.indexers {
		if m.clock.Now().Sub(indexer.lastUpdate) <= m.options.DeathThreshold {
			continue
		}

		metas = append(metas, indexer.metas...)
		delete(m.indexers, name)
	}

	return metas
}
