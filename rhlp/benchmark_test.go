package rhlp_test

import (
	"fmt"
	"math/rand"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"hashfw/hashfw/lp"
	"hashfw/hashfw/rh"
)

func TestRHLP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Robin Hood vs Linear Probing Suite")
}

// generateKeys creates a slice of random string keys
func generateKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("key_%d_%d", i, rand.Int63())
	}
	return keys
}

var _ = Describe("PSL Distribution Comparison", func() {
	const numElements = 1000

	It("should show Robin Hood maintains low PSL variance", func() {
		ht := rh.NewStringIntHashtable()
		keys := generateKeys(numElements)

		for i, key := range keys {
			ht.Put(key, i)
		}

		stats := ht.Stats()

		GinkgoWriter.Printf("Robin Hood Stats:\n")
		GinkgoWriter.Printf("  Count: %d\n", stats.Count)
		GinkgoWriter.Printf("  Capacity: %d\n", stats.Capacity)
		GinkgoWriter.Printf("  Load Factor: %.2f%%\n", stats.LoadFactor*100)
		GinkgoWriter.Printf("  Max PSL: %d\n", stats.MaxPSL)
		GinkgoWriter.Printf("  Avg PSL: %.2f\n", stats.AvgPSL)

		Expect(stats.Count).To(Equal(numElements))
		Expect(stats.MaxPSL).To(BeNumerically("<", uint32(20)))
		Expect(stats.AvgPSL).To(BeNumerically("<", 5.0))
	})
})

var _ = Describe("Functional Comparison", func() {
	Describe("Linear Probing vs Robin Hood", func() {
		const numElements = 500

		Context("Insert operations", func() {
			It("should insert same elements correctly in both implementations", func() {
				lpHT := lp.NewStringIntHashtable()
				rhHT := rh.NewStringIntHashtable()
				keys := generateKeys(numElements)

				for i, key := range keys {
					lpHT.Put(key, i)
					rhHT.Put(key, i)
				}

				Expect(lpHT.Size()).To(Equal(numElements))
				Expect(rhHT.Size()).To(Equal(numElements))
			})
		})

		Context("Lookup operations", func() {
			It("should find same elements in both implementations", func() {
				lpHT := lp.NewStringIntHashtable()
				rhHT := rh.NewStringIntHashtable()
				keys := generateKeys(numElements)

				for i, key := range keys {
					lpHT.Put(key, i)
					rhHT.Put(key, i)
				}

				for i, key := range keys {
					lpVal, lpFound := lpHT.Lookup(key)
					rhVal, rhFound := rhHT.Lookup(key)

					Expect(lpFound).To(BeTrue())
					Expect(rhFound).To(BeTrue())
					Expect(lpVal).To(Equal(i))
					Expect(rhVal).To(Equal(i))
				}
			})
		})

		Context("Delete operations", func() {
			It("should delete same elements correctly in both implementations", func() {
				lpHT := lp.NewStringIntHashtable()
				rhHT := rh.NewStringIntHashtable()
				keys := generateKeys(numElements)

				for i, key := range keys {
					lpHT.Put(key, i)
					rhHT.Put(key, i)
				}

				// Delete every other key
				for i := 0; i < numElements; i += 2 {
					lpRemoved := lpHT.Remove(keys[i])
					rhRemoved := rhHT.Remove(keys[i])

					Expect(lpRemoved).To(BeTrue())
					Expect(rhRemoved).To(BeTrue())
				}

				// Verify remaining keys
				for i := 1; i < numElements; i += 2 {
					Expect(lpHT.Contains(keys[i])).To(BeTrue())
					Expect(rhHT.Contains(keys[i])).To(BeTrue())
				}

				// Verify deleted keys are gone
				for i := 0; i < numElements; i += 2 {
					Expect(lpHT.Contains(keys[i])).To(BeFalse())
					Expect(rhHT.Contains(keys[i])).To(BeFalse())
				}
			})
		})
	})
})

// Benchmarks (kept as standard Go benchmarks)

// BenchmarkLP_Insert benchmarks Linear Probing insert performance
func BenchmarkLP_Insert(b *testing.B) {
	keys := generateKeys(b.N)

	b.ResetTimer()
	ht := lp.NewStringIntHashtable()
	for i := 0; i < b.N; i++ {
		ht.Put(keys[i], i)
	}
}

