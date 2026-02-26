package grower

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prefetcher", func() {
	Describe("clamp", func() {
		DescribeTable("should clamp values correctly",
			func(value, min, max, expected float64) {
				result := clamp(value, min, max)
				Expect(result).To(Equal(expected))
			},
			Entry("value within range", 5.0, 1.0, 10.0, 5.0),
			Entry("value below min", 0.5, 1.0, 10.0, 1.0),
			Entry("value above max", 15.0, 1.0, 10.0, 10.0),
			Entry("value equals min", 1.0, 1.0, 10.0, 1.0),
			Entry("value equals max", 10.0, 1.0, 10.0, 10.0),
			Entry("negative range", -5.0, -10.0, -1.0, -5.0),
			Entry("negative below min", -15.0, -10.0, -1.0, -10.0),
		)
	})

	Describe("calcPrefetchLookAhead", func() {
		It("should return value within defined bounds", func() {
			result := calcPrefetchLookAhead()

			Expect(result).To(BeNumerically(">=", uint64(minLookAheadValue)))
			Expect(result).To(BeNumerically("<=", uint64(maxLookAheadValue)))
		})

		It("should return consistent valid results across multiple calls", func() {
			for i := 0; i < 10; i++ {
				result := calcPrefetchLookAhead()

				Expect(result).To(BeNumerically(">=", uint64(minLookAheadValue)))
				Expect(result).To(BeNumerically("<=", uint64(maxLookAheadValue)))
			}
		})
	})

	Describe("Constants", func() {
		It("should have correct default values", func() {
			Expect(iterationsToMeasure).To(Equal(100))
			Expect(minLookAheadValue).To(Equal(4))
			Expect(maxLookAheadValue).To(Equal(32))
			Expect(assumedLoadLatencyNs).To(Equal(100))
			Expect(justCoefficient).To(Equal(4))
		})
	})
})
