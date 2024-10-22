package hash

// Finalizer /*
type Finalizer[T any, R any] interface {
	finalizer([]T) R
}

/****/

type Murmur3Finalizer Finalizer[string, uint64]

type IntFinalizer Finalizer[uint64, uint32]
