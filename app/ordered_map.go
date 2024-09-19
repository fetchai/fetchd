package app

import (
	"fmt"
	"log"
	"sort"
)

type OrderedMap[K comparable, V any] struct {
	keys   *[]K
	values map[K]V
	sorted *bool
}

// NewOrderedMap creates a new OrderedMap instance
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	keys := []K{}
	sorted := false
	return &OrderedMap[K, V]{
		keys:   &keys,
		values: make(map[K]V),
		sorted: &sorted,
	}
}

// Set adds a key-value pair to the map
func (om *OrderedMap[K, V]) Set(key K, value V) {
	if _, exists := om.values[key]; !exists {
		*om.keys = append(*om.keys, key)
		*om.sorted = false
	}
	om.values[key] = value
}

// Set adds a key-value pair to the map - it must not exist before
func (om *OrderedMap[K, V]) SetNew(key K, value V) {
	if om.Has(key) {
		log.Panicf("key %v already exist", key)
	}
	om.Set(key, value)
}

// Get retrieves the value associated with the key
func (om *OrderedMap[K, V]) Get(key K) (*V, bool) {
	value, exists := om.values[key]
	return &value, exists
}

// Get retrieves the value associated with the key, panics otherwise
func (om *OrderedMap[K, V]) MustGet(key K) *V {
	value, exists := om.Get(key)
	if !exists {
		log.Panicf("key %v does not exist", key)
	}
	return value
}

func (om *OrderedMap[K, V]) Has(key K) bool {
	_, exists := om.Get(key)
	return exists
}

// Delete removes a key-value pair from the map
func (om *OrderedMap[K, V]) Delete(key K) {
	if _, exists := om.values[key]; exists {
		delete(om.values, key)
		// Remove key from slice
		for i, k := range *om.keys {
			if k == key {
				*om.keys = append((*om.keys)[:i], (*om.keys)[i+1:]...)
				break
			}
		}
	}
}

// Keys returns the keys in insertion order
func (om *OrderedMap[K, V]) Keys() *[]K {
	return om.keys
}

// PrintOrdered prints the map in current order
func (om *OrderedMap[K, V]) PrintOrdered() {
	for _, key := range *om.keys {
		fmt.Printf("%v: %v\n", key, om.values[key])
	}
}

// SortKeys sorts the keys in ascending order
func (om *OrderedMap[K, V]) SortKeys(lessFunc func(i, j K) bool) {
	if *om.sorted {
		return
	}
	sort.Slice(*om.keys, func(i, j int) bool {
		return lessFunc((*om.keys)[i], (*om.keys)[j])
	})
	*om.sorted = true
}

func sortUint64Keys[V any](orderedMap *OrderedMap[uint64, V]) {
	orderedMap.SortKeys(func(i, j uint64) bool {
		return i < j
	})
}
