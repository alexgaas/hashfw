package alloc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"hashfw/hashfw/fw/alloc"
)

func TestAlloc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Allocator Suite")
}

var _ = Describe("SimpleAllocator", func() {
	var allocator *alloc.SimpleAllocator[int]

	BeforeEach(func() {
		allocator = alloc.NewSimpleAllocator[int]()
	})

	Describe("Alloc", func() {
		It("should allocate a slice of the requested size", func() {
			slice := allocator.Alloc(100)

			Expect(slice).NotTo(BeNil())
			Expect(slice).To(HaveLen(100))
		})

		It("should allocate slices with zero values", func() {
			slice := allocator.Alloc(10)

			for i := range slice {
				Expect(slice[i]).To(Equal(0))
			}
		})
	})

	Describe("Release", func() {
		It("should not panic when releasing a slice", func() {
			slice := allocator.Alloc(100)

			Expect(func() {
				allocator.Release(slice)
			}).NotTo(Panic())
		})
	})
})

var _ = Describe("SliceAllocator", func() {
	var allocator *alloc.SliceAllocator[int]
	const poolCapacity = 256

	BeforeEach(func() {
		allocator = alloc.NewSliceAllocator[int](poolCapacity)
	})

	Describe("Alloc", func() {
		It("should allocate a slice of the requested size", func() {
			slice := allocator.Alloc(100)

			Expect(slice).NotTo(BeNil())
			Expect(slice).To(HaveLen(100))
		})

		It("should allocate larger slices than pool capacity", func() {
			slice := allocator.Alloc(500)

			Expect(slice).NotTo(BeNil())
			Expect(slice).To(HaveLen(500))
		})
	})

	Describe("Release", func() {
		It("should allow reallocation after release", func() {
			slice := allocator.Alloc(100)
			allocator.Release(slice)

			slice2 := allocator.Alloc(100)
			Expect(slice2).NotTo(BeNil())
		})

		It("should clear the slice when returning to pool", func() {
			slice := allocator.Alloc(poolCapacity)
			for i := range slice {
				slice[i] = i
			}
			allocator.Release(slice)

			slice2 := allocator.Alloc(poolCapacity)
			for i := range slice2 {
				Expect(slice2[i]).To(Equal(0), "reused slice should be cleared")
			}
		})

		It("should not panic when releasing wrong size slice", func() {
			slice := allocator.Alloc(500)

			Expect(func() {
				allocator.Release(slice)
			}).NotTo(Panic())
		})
	})
})

type TestStruct struct {
	ID    int
	Name  string
	Value float64
}

var _ = Describe("Allocator with Structs", func() {
	Describe("SliceAllocator with struct type", func() {
		It("should allocate and use struct slices", func() {
			allocator := alloc.NewSliceAllocator[TestStruct](100)

			slice := allocator.Alloc(50)
			Expect(slice).To(HaveLen(50))

			slice[0] = TestStruct{ID: 1, Name: "test", Value: 1.5}
			Expect(slice[0].ID).To(Equal(1))
			Expect(slice[0].Name).To(Equal("test"))
		})
	})

	Describe("SimpleAllocator with struct type", func() {
		It("should allocate and use struct slices", func() {
			allocator := alloc.NewSimpleAllocator[TestStruct]()

			slice := allocator.Alloc(50)
			Expect(slice).To(HaveLen(50))

			slice[0] = TestStruct{ID: 1, Name: "test", Value: 1.5}
			Expect(slice[0].ID).To(Equal(1))
		})
	})
})
