package indexer

import (
	"sort"
	"sync"
)

type IDSet struct {
	sync.RWMutex
	ids map[int]struct{}
}

func newIDSet() *IDSet {
	return &IDSet{ids: map[int]struct{}{}}
}

func (i *IDSet) Add(indexID int) {
	i.Lock()
	i.ids[indexID] = struct{}{}
	i.Unlock()
}

func (i *IDSet) Remove(indexID int) {
	i.Lock()
	delete(i.ids, indexID)
	i.Unlock()
}

func (i *IDSet) Slice() []int {
	i.RLock()
	defer i.RUnlock()

	ids := make([]int, 0, len(i.ids))
	for id := range i.ids {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	return ids
}
