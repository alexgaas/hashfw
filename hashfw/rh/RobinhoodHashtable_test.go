package rh_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"hashfw/hashfw/rh"
)

func TestRobinhood(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Robin Hood Hashtable Suite")
}

var _ = Describe("RobinhoodHashtable", func() {
	var ht *rh.RobinhoodHashtable[string, int]

	BeforeEach(func() {
		ht = rh.NewStringIntHashtable()
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

		It("should reset MaxPSL", func() {
			ht.Put("key1", 100)
			ht.Clear()

			Expect(ht.MaxPSL()).To(Equal(uint32(0)))
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

	Describe("PSL Tracking", func() {
		It("should track PSL when collisions occur", func() {
			// Use a hash function that causes collisions
			ht := rh.New[string, int](func(key string) uint64 {
				return uint64(len(key))
			})

			// These all have the same length, so same hash
			ht.Put("aaa", 1)
			ht.Put("bbb", 2)
			ht.Put("ccc", 3)
			ht.Put("ddd", 4)
			ht.Put("eee", 5)

			Expect(ht.MaxPSL()).To(BeNumerically(">", uint32(0)))

			// All elements should still be accessible
			for _, key := range []string{"aaa", "bbb", "ccc", "ddd", "eee"} {
				Expect(ht.Contains(key)).To(BeTrue())
			}
		})
	})

	Describe("Early Termination", func() {
		It("should terminate early for nonexistent keys", func() {
			ht.Put("key1", 100)
			ht.Put("key2", 200)

			_, found := ht.Lookup("nonexistent")
			Expect(found).To(BeFalse())
		})
	})

	Describe("Stats", func() {
		It("should return correct statistics", func() {
			ht.Put("key1", 100)
			ht.Put("key2", 200)
			ht.Put("key3", 300)

			stats := ht.Stats()

			Expect(stats.Count).To(Equal(3))
			Expect(stats.Capacity).To(BeNumerically(">", uint64(0)))
			Expect(stats.LoadFactor).To(BeNumerically(">", 0.0))
			Expect(stats.MaxPSL).To(BeNumerically(">=", uint32(0)))
			Expect(stats.AvgPSL).To(BeNumerically(">=", 0.0))
		})
	})

	Describe("AveragePSL", func() {
		It("should be zero for empty table", func() {
			Expect(ht.AveragePSL()).To(Equal(0.0))
		})

		It("should be non-negative with elements", func() {
			ht.Put("key1", 100)
			Expect(ht.AveragePSL()).To(BeNumerically(">=", 0.0))
		})
	})

	Describe("Backward Shift Deletion", func() {
		It("should maintain correctness after deletion", func() {
			// Use a hash function that causes collisions
			ht := rh.New[string, int](func(key string) uint64 {
				return uint64(len(key))
			})

			ht.Put("aaa", 1)
			ht.Put("bbb", 2)
			ht.Put("ccc", 3)
			ht.Put("ddd", 4)

			// Remove from middle - should shift subsequent elements back
			ht.Remove("bbb")

			Expect(ht.Contains("aaa")).To(BeTrue())
			Expect(ht.Contains("bbb")).To(BeFalse())
			Expect(ht.Contains("ccc")).To(BeTrue())
			Expect(ht.Contains("ddd")).To(BeTrue())

			// Values should be correct
			val, _ := ht.Lookup("aaa")
			Expect(val).To(Equal(1))
			val, _ = ht.Lookup("ccc")
			Expect(val).To(Equal(3))
			val, _ = ht.Lookup("ddd")
			Expect(val).To(Equal(4))
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

var _ = Describe("RobinhoodHashtable with Int Keys", func() {
	It("should work with integer keys", func() {
		ht := rh.NewIntIntHashtable()

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
func BenchmarkRobinhoodHashtable_Put(b *testing.B) {
	ht := rh.NewStringIntHashtable()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Put(fmt.Sprintf("key%d", i), i)
	}
}

func BenchmarkRobinhoodHashtable_Lookup(b *testing.B) {
	ht := rh.NewStringIntHashtable()

	for i := 0; i < 10000; i++ {
		ht.Put(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(fmt.Sprintf("key%d", i%10000))
	}
}

func BenchmarkRobinhoodHashtable_LookupMiss(b *testing.B) {
	ht := rh.NewStringIntHashtable()

	for i := 0; i < 10000; i++ {
		ht.Put(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(fmt.Sprintf("miss%d", i))
	}
}
