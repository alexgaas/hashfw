## Hash Map Framework

A Go framework for building and comparing Open Addressing hash table implementations.

### Project Structure

```
hashfw/
├── hashfw/
│   ├── fw/                     # Framework core components
│   │   ├── hashtable.go        # Base interfaces and cell structure
│   │   ├── alloc/              # Memory allocation
│   │   │   ├── allocator.go    # Allocator interface
│   │   │   └── alloc.go        # Simple and pooled allocator implementations
│   │   ├── grower/             # Buffer sizing and growth
│   │   │   ├── grower.go       # Power-of-2 sizing with efficient masking
│   │   │   └── prefetcher.go   # CPU cache prefetch optimization
│   │   └── hash/               # Hash functions
│   │       ├── finalizer.go    # Finalizer interfaces
│   │       └── hash.go         # MurmurHash64 and IntHash32 implementations
│   ├── lp/                     # Linear Probing implementation
│   │   └── LinearProbingHashtable.go
│   └── rh/                     # Robin Hood implementation
│       └── RobinhoodHashtable.go
└── rhlp/                       # Benchmark and comparison
    ├── README.md               # Theory and results
    └── benchmark_test.go       # Performance comparison tests
```

### Features

**Framework (`fw/`):**
- Generic cell structure with tombstone support for deletion
- Pluggable allocators (simple GC-based or pooled)
- Power-of-2 sizing with efficient bit masking
- MurmurHash64 finalizer for well-distributed hashing

**Linear Probing (`lp/`):**
- Simple collision resolution by linear search
- Tombstone-based deletion
- Automatic resizing at 50% load factor

**Robin Hood Hashing (`rh/`):**
- PSL (Probe Sequence Length) tracking
- "Steal from the rich" insertion for balanced distribution
- Early termination on lookup misses
- Backward shift deletion (no tombstones needed)

### Usage

```go
// Linear Probing
import "hashfw/hashfw/lp"

ht := lp.NewStringIntHashtable()
ht.Put("key1", 100)
val, found := ht.Lookup("key1")
ht.Remove("key1")

// Robin Hood
import "hashfw/hashfw/rh"

ht := rh.NewStringIntHashtable()
ht.Put("key1", 100)
val, found := ht.Lookup("key1")
stats := ht.Stats() // Includes PSL statistics
```

### Custom Key Types

```go
// Custom hasher for any comparable type
ht := lp.New[MyType, int](func(key MyType) uint64 {
    return hash.MurmurFinalizerHash64(uint64(key.ID))
})
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run benchmarks
go test ./rhlp -bench=. -benchmem

# Run specific implementation tests
go test ./hashfw/lp -v
go test ./hashfw/rh -v
```

### Performance Comparison

See [rhlp/README.md](rhlp/README.md) for detailed theory and benchmark results comparing Linear Probing vs Robin Hood hashing.

**Summary:**
- Linear Probing: Faster inserts, simpler implementation
- Robin Hood: Better lookup performance at high load factors, more consistent probe lengths
