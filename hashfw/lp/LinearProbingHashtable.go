package lp

import (
	"hashfw/hashfw/fw/alloc"
	"hashfw/hashfw/fw/grower"
	"hashfw/hashfw/fw/hash"
	"unsafe"
)

// Cell represents a single cell in the hashtable
type Cell[K comparable, V any] struct {
	Key       K
	Value     V
	Occupied  bool
	Tombstone bool
}

// LinearProbingHashtable implements Open Addressing with Linear Probing collision resolution
type LinearProbingHashtable[K comparable, V any] struct {
	cells     []Cell[K, V]
	grower    *grower.Grower
	allocator *alloc.SimpleAllocator[Cell[K, V]]
	count     int
	hashFunc  func(K) uint64
}

// New creates a new LinearProbingHashtable
func New[K comparable, V any](hashFunc func(K) uint64) *LinearProbingHashtable[K, V] {
	g := grower.NewGrower()
	alloc := alloc.NewSimpleAllocator[Cell[K, V]]()

	return &LinearProbingHashtable[K, V]{
		cells:     alloc.Alloc(int(g.BufSize())),
		grower:    g,
		allocator: alloc,
		count:     0,
		hashFunc:  hashFunc,
	}
}

// NewWithCapacity creates a hashtable with expected capacity
func NewWithCapacity[K comparable, V any](capacity uint64, hashFunc func(K) uint64) *LinearProbingHashtable[K, V] {
	g := grower.NewGrowerWithSize(capacity)
	alloc := alloc.NewSimpleAllocator[Cell[K, V]]()

	return &LinearProbingHashtable[K, V]{
		cells:     alloc.Alloc(int(g.BufSize())),
		grower:    g,
		allocator: alloc,
		count:     0,
		hashFunc:  hashFunc,
	}
}

// Put inserts or updates a key-value pair
func (ht *LinearProbingHashtable[K, V]) Put(key K, value V) {
	// Check if resize needed
	if ht.grower.Overflow(ht.count + 1) {
		ht.resize()
	}

	h := ht.hashFunc(key)
	pos := ht.grower.Place(h)
	bufSize := ht.grower.BufSize()

	for i := uint64(0); i < bufSize; i++ {
		idx := (pos + i) & ht.grower.Mask()

		cell := &ht.cells[idx]

		// Found empty or tombstone slot
		if !cell.Occupied || cell.Tombstone {
			cell.Key = key
			cell.Value = value
			cell.Occupied = true
			cell.Tombstone = false
			ht.count++
			return
		}

		// Key already exists, update value
		if cell.Key == key {
			cell.Value = value
			return
		}
	}
}

// Lookup retrieves a value by key, returns (value, found)
func (ht *LinearProbingHashtable[K, V]) Lookup(key K) (V, bool) {
	h := ht.hashFunc(key)
	pos := ht.grower.Place(h)
	bufSize := ht.grower.BufSize()

	for i := uint64(0); i < bufSize; i++ {
		idx := (pos + i) & ht.grower.Mask()
		cell := &ht.cells[idx]

		// Empty slot - key not found
		if !cell.Occupied {
			var zero V
			return zero, false
		}

		// Skip tombstones
		if cell.Tombstone {
			continue
		}

		// Found the key
		if cell.Key == key {
			return cell.Value, true
		}
	}

	var zero V
	return zero, false
}

// Remove deletes a key from the hashtable
func (ht *LinearProbingHashtable[K, V]) Remove(key K) bool {
	h := ht.hashFunc(key)
	pos := ht.grower.Place(h)
	bufSize := ht.grower.BufSize()

	for i := uint64(0); i < bufSize; i++ {
		idx := (pos + i) & ht.grower.Mask()
		cell := &ht.cells[idx]

		// Empty slot - key not found
		if !cell.Occupied {
			return false
		}

		// Skip tombstones
		if cell.Tombstone {
			continue
		}

		// Found the key - mark as tombstone
		if cell.Key == key {
			cell.Tombstone = true
			ht.count--
			return true
		}
	}

	return false
}

// Contains checks if a key exists
func (ht *LinearProbingHashtable[K, V]) Contains(key K) bool {
	_, found := ht.Lookup(key)
	return found
}

// Size returns the number of elements
func (ht *LinearProbingHashtable[K, V]) Size() int {
	return ht.count
}

// Capacity returns the current buffer size
func (ht *LinearProbingHashtable[K, V]) Capacity() uint64 {
	return ht.grower.BufSize()
}

// LoadFactor returns the current load factor
func (ht *LinearProbingHashtable[K, V]) LoadFactor() float64 {
	return float64(ht.count) / float64(ht.grower.BufSize())
}

// Clear removes all elements
func (ht *LinearProbingHashtable[K, V]) Clear() {
	for i := range ht.cells {
		ht.cells[i] = Cell[K, V]{}
	}
	ht.count = 0
}

// resize grows the hashtable and rehashes all elements
func (ht *LinearProbingHashtable[K, V]) resize() {
	oldCells := ht.cells

	ht.grower.IncreaseSize()
	ht.cells = ht.allocator.Alloc(int(ht.grower.BufSize()))
	ht.count = 0

	// Rehash all existing elements
	for _, cell := range oldCells {
		if cell.Occupied && !cell.Tombstone {
			ht.Put(cell.Key, cell.Value)
		}
	}

	ht.allocator.Release(oldCells)
}

// Keys returns all keys in the hashtable
func (ht *LinearProbingHashtable[K, V]) Keys() []K {
	keys := make([]K, 0, ht.count)
	for _, cell := range ht.cells {
		if cell.Occupied && !cell.Tombstone {
			keys = append(keys, cell.Key)
		}
	}
	return keys
}

// Values returns all values in the hashtable
func (ht *LinearProbingHashtable[K, V]) Values() []V {
	values := make([]V, 0, ht.count)
	for _, cell := range ht.cells {
		if cell.Occupied && !cell.Tombstone {
			values = append(values, cell.Value)
		}
	}
	return values
}

// ForEach iterates over all key-value pairs
func (ht *LinearProbingHashtable[K, V]) ForEach(fn func(K, V)) {
	for _, cell := range ht.cells {
		if cell.Occupied && !cell.Tombstone {
			fn(cell.Key, cell.Value)
		}
	}
}

/****/

// StringHasher provides a hash function for string keys
func StringHasher(key string) uint64 {
	if len(key) == 0 {
		return 0
	}
	data := []byte(key)
	var h uint64
	if len(data) >= 8 {
		h = *(*uint64)(unsafe.Pointer(&data[0]))
	} else {
		for i := 0; i < len(data); i++ {
			h = h<<8 | uint64(data[i])
		}
	}
	return hash.MurmurFinalizerHash64(h)
}

// IntHasher provides a hash function for int keys
func IntHasher(key int) uint64 {
	return hash.MurmurFinalizerHash64(uint64(key))
}

// Int64Hasher provides a hash function for int64 keys
func Int64Hasher(key int64) uint64 {
	return hash.MurmurFinalizerHash64(uint64(key))
}

// Uint64Hasher provides a hash function for uint64 keys
func Uint64Hasher(key uint64) uint64 {
	return hash.MurmurFinalizerHash64(key)
}

/****/

// NewStringIntHashtable creates a string->int hashtable
func NewStringIntHashtable() *LinearProbingHashtable[string, int] {
	return New[string, int](StringHasher)
}

// NewIntIntHashtable creates an int->int hashtable
func NewIntIntHashtable() *LinearProbingHashtable[int, int] {
	return New[int, int](IntHasher)
}
