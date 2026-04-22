package store

import "sync"

type Store struct {
	// Add mutex to make it concurrency safe and it's fast
	// as both sits right next to each other in memory layout
	cred map[string]map[string]any
	mu   sync.RWMutex
}

// constructor
func NewStore() *Store {
	return &Store{
		cred: make(map[string]map[string]any),
	}
}

// Add or Update
func (s *Store) Set(id string, values map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cred[id] = values
}

// Read
func (s *Store) Get(id string) (map[string]any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.cred[id]

	// caller can mutate the values, returning internal map
	// make shallow copy and return
	copy := make(map[string]any)
	for k, v := range val {
		copy[k] = v
	}
	return copy, ok
}

// delete
func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cred, id)
}
