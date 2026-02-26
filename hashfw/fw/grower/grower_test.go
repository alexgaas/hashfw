package grower

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Grower", func() {
	var g *Grower

	BeforeEach(func() {
		g = NewGrower()
	})

	Describe("NewGrower", func() {
		It("should create a grower with initial size degree", func() {
			Expect(g).NotTo(BeNil())
			Expect(g.SizeDegree()).To(Equal(uint8(8)))
			Expect(g.BufSize()).To(Equal(uint64(256)))
		})
	})

	Describe("NewGrowerWithSize", func() {
		It("should create a grower sized for expected elements", func() {
			g := NewGrowerWithSize(1000)

			Expect(g).NotTo(BeNil())
			Expect(g.BufSize()).To(BeNumerically(">=", uint64(1000)))
		})
	})

	Describe("BufSize", func() {
		It("should return the current buffer size as power of 2", func() {
			Expect(g.BufSize()).To(Equal(uint64(256)))
		})
	})

	Describe("Place", func() {
		It("should map hash to valid index within buffer bounds", func() {
			pos := g.Place(12345)

			Expect(pos).To(BeNumerically("<", g.BufSize()))
		})

		It("should return consistent results for same hash", func() {
			pos1 := g.Place(12345)
			pos2 := g.Place(12345)

			Expect(pos1).To(Equal(pos2))
		})

		It("should use mask for efficient modulo", func() {
			hash := uint64(0xFFFFFFFFFFFFFFFF)
			pos := g.Place(hash)

			Expect(pos).To(BeNumerically("<", g.BufSize()))
			Expect(pos).To(Equal(hash & g.Mask()))
		})
	})

	Describe("Next", func() {
		It("should wrap around at buffer boundary", func() {
			pos := g.BufSize() - 1
			next := g.Next(pos)

			Expect(next).To(Equal(uint64(0)))
		})

		It("should increment position normally", func() {
			next := g.Next(5)

			Expect(next).To(Equal(uint64(6)))
		})
	})

	Describe("Overflow", func() {
		It("should not overflow with few elements", func() {
			Expect(g.Overflow(10)).To(BeFalse())
		})

		It("should not overflow at exactly 50% load factor", func() {
			halfCapacity := int(g.BufSize() / 2)
			Expect(g.Overflow(halfCapacity)).To(BeFalse())
		})

		It("should overflow when exceeding 50% load factor", func() {
			halfCapacity := int(g.BufSize() / 2)
			Expect(g.Overflow(halfCapacity + 1)).To(BeTrue())
		})
	})

	Describe("IncreaseSize", func() {
		It("should increase buffer size by factor of 4", func() {
			initialSize := g.BufSize()
			g.IncreaseSize()

			Expect(g.BufSize()).To(Equal(initialSize * 4))
		})
	})

	Describe("Set", func() {
		It("should size appropriately for given element count", func() {
			g.Set(1000)

			Expect(g.BufSize()).To(BeNumerically(">=", uint64(1000)))
		})

		It("should reset to initial size for small counts", func() {
			g.Set(1)

			Expect(g.BufSize()).To(Equal(uint64(256)))
		})
	})

	Describe("SetBufSize", func() {
		It("should set size based on buffer size", func() {
			g.SetBufSize(1024)

			Expect(g.BufSize()).To(BeNumerically(">=", uint64(1024)))
		})
	})

	Describe("Mask", func() {
		It("should equal BufSize minus 1", func() {
			Expect(g.Mask()).To(Equal(g.BufSize() - 1))
		})
	})
})
