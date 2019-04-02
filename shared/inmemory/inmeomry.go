package inmemory

import (
	"sync"
	"sync/atomic"
)

func NewIndexMap() *IndexMap {
	return &IndexMap{items: make(map[uint64]interface{})}
}

type IndexMap struct {
	lock    sync.RWMutex
	counter uint64
	items   map[uint64]interface{}
}

func (m *IndexMap) Index() uint64 {
	return atomic.AddUint64(&m.counter, 1)
}

func (m *IndexMap) Set(idx uint64, value interface{}) {
	m.lock.Lock()
	m.items[idx] = value
	m.lock.Unlock()
}

func (m *IndexMap) Get(idx uint64) (value interface{}, ok bool) {
	m.lock.RLock()
	v, ok := m.items[idx]
	m.lock.RUnlock()
	return v, ok
}

func (m *IndexMap) Range(f func(idx uint64, value interface{}) bool) {
	m.lock.Lock()
	for k, v := range m.items {
		if !f(k, v) {
			break
		}
	}
	m.lock.Unlock()
}

func (m *IndexMap) Remove(idx uint64) {
	m.lock.Lock()
	delete(m.items, idx)
	m.lock.Unlock()
}