// BenchmarkRH_Insert benchmarks Robin Hood insert performance
func BenchmarkRH_Insert(b *testing.B) {
	keys := generateKeys(b.N)

	b.ResetTimer()
	ht := rh.NewStringIntHashtable()
	for i := 0; i < b.N; i++ {
		ht.Put(keys[i], i)
	}
}

// BenchmarkLP_LookupHit benchmarks Linear Probing lookup (100% hit rate)
func BenchmarkLP_LookupHit(b *testing.B) {
	ht := lp.NewStringIntHashtable()
	keys := generateKeys(10000)

	for i, key := range keys {
		ht.Put(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(keys[i%len(keys)])
	}
}

// BenchmarkRH_LookupHit benchmarks Robin Hood lookup (100% hit rate)
func BenchmarkRH_LookupHit(b *testing.B) {
	ht := rh.NewStringIntHashtable()
	keys := generateKeys(10000)

	for i, key := range keys {
		ht.Put(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(keys[i%len(keys)])
	}
}

// BenchmarkLP_LookupMiss benchmarks Linear Probing lookup (0% hit rate)
func BenchmarkLP_LookupMiss(b *testing.B) {
	ht := lp.NewStringIntHashtable()
	keys := generateKeys(10000)

	for i, key := range keys {
		ht.Put(key, i)
	}

	missKeys := generateKeys(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(missKeys[i])
	}
}

// BenchmarkRH_LookupMiss benchmarks Robin Hood lookup (0% hit rate)
func BenchmarkRH_LookupMiss(b *testing.B) {
	ht := rh.NewStringIntHashtable()
	keys := generateKeys(10000)

	for i, key := range keys {
		ht.Put(key, i)
	}

	missKeys := generateKeys(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(missKeys[i])
	}
}

// BenchmarkLP_HighLoadFactor benchmarks LP at ~45% load factor
func BenchmarkLP_HighLoadFactor(b *testing.B) {
	ht := lp.NewWithCapacity[string, int](uint64(b.N*2), lp.StringHasher)
	keys := generateKeys(b.N)

	for i, key := range keys {
		ht.Put(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(keys[i%len(keys)])
	}
}

// BenchmarkRH_HighLoadFactor benchmarks RH at ~45% load factor
func BenchmarkRH_HighLoadFactor(b *testing.B) {
	ht := rh.NewWithCapacity[string, int](uint64(b.N*2), rh.StringHasher)
	keys := generateKeys(b.N)

	for i, key := range keys {
		ht.Put(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Lookup(keys[i%len(keys)])
	}
}

// BenchmarkLP_Delete benchmarks Linear Probing delete
func BenchmarkLP_Delete(b *testing.B) {
	ht := lp.NewStringIntHashtable()
	keys := generateKeys(b.N)

	for i, key := range keys {
		ht.Put(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Remove(keys[i])
	}
}

// BenchmarkRH_Delete benchmarks Robin Hood delete (with backward shift)
func BenchmarkRH_Delete(b *testing.B) {
	ht := rh.NewStringIntHashtable()
	keys := generateKeys(b.N)

	for i, key := range keys {
		ht.Put(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Remove(keys[i])
	}
}

// BenchmarkLP_MixedOperations benchmarks mixed insert/lookup/delete
func BenchmarkLP_MixedOperations(b *testing.B) {
	ht := lp.NewStringIntHashtable()
	keys := generateKeys(10000)

	for i := 0; i < 5000; i++ {
		ht.Put(keys[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op := i % 3
		idx := i % len(keys)
		switch op {
		case 0:
			ht.Put(keys[idx], i)
		case 1:
			ht.Lookup(keys[idx])
		case 2:
			ht.Remove(keys[idx])
		}
	}
}

// BenchmarkRH_MixedOperations benchmarks mixed insert/lookup/delete
func BenchmarkRH_MixedOperations(b *testing.B) {
	ht := rh.NewStringIntHashtable()
	keys := generateKeys(10000)

	for i := 0; i < 5000; i++ {
		ht.Put(keys[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op := i % 3
		idx := i % len(keys)
		switch op {
		case 0:
			ht.Put(keys[idx], i)
		case 1:
			ht.Lookup(keys[idx])
		case 2:
			ht.Remove(keys[idx])
		}
	}
}
