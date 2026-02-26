package grower

import "math"

var initialSizeDegree uint8 = 8
var sizeDegree = initialSizeDegree
var precalculatedMask uint64 = (1 << sizeDegree) - 1
var precalculatedMaxFill = 1 << (sizeDegree - 1)
var maxSizeDegree uint8 = 23

func getSizeDegree() uint8 {
	return sizeDegree
}

func increaseSizeDegree(delta uint8) {
	sizeDegree += delta
	precalculatedMask = (1 << sizeDegree) - 1
	precalculatedMaxFill = 1 << (sizeDegree - 1)
}

var initialCount = 1 << sizeDegree
var performsLinearProbingWithSingleStep = true

func bufSize() uint64 {
	return 1 << sizeDegree
}

func place(x uint64) uint64 {
	return x & precalculatedMask
}

func next(pos uint64) uint64 {
	return (pos + 1) & precalculatedMask
}

func overflow(elems int) bool {
	return elems > precalculatedMaxFill
}

func increaseSize() {
	var res uint8 = 2
	if sizeDegree >= maxSizeDegree {
		res = 1
	}
	increaseSizeDegree(res)
}

func set(numElems uint64) {
	if numElems <= 1 {
		sizeDegree = initialSizeDegree
	} else if sizeDegree > uint8(math.Log2(float64(numElems-1))+2) {
		sizeDegree = uint8(math.Log2(float64(numElems-1)) + 2)
	}
	increaseSizeDegree(0)
}

func setBufSize(bufSize uint64) {
	sizeDegree = uint8(math.Log2(float64(bufSize-1)) + 1)
	increaseSizeDegree(0)
}

/****/

// Grower encapsulates sizing logic for a single hashtable instance
type Grower struct {
	sizeDegree          uint8
	precalculatedMask   uint64
	precalculatedMaxFill int
}

// NewGrower creates a new Grower with initial size
func NewGrower() *Grower {
	g := &Grower{
		sizeDegree: initialSizeDegree,
	}
	g.recalculate()
	return g
}

// NewGrowerWithSize creates a Grower sized for expected elements
func NewGrowerWithSize(expectedElements uint64) *Grower {
	g := NewGrower()
	g.Set(expectedElements)
	return g
}

func (g *Grower) recalculate() {
	g.precalculatedMask = (1 << g.sizeDegree) - 1
	g.precalculatedMaxFill = 1 << (g.sizeDegree - 1)
}

// BufSize returns the current buffer size
func (g *Grower) BufSize() uint64 {
	return 1 << g.sizeDegree
}

// Place maps a hash value to an index in the buffer
func (g *Grower) Place(x uint64) uint64 {
	return x & g.precalculatedMask
}

// Next returns the next position with wraparound
func (g *Grower) Next(pos uint64) uint64 {
	return (pos + 1) & g.precalculatedMask
}

// Overflow checks if the number of elements exceeds load factor threshold
func (g *Grower) Overflow(elems int) bool {
	return elems > g.precalculatedMaxFill
}

// IncreaseSize grows the buffer
func (g *Grower) IncreaseSize() {
	var delta uint8 = 2
	if g.sizeDegree >= maxSizeDegree {
		delta = 1
	}
	g.sizeDegree += delta
	g.recalculate()
}

// Set calculates appropriate size for given element count
func (g *Grower) Set(numElems uint64) {
	if numElems <= 1 {
		g.sizeDegree = initialSizeDegree
	} else {
		degree := uint8(math.Log2(float64(numElems-1)) + 2)
		if degree > g.sizeDegree {
			g.sizeDegree = degree
		}
	}
	g.recalculate()
}

// SetBufSize sets size directly from buffer size
func (g *Grower) SetBufSize(bufSize uint64) {
	g.sizeDegree = uint8(math.Log2(float64(bufSize-1)) + 1)
	g.recalculate()
}

// SizeDegree returns the current size degree
func (g *Grower) SizeDegree() uint8 {
	return g.sizeDegree
}

// Mask returns the current mask
func (g *Grower) Mask() uint64 {
	return g.precalculatedMask
}
