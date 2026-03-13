package typedmap

import (
	"maps"
	"slices"
	"sync"
)

// TypedMap is a thread-safe map of strings to T.
type TypedMap[T any] struct {
	m  map[string]T
	mu sync.Mutex
}

// New creates a new TypedMap instance.
func New[T any]() *TypedMap[T] {
	return &TypedMap[T]{
		m: make(map[string]T),
	}
}

// Load returns the value for key.
func (m *TypedMap[T]) Load(key string) (T, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.m[key]
	return val, ok
}

// Store sets key with val.
func (m *TypedMap[T]) Store(key string, val T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = val
}

// SwapFunc allows f to return a new value for key and change it,
// without releasing the lock on the TypedMap.
// Alternatively, f can error out so that nothing gets changed.
//
// If key is present in the TypedMap,
// f is called with the key, the previous value at key,
// and true for the ok value.
// If key is not in the TypedMap, f is passed the key,
// the zero value of T, and false for the ok value.
//
// If f returns with a nil error,
// the T value returned gets written into the TypedMap at key.
// If f returns an error, SwapFunc returns that error
// without changing the TypedMap.
func (m *TypedMap[T]) SwapFunc(
	key string,
	f func(key string, prev T, ok bool) (T, error),
) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	prev, ok := m.m[key]
	next, err := f(key, prev, ok)
	if err != nil {
		return err
	}
	m.m[key] = next
	return nil
}

func (m *TypedMap[T]) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Collect(maps.Keys(m.m))
}
