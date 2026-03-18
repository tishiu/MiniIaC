package state

import (
	"context"
	"fmt"
)

type ResourceIndex interface {
	Get(logicalID string) (*ResourceState, bool)
	Put(logicalID string, rs *ResourceState)
	Delete(logicalID string)
	DeleteByProviderID(providerID string)
	ForEach(func(logicalID string, rs *ResourceState) bool)
}

type Txn struct {
	State *State
	Index ResourceIndex
}

type memoryIndex struct {
	state             *State
	providerToLogical map[string]string
}

func newMemoryIndex(s *State) *memoryIndex {
	idx := &memoryIndex{
		state:             s,
		providerToLogical: make(map[string]string, len(s.Resources)),
	}
	for logicalID, rs := range s.Resources {
		idx.providerToLogical[rs.ID] = logicalID
	}
	return idx
}

func (i *memoryIndex) Get(logicalID string) (*ResourceState, bool) {
	rs, ok := i.state.Resources[logicalID]
	return rs, ok
}

func (i *memoryIndex) Put(logicalID string, rs *ResourceState) {
	if existing, ok := i.state.Resources[logicalID]; ok {
		delete(i.providerToLogical, existing.ID)
	}
	i.state.Resources[logicalID] = rs
	i.providerToLogical[rs.ID] = logicalID
}

func (i *memoryIndex) Delete(logicalID string) {
	if existing, ok := i.state.Resources[logicalID]; ok {
		delete(i.providerToLogical, existing.ID)
	}
	delete(i.state.Resources, logicalID)
}

func (i *memoryIndex) DeleteByProviderID(providerID string) {
	logicalID, ok := i.providerToLogical[providerID]
	if !ok {
		return
	}
	i.Delete(logicalID)
}

func (i *memoryIndex) ForEach(fn func(logicalID string, rs *ResourceState) bool) {
	for logicalID, rs := range i.state.Resources {
		if !fn(logicalID, rs) {
			return
		}
	}
}

func (m *Manager) Transact(ctx context.Context, fn func(tx *Txn) error) error {
	if ctx != nil && ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	return m.WithLock(func() error {
		current, err := m.Load()
		if err != nil {
			return err
		}

		tx := &Txn{
			State: current,
			Index: newMemoryIndex(current),
		}

		if err := fn(tx); err != nil {
			// Preserve partial progress by design: this matches existing apply/destroy behavior.
			_ = m.Save(tx.State)
			return err
		}

		return m.Save(tx.State)
	})
}
