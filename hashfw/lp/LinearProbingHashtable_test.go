package lp_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"hashfw/hashfw/lp"
)

func TestLinearProbing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linear Probing Hashtable Suite")
}

var _ = Describe("LinearProbingHashtable", func() {
	var ht *lp.LinearProbingHashtable[string, int]

	BeforeEach(func() {
		ht = lp.NewStringIntHashtable()
	})

	Describe("Basic Operations", func() {
		Context("Put and Lookup", func() {
			It("should store and retrieve values", func() {
				ht.Put("key1", 100)
				ht.Put("key2", 200)
				ht.Put("key3", 300)

				val, found := ht.Lookup("key1")
				Expect(found).To(BeTrue())
				Expect(val).To(Equal(100))

				val, found = ht.Lookup("key2")
				Expect(found).To(BeTrue())
				Expect(val).To(Equal(200))

				val, found = ht.Lookup("key3")
				Expect(found).To(BeTrue())
				Expect(val).To(Equal(300))
			})

			It("should return false for nonexistent keys", func() {
				ht.Put("key1", 100)

				_, found := ht.Lookup("nonexistent")
				Expect(found).To(BeFalse())
			})
		})

		Context("Update", func() {
			It("should update existing keys", func() {
				ht.Put("key1", 100)
				val, _ := ht.Lookup("key1")
				Expect(val).To(Equal(100))

				ht.Put("key1", 999)
				val, found := ht.Lookup("key1")
				Expect(found).To(BeTrue())
				Expect(val).To(Equal(999))
			})

			It("should not increase size when updating", func() {
				ht.Put("key1", 100)
				ht.Put("key1", 999)

				Expect(ht.Size()).To(Equal(1))
			})
		})

		Context("Remove", func() {
			It("should remove existing keys", func() {
				ht.Put("key1", 100)
				ht.Put("key2", 200)
				ht.Put("key3", 300)

				Expect(ht.Size()).To(Equal(3))

				removed := ht.Remove("key2")
				Expect(removed).To(BeTrue())
				Expect(ht.Size()).To(Equal(2))

				_, found := ht.Lookup("key2")
				Expect(found).To(BeFalse())
			})

			It("should preserve other keys after removal", func() {
				ht.Put("key1", 100)
				ht.Put("key2", 200)
				ht.Put("key3", 300)

				ht.Remove("key2")

				val, found := ht.Lookup("key1")
				Expect(found).To(BeTrue())
				Expect(val).To(Equal(100))

				val, found = ht.Lookup("key3")
				Expect(found).To(BeTrue())
				Expect(val).To(Equal(300))
			})

			It("should return false for nonexistent keys", func() {
				removed := ht.Remove("nonexistent")
				Expect(removed).To(BeFalse())
			})
		})

		Context("Contains", func() {
			It("should return true for existing keys", func() {
				ht.Put("key1", 100)
				Expect(ht.Contains("key1")).To(BeTrue())
			})

			It("should return false for nonexistent keys", func() {
				Expect(ht.Contains("key2")).To(BeFalse())
			})
		})
	})

	Describe("Resize", func() {
		It("should resize when capacity is exceeded", func() {
			initialCapacity := ht.Capacity()

			for i := 0; i < 200; i++ {
				ht.Put(fmt.Sprintf("key%d", i), i)
			}

			Expect(ht.Capacity()).To(BeNumerically(">", initialCapacity))
			Expect(ht.Size()).To(Equal(200))
		})

		It("should preserve all elements after resize", func() {
			for i := 0; i < 200; i++ {
				ht.Put(fmt.Sprintf("key%d", i), i)
			}

			for i := 0; i < 200; i++ {
				val, found := ht.Lookup(fmt.Sprintf("key%d", i))
				Expect(found).To(BeTrue(), "key%d should exist", i)
				Expect(val).To(Equal(i))
			}
		})
	})

	Describe("Clear", func() {
		It("should remove all elements", func() {
			ht.Put("key1", 100)
			ht.Put("key2", 200)

			Expect(ht.Size()).To(Equal(2))

			ht.Clear()

			Expect(ht.Size()).To(Equal(0))
			Expect(ht.Contains("key1")).To(BeFalse())
			Expect(ht.Contains("key2")).To(BeFalse())
		})
	})

	Describe("Keys", func() {
		It("should return all keys", func() {
			ht.Put("key1", 100)
			ht.Put("key2", 200)
			ht.Put("key3", 300)

			keys := ht.Keys()
			Expect(keys).To(HaveLen(3))
			Expect(keys).To(ContainElements("key1", "key2", "key3"))
		})
	})

	Describe("Values", func() {
		It("should return all values", func() {
			ht.Put("key1", 100)
			ht.Put("key2", 200)
			ht.Put("key3", 300)

			values := ht.Values()
			Expect(values).To(HaveLen(3))
			Expect(values).To(ContainElements(100, 200, 300))
		})
	})

	Describe("ForEach", func() {
		It("should iterate over all key-value pairs", func() {
			ht.Put("key1", 100)
			ht.Put("key2", 200)
			ht.Put("key3", 300)

			visited := make(map[string]int)
			ht.ForEach(func(k string, v int) {
				visited[k] = v
			})

			Expect(visited).To(HaveLen(3))
			Expect(visited["key1"]).To(Equal(100))
			Expect(visited["key2"]).To(Equal(200))
			Expect(visited["key3"]).To(Equal(300))
		})
	})

	Describe("LoadFactor", func() {
		It("should be zero when empty", func() {
			Expect(ht.LoadFactor()).To(Equal(0.0))
		})

		It("should increase as elements are added", func() {
			ht.Put("key1", 100)
			lf := ht.LoadFactor()

			Expect(lf).To(BeNumerically(">", 0.0))
			Expect(lf).To(BeNumerically("<", 1.0))
		})
	})

	Describe("Collision Handling", func() {
		It("should handle keys with same hash", func() {
			// Use a simple hash function that causes collisions
			ht := lp.New[string, int](func(key string) uint64 {
				return uint64(len(key))
			})

			// These all have the same length, so same hash
			ht.Put("aaa", 1)
			ht.Put("bbb", 2)
			ht.Put("ccc", 3)

			val, found := ht.Lookup("aaa")
			Expect(found).To(BeTrue())
			Expect(val).To(Equal(1))

			val, found = ht.Lookup("bbb")
			Expect(found).To(BeTrue())
			Expect(val).To(Equal(2))

			val, found = ht.Lookup("ccc")
			Expect(found).To(BeTrue())
			Expect(val).To(Equal(3))
		})
	})

	Describe("Remove and Reinsert", func() {
		It("should allow reinsertion after removal", func() {
			ht.Put("key1", 100)
			ht.Put("key2", 200)

			ht.Remove("key1")
			Expect(ht.Contains("key1")).To(BeFalse())

			ht.Put("key1", 999)
			val, found := ht.Lookup("key1")
			Expect(found).To(BeTrue())
			Expect(val).To(Equal(999))
		})
	})

	Describe("Empty Key", func() {
		It("should handle empty string as key", func() {
			ht.Put("", 100)

			val, found := ht.Lookup("")
			Expect(found).To(BeTrue())
			Expect(val).To(Equal(100))
		})
	})
})

var _ = Describe("LinearProbingHashtable with Int Keys", func() {
	It("should work with integer keys", func() {
		ht := lp.NewIntIntHashtable()

		ht.Put(1, 100)
		ht.Put(2, 200)
		ht.Put(3, 300)

		val, found := ht.Lookup(1)
		Expect(found).To(BeTrue())
		Expect(val).To(Equal(100))

		val, found = ht.Lookup(2)
		Expect(found).To(BeTrue())
		Expect(val).To(Equal(200))
	})
})

// Benchmarks (kept as standard Go benchmarks)
func BenchmarkLinearProbingHashtable_Put(b *testing.B) {
	ht := lp.NewStringIntHashtable()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Put(fmt.Sprintf("key%d", i), i)
	}
}

func BenchmarkLinearProbingHashtable_Lookup(b *testing.B) {
	ht := lp.NewStringIntHashtable()

	for i := 0; i < 10000; i++ {
		ht.Put(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(fmt.Sprintf("key%d", i%10000))
	}
}
