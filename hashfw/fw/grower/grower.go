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
