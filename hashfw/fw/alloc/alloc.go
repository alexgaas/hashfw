package alloc

import "sync"

// SliceAllocator implements HashtableAllocator using Go slices with optional pooling
type SliceAllocator[T any] struct {
	pool     sync.Pool
	capacity int
}

// NewSliceAllocator creates a new allocator with the given initial capacity
func NewSliceAllocator[T any](capacity int) *SliceAllocator[T] {
	return &SliceAllocator[T]{
		capacity: capacity,
		pool: sync.Pool{
			New: func() any {
				return make([]T, capacity)
			},
		},
	}
}

// alloc allocates a new slice of the given size
func (a *SliceAllocator[T]) alloc(size int) []T {
	if size <= a.capacity {
		slice := a.pool.Get().([]T)
		if len(slice) >= size {
			return slice[:size]
		}
	}
	return make([]T, size)
}

// release returns the slice to the pool for reuse
func (a *SliceAllocator[T]) release(slice []T) {
	if cap(slice) == a.capacity {
		// Clear the slice before returning to pool
		var zero T
		for i := range slice {
			slice[i] = zero
		}
		a.pool.Put(slice[:a.capacity])
	}
}

// Alloc is the exported version of alloc
func (a *SliceAllocator[T]) Alloc(size int) []T {
	return a.alloc(size)
}

// Release is the exported version of release
func (a *SliceAllocator[T]) Release(slice []T) {
	a.release(slice)
}

/****/

// SimpleAllocator is a basic allocator without pooling
type SimpleAllocator[T any] struct{}

// NewSimpleAllocator creates a new simple allocator
func NewSimpleAllocator[T any]() *SimpleAllocator[T] {
	return &SimpleAllocator[T]{}
}

func (a *SimpleAllocator[T]) alloc(size int) []T {
	return make([]T, size)
}

func (a *SimpleAllocator[T]) release(slice []T) {
	// No-op for simple allocator, let GC handle it
}

// Alloc is the exported version
func (a *SimpleAllocator[T]) Alloc(size int) []T {
	return a.alloc(size)
}

// Release is the exported version
func (a *SimpleAllocator[T]) Release(slice []T) {
	a.release(slice)
}
