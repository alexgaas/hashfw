package fw

import (
	"hashfw/hashfw/fw/alloc"
	"hashfw/hashfw/fw/hash"
)

// Cell /*
type Cell[K string | int, V any] struct {
	Key       K
	Value     V
	Tombstone bool
}

// HashtableGrower /*
type HashtableGrower interface {
	set(size uint64)
}

/****/

// HashtableOps /*
// Remove /*
/*
https://arxiv.org/pdf/1808.04602
Procedure delete(e)
	for i := h(e) while t[i] 6= e do 							–– search for e
		if t[i] = ⊥ then return 								–– e is not in t
	t[i] := † 													–– tombstone – may be removed later
	hˇ := m 													–– initialize the smallest hash function value encountered
	for j := i + 1 while t[j] 6= ⊥ 								–– scan to the right
		if t[j] 6= † then if h(t[j]) < hˇ then hˇ := h(t[j]) 	–– update smallest hash function value
	for k := i downto h(e) do 									–– scan to the left
		if t[j] = † then if h > k ˇ then t[j] := ⊥ 				–– remove tombstone
		else if h(t[j]) < hˇ then hˇ := h(t[j]) 				–– update smallest hash function value

*/
type HashtableOps[K string | int, V any] interface {
	lookup(k K) V
	put(k K, v V)
	remove(k K)
}

/****/

type Hashtable[K string | int, V any, FT uint64, FR uint64 | uint32] struct {
	Cells     []Cell[K, V]
	allocator alloc.HashtableAllocator[Cell[K, V]]
	finalizer hash.Finalizer[FT, FR]
	grower    HashtableGrower
	op        HashtableOps[K, V]
}
