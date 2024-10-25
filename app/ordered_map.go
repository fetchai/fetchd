package app

import (
	"fmt"
	"log"
	"sort"
)

type OrderedMap[K comparable, V any] struct {
	keys     []K
	values   map[K]V
	isSorted bool
}

type Pair[K comparable, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}

// NewOrderedSet creates an OrderedMap with the keys set to true
func NewOrderedSet[K comparable](keys []K) *OrderedMap[K, bool] {
	newMap := NewOrderedMap[K, bool]()

	for _, key := range keys {
		newMap.Set(key, true)
	}

	return newMap
}

// NewOrderedMapFromPairs creates an OrderedMap from the keys-value pairs
func NewOrderedMapFromPairs[K comparable, V any](pairs []Pair[K, V]) *OrderedMap[K, V] {
	newMap := NewOrderedMap[K, V]()

	for _, pair := range pairs {
		newMap.SetNew(pair.Key, pair.Value)
	}

	return newMap
}

// NewOrderedMap creates a new OrderedMap empty instance
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	keys := []K{}
	sorted := false
	return &OrderedMap[K, V]{
		keys:     keys,
		values:   make(map[K]V),
		isSorted: sorted,
	}
}

// Set adds a key-value pair to the map
func (om *OrderedMap[K, V]) Set(key K, value V) {
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key)
		om.isSorted = false
	}
	om.values[key] = value
}

// SetNew adds a key-value pair to the map - it must not exist before
func (om *OrderedMap[K, V]) SetNew(key K, value V) {
	if om.Has(key) {
		log.Panicf("key %v already exist", key)
	}
	om.Set(key, value)
}

// Get retrieves the value associated with the key
func (om *OrderedMap[K, V]) Get(key K) (V, bool) {
	value, exists := om.values[key]
	return value, exists
}

// MustGet retrieves the value associated with the key, panics otherwise
func (om *OrderedMap[K, V]) MustGet(key K) V {
	value, exists := om.Get(key)
	if !exists {
		log.Panicf("key %v does not exist", key)
	}
	return value
}

func (om *OrderedMap[K, V]) GetOrSetDefault(key K, defaultValue V) (V, bool) {
	if value, exists := om.Get(key); exists {
		return value, false
	} else {
		om.Set(key, defaultValue)
		return om.MustGet(key), true
	}
}

// Iterate returns a channel that yields key-value pairs in insertion order
func (om *OrderedMap[K, V]) Iterate() <-chan Pair[K, V] {
	ch := make(chan Pair[K, V])
	go func() {
		for _, key := range om.keys {
			ch <- Pair[K, V]{Key: key, Value: om.MustGet(key)}
		}
		close(ch)
	}()
	return ch
}

func (om *OrderedMap[K, V]) Has(key K) bool {
	_, exists := om.Get(key)
	return exists
}

// Delete removes a key-value pair from the map
func (om *OrderedMap[K, V]) Delete(key K) {
	if _, exists := om.values[key]; exists {
		delete(om.values, key)
		// Create a new slice to avoid modifying the original reference
		newKeys := make([]K, 0, len(om.keys)-1)

		// Remove the key from the keys slice
		for _, k := range om.keys {
			if k != key {
				newKeys = append(newKeys, k)
			}
		}

		om.keys = newKeys // Assign the newly created slice to om.keys
	}
}

// Keys returns the keys in insertion order
func (om *OrderedMap[K, V]) Keys() []K {
	return om.keys
}

// PrintOrdered prints the map in current order
func (om *OrderedMap[K, V]) PrintOrdered() {
	for _, key := range om.keys {
		fmt.Printf("%v: %v\n", key, om.values[key])
	}
}

// SortKeys sorts the keys in ascending order
func (om *OrderedMap[K, V]) SortKeys(lessFunc func(i, j K) bool) {
	if om.isSorted {
		return
	}
	sort.Slice(om.keys, func(i, j int) bool {
		return lessFunc(om.keys[i], om.keys[j])
	})
	om.isSorted = true
}

// IsSorted returns true *IF* order in the map is *still* sorted after calling `SortKeys(...)` method.
func (om *OrderedMap[K, V]) IsSorted() bool {
	return om.isSorted
}

func sortUint64Keys[V any](orderedMap *OrderedMap[uint64, V]) {
	orderedMap.SortKeys(func(i, j uint64) bool {
		return i < j
	})
}
