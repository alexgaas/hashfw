package rh

import (
	"hashfw/hashfw/fw/alloc"
	"hashfw/hashfw/fw/grower"
	"hashfw/hashfw/fw/hash"
	"unsafe"
)

// Cell represents a single cell in the Robin Hood hashtable with PSL tracking
type Cell[K comparable, V any] struct {
	Key      K
	Value    V
	PSL      uint32 // Probe Sequence Length - distance from ideal position
	Occupied bool
}

// RobinhoodHashtable implements Open Addressing with Robin Hood collision resolution
// Robin Hood hashing reduces variance in probe lengths by "stealing" from rich entries
type RobinhoodHashtable[K comparable, V any] struct {
	cells     []Cell[K, V]
	grower    *grower.Grower
	allocator *alloc.SimpleAllocator[Cell[K, V]]
	count     int
	maxPSL    uint32 // Track maximum PSL for statistics
	hashFunc  func(K) uint64
}

// New creates a new RobinhoodHashtable
func New[K comparable, V any](hashFunc func(K) uint64) *RobinhoodHashtable[K, V] {
	g := grower.NewGrower()
	alloc := alloc.NewSimpleAllocator[Cell[K, V]]()

	return &RobinhoodHashtable[K, V]{
		cells:     alloc.Alloc(int(g.BufSize())),
		grower:    g,
		allocator: alloc,
		count:     0,
		maxPSL:    0,
		hashFunc:  hashFunc,
	}
}

// NewWithCapacity creates a hashtable with expected capacity
func NewWithCapacity[K comparable, V any](capacity uint64, hashFunc func(K) uint64) *RobinhoodHashtable[K, V] {
	g := grower.NewGrowerWithSize(capacity)
	alloc := alloc.NewSimpleAllocator[Cell[K, V]]()

	return &RobinhoodHashtable[K, V]{
		cells:     alloc.Alloc(int(g.BufSize())),
		grower:    g,
		allocator: alloc,
		count:     0,
		maxPSL:    0,
		hashFunc:  hashFunc,
	}
}

// Put inserts or updates a key-value pair using Robin Hood insertion
func (ht *RobinhoodHashtable[K, V]) Put(key K, value V) {
	// Check if resize needed
	if ht.grower.Overflow(ht.count + 1) {
		ht.resize()
	}

	h := ht.hashFunc(key)
	pos := ht.grower.Place(h)
	bufSize := ht.grower.BufSize()

	insertKey := key
	insertValue := value
	insertPSL := uint32(0)

	for i := uint64(0); i < bufSize; i++ {
		idx := (pos + i) & ht.grower.Mask()
		cell := &ht.cells[idx]

		// Found empty slot - insert here
		if !cell.Occupied {
			cell.Key = insertKey
			cell.Value = insertValue
			cell.PSL = insertPSL
			cell.Occupied = true
			ht.count++
			if insertPSL > ht.maxPSL {
				ht.maxPSL = insertPSL
			}
			return
		}

		// Key already exists - update value
		if cell.Key == insertKey {
			cell.Value = insertValue
			return
		}

		// Robin Hood: steal from the rich (swap if current entry has lower PSL)
		if cell.PSL < insertPSL {
			// Swap current entry with the one we're inserting
			insertKey, cell.Key = cell.Key, insertKey
			insertValue, cell.Value = cell.Value, insertValue
			insertPSL, cell.PSL = cell.PSL, insertPSL
		}

		insertPSL++
	}
}

// Lookup retrieves a value by key with early termination based on PSL
func (ht *RobinhoodHashtable[K, V]) Lookup(key K) (V, bool) {
	h := ht.hashFunc(key)
	pos := ht.grower.Place(h)
	bufSize := ht.grower.BufSize()

	for psl := uint32(0); uint64(psl) < bufSize; psl++ {
		idx := (pos + uint64(psl)) & ht.grower.Mask()
		cell := &ht.cells[idx]

		// Empty slot - key not found
		if !cell.Occupied {
			var zero V
			return zero, false
		}

		// Early termination: if current cell's PSL < our search length,
		// the key cannot exist further (Robin Hood invariant)
		if cell.PSL < psl {
			var zero V
			return zero, false
		}

		// Found the key
		if cell.Key == key {
			return cell.Value, true
		}
	}

	var zero V
	return zero, false
}

