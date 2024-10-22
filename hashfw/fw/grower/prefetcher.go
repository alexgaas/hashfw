package grower

import (
	"math"
	"time"
)

/**
 * The purpose of this helper class is to provide a good value for prefetch look ahead (how distant row we should prefetch on the given iteration)
 * based on the latency of a single iteration of the given cycle.
 *
 * Assumed usage pattern is the following:
 *
 * PrefetchingHelper prefetching; /// When object is created, it starts a watch to measure iteration latency.
 * size_t prefetch_look_ahead = prefetching.getInitialLookAheadValue(); // Initially it provides you with some reasonable default value.
 *
 * for (i = 0; i < end; ++i) {
 *     if i == prefetching.iterationsToMeasure() // When enough iterations passed, we are able to make a fairly accurate estimation of a single iteration latency.
 *         prefetch_look_ahead = prefetching.calcPrefetchLookAhead() // Based on this estimation we can choose a good value for prefetch_look_ahead.
 *
 *     ... main loop body ...
 * }
 *
 */

const (
	iterationsToMeasure  = 100
	minLookAheadValue    = 4
	maxLookAheadValue    = 32
	assumedLoadLatencyNs = 100
	justCoefficient      = 4
)

func calcPrefetchLookAhead() uint64 {
	start := time.Now()
	var singleIterationLatency = math.Max(float64(time.Since(start).Nanoseconds()/iterationsToMeasure), 1.)
	return uint64(clamp(
		math.Ceil(justCoefficient*assumedLoadLatencyNs/singleIterationLatency),
		minLookAheadValue,
		maxLookAheadValue,
	))
}

func clamp(d float64, min float64, max float64) float64 {
	t := d
	if d < min {
		t = min
	}

	if t > max {
		return max
	} else {
		return t
	}
}