// Remove deletes a key from the hashtable using backward shift deletion
func (ht *RobinhoodHashtable[K, V]) Remove(key K) bool {
	h := ht.hashFunc(key)
	pos := ht.grower.Place(h)
	bufSize := ht.grower.BufSize()

	// Find the key
	var foundIdx uint64
	found := false

	for psl := uint32(0); uint64(psl) < bufSize; psl++ {
		idx := (pos + uint64(psl)) & ht.grower.Mask()
		cell := &ht.cells[idx]

		if !cell.Occupied {
			return false
		}

		if cell.PSL < psl {
			return false
		}

		if cell.Key == key {
			foundIdx = idx
			found = true
			break
		}
	}

	if !found {
		return false
	}

	// Backward shift deletion: shift subsequent entries back
	ht.cells[foundIdx] = Cell[K, V]{}

	for i := uint64(1); i < bufSize; i++ {
		nextIdx := (foundIdx + i) & ht.grower.Mask()
		nextCell := &ht.cells[nextIdx]

		// Stop if empty or at home position (PSL == 0)
		if !nextCell.Occupied || nextCell.PSL == 0 {
			break
		}

		// Shift entry back
		prevIdx := (foundIdx + i - 1) & ht.grower.Mask()
		ht.cells[prevIdx] = *nextCell
		ht.cells[prevIdx].PSL--
		ht.cells[nextIdx] = Cell[K, V]{}
	}

	ht.count--
	return true
}

// Contains checks if a key exists
func (ht *RobinhoodHashtable[K, V]) Contains(key K) bool {
	_, found := ht.Lookup(key)
	return found
}

// Size returns the number of elements
func (ht *RobinhoodHashtable[K, V]) Size() int {
	return ht.count
}

// Capacity returns the current buffer size
func (ht *RobinhoodHashtable[K, V]) Capacity() uint64 {
	return ht.grower.BufSize()
}

// LoadFactor returns the current load factor
func (ht *RobinhoodHashtable[K, V]) LoadFactor() float64 {
	return float64(ht.count) / float64(ht.grower.BufSize())
}

// MaxPSL returns the maximum probe sequence length
func (ht *RobinhoodHashtable[K, V]) MaxPSL() uint32 {
	return ht.maxPSL
}

// AveragePSL calculates the average probe sequence length
func (ht *RobinhoodHashtable[K, V]) AveragePSL() float64 {
	if ht.count == 0 {
		return 0
	}
	var total uint64
	for _, cell := range ht.cells {
		if cell.Occupied {
			total += uint64(cell.PSL)
		}
	}
	return float64(total) / float64(ht.count)
}

// Clear removes all elements
func (ht *RobinhoodHashtable[K, V]) Clear() {
	for i := range ht.cells {
		ht.cells[i] = Cell[K, V]{}
	}
	ht.count = 0
	ht.maxPSL = 0
}

// resize grows the hashtable and rehashes all elements
func (ht *RobinhoodHashtable[K, V]) resize() {
	oldCells := ht.cells

	ht.grower.IncreaseSize()
	ht.cells = ht.allocator.Alloc(int(ht.grower.BufSize()))
	ht.count = 0
	ht.maxPSL = 0

	// Rehash all existing elements
	for _, cell := range oldCells {
		if cell.Occupied {
			ht.Put(cell.Key, cell.Value)
		}
	}

	ht.allocator.Release(oldCells)
}

// Keys returns all keys in the hashtable
func (ht *RobinhoodHashtable[K, V]) Keys() []K {
	keys := make([]K, 0, ht.count)
	for _, cell := range ht.cells {
		if cell.Occupied {
			keys = append(keys, cell.Key)
		}
	}
	return keys
}

// Values returns all values in the hashtable
func (ht *RobinhoodHashtable[K, V]) Values() []V {
	values := make([]V, 0, ht.count)
	for _, cell := range ht.cells {
		if cell.Occupied {
			values = append(values, cell.Value)
		}
	}
	return values
}

// ForEach iterates over all key-value pairs
func (ht *RobinhoodHashtable[K, V]) ForEach(fn func(K, V)) {
	for _, cell := range ht.cells {
		if cell.Occupied {
			fn(cell.Key, cell.Value)
		}
	}
}

// Stats returns statistics about the hashtable
type Stats struct {
	Count      int
	Capacity   uint64
	LoadFactor float64
	MaxPSL     uint32
	AvgPSL     float64
}

func (ht *RobinhoodHashtable[K, V]) Stats() Stats {
	return Stats{
		Count:      ht.count,
		Capacity:   ht.grower.BufSize(),
		LoadFactor: ht.LoadFactor(),
		MaxPSL:     ht.maxPSL,
		AvgPSL:     ht.AveragePSL(),
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
func NewStringIntHashtable() *RobinhoodHashtable[string, int] {
	return New[string, int](StringHasher)
}

// NewIntIntHashtable creates an int->int hashtable
func NewIntIntHashtable() *RobinhoodHashtable[int, int] {
	return New[int, int](IntHasher)
}
