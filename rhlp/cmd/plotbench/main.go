package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	vgdraw "gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"

	"hashfw/hashfw/lp"
	"hashfw/hashfw/rh"
)

// Colors
var (
	colorLP    = color.RGBA{R: 66, G: 133, B: 244, A: 255}  // Blue
	colorRH    = color.RGBA{R: 234, G: 67, B: 53, A: 255}   // Red
	colorGreen = color.RGBA{R: 52, G: 168, B: 83, A: 255}   // Green
	colorAmber = color.RGBA{R: 251, G: 188, B: 4, A: 255}   // Amber
)

// BenchmarkResult holds timing data for a single benchmark point.
type BenchmarkResult struct {
	Name       string
	LP         float64 // nanoseconds per operation
	RH         float64 // nanoseconds per operation
	Size       int     // table size (for X-axis on size-based charts)
	Iterations int     // number of operations measured
}

// PSLStats holds PSL distribution data at a given load factor.
type PSLStats struct {
	LoadFactor float64
	MaxPSL     uint32
	AvgPSL     float64
}

// ────────────────────────────────────────────
// Key generators
// ────────────────────────────────────────────

func generateKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("key_%d_%d", i, rand.Int63())
	}
	return keys
}

// ────────────────────────────────────────────
// Warmup
// ────────────────────────────────────────────

// warmup runs a throwaway insert+lookup cycle on both implementations
// to prime CPU caches, branch predictors, and the Go runtime allocator.
func warmup() {
	keys := generateKeys(5000)
	lpHT := lp.NewStringIntHashtable()
	rhHT := rh.NewStringIntHashtable()
	for j, key := range keys {
		lpHT.Put(key, j)
		rhHT.Put(key, j)
	}
	for _, key := range keys {
		lpHT.Lookup(key)
		rhHT.Lookup(key)
	}
}

// ────────────────────────────────────────────
// Benchmark helpers
// ────────────────────────────────────────────

// Number of iterations per data point. The minimum is taken to eliminate
// GC pauses, allocator noise, and scheduling jitter.
const benchRounds = 5

func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// ────────────────────────────────────────────
// Benchmark functions
// ────────────────────────────────────────────

func benchmarkInsert(sizes []int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(sizes))
	for i, size := range sizes {
		keys := generateKeys(size)

		bestLP := time.Duration(1<<63 - 1)
		bestRH := time.Duration(1<<63 - 1)

		for round := 0; round < benchRounds; round++ {
			lpHT := lp.NewStringIntHashtable()
			start := time.Now()
			for j, key := range keys {
				lpHT.Put(key, j)
			}
			bestLP = minDur(bestLP, time.Since(start))

			rhHT := rh.NewStringIntHashtable()
			start = time.Now()
			for j, key := range keys {
				rhHT.Put(key, j)
			}
			bestRH = minDur(bestRH, time.Since(start))
		}

		results[i] = BenchmarkResult{
			Name: fmt.Sprintf("%d", size),
			LP:   float64(bestLP.Nanoseconds()) / float64(size),
			RH:   float64(bestRH.Nanoseconds()) / float64(size),
			Size: size,
		}
		fmt.Printf("  Insert %d: LP=%.1f ns/op, RH=%.1f ns/op\n", size, results[i].LP, results[i].RH)
	}
	return results
}

func benchmarkLookupHit(sizes []int, lookups int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(sizes))
	for i, size := range sizes {
		keys := generateKeys(size)

		bestLP := time.Duration(1<<63 - 1)
		bestRH := time.Duration(1<<63 - 1)

		for round := 0; round < benchRounds; round++ {
			lpHT := lp.NewStringIntHashtable()
			rhHT := rh.NewStringIntHashtable()
			for j, key := range keys {
				lpHT.Put(key, j)
				rhHT.Put(key, j)
			}

			start := time.Now()
			for j := 0; j < lookups; j++ {
				lpHT.Lookup(keys[j%len(keys)])
			}
			bestLP = minDur(bestLP, time.Since(start))

			start = time.Now()
			for j := 0; j < lookups; j++ {
				rhHT.Lookup(keys[j%len(keys)])
			}
			bestRH = minDur(bestRH, time.Since(start))
		}

		results[i] = BenchmarkResult{
			Name:       fmt.Sprintf("%d", size),
			LP:         float64(bestLP.Nanoseconds()) / float64(lookups),
			RH:         float64(bestRH.Nanoseconds()) / float64(lookups),
			Size:       size,
			Iterations: lookups,
		}
		fmt.Printf("  LookupHit size=%d: LP=%.1f ns/op, RH=%.1f ns/op\n", size, results[i].LP, results[i].RH)
	}
	return results
}

func benchmarkLookupMiss(sizes []int, lookups int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(sizes))
	for i, size := range sizes {
		keys := generateKeys(size)
		missKeys := generateKeys(lookups)

		bestLP := time.Duration(1<<63 - 1)
		bestRH := time.Duration(1<<63 - 1)

		for round := 0; round < benchRounds; round++ {
			lpHT := lp.NewStringIntHashtable()
			rhHT := rh.NewStringIntHashtable()
			for j, key := range keys {
				lpHT.Put(key, j)
				rhHT.Put(key, j)
			}

			start := time.Now()
			for j := 0; j < lookups; j++ {
				lpHT.Lookup(missKeys[j%len(missKeys)])
			}
			bestLP = minDur(bestLP, time.Since(start))

			start = time.Now()
			for j := 0; j < lookups; j++ {
				rhHT.Lookup(missKeys[j%len(missKeys)])
			}
			bestRH = minDur(bestRH, time.Since(start))
		}

		results[i] = BenchmarkResult{
			Name:       fmt.Sprintf("%d", size),
			LP:         float64(bestLP.Nanoseconds()) / float64(lookups),
			RH:         float64(bestRH.Nanoseconds()) / float64(lookups),
			Size:       size,
			Iterations: lookups,
		}
		fmt.Printf("  LookupMiss size=%d: LP=%.1f ns/op, RH=%.1f ns/op\n", size, results[i].LP, results[i].RH)
	}
	return results
}

func benchmarkLoadFactor(loadFactors []float64, capacity int, lookups int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(loadFactors))
	for i, lf := range loadFactors {
		numElements := int(float64(capacity) * lf)
		keys := generateKeys(numElements)

		bestLP := time.Duration(1<<63 - 1)
		bestRH := time.Duration(1<<63 - 1)
		var lastRH *rh.RobinhoodHashtable[string, int]

		for round := 0; round < benchRounds; round++ {
			lpHT := lp.NewWithCapacity[string, int](uint64(capacity), lp.StringHasher)
			rhHT := rh.NewWithCapacity[string, int](uint64(capacity), rh.StringHasher)
			for j, key := range keys {
				lpHT.Put(key, j)
				rhHT.Put(key, j)
			}

			start := time.Now()
			for j := 0; j < lookups; j++ {
				lpHT.Lookup(keys[j%len(keys)])
			}
			bestLP = minDur(bestLP, time.Since(start))

			start = time.Now()
			for j := 0; j < lookups; j++ {
				rhHT.Lookup(keys[j%len(keys)])
			}
			bestRH = minDur(bestRH, time.Since(start))
			lastRH = rhHT
		}

		results[i] = BenchmarkResult{
			Name: fmt.Sprintf("%.0f%%", lf*100),
			LP:   float64(bestLP.Nanoseconds()) / float64(lookups),
			RH:   float64(bestRH.Nanoseconds()) / float64(lookups),
		}
		fmt.Printf("  LoadFactor %.0f%%: LP=%.1f ns/op, RH=%.1f ns/op (MaxPSL=%d, AvgPSL=%.2f)\n",
			lf*100, results[i].LP, results[i].RH, lastRH.MaxPSL(), lastRH.AveragePSL())
	}
	return results
}

func collectPSLDistribution(loadFactors []float64, capacity int) []PSLStats {
	results := make([]PSLStats, len(loadFactors))
	for i, lf := range loadFactors {
		numElements := int(float64(capacity) * lf)
		keys := generateKeys(numElements)
		rhHT := rh.NewWithCapacity[string, int](uint64(capacity), rh.StringHasher)
		for j, key := range keys {
			rhHT.Put(key, j)
		}
		results[i] = PSLStats{
			LoadFactor: lf,
			MaxPSL:     rhHT.MaxPSL(),
			AvgPSL:     rhHT.AveragePSL(),
		}
		fmt.Printf("  PSL at %.0f%%: MaxPSL=%d, AvgPSL=%.2f\n", lf*100, results[i].MaxPSL, results[i].AvgPSL)
	}
	return results
}

// ────────────────────────────────────────────
// NEW scenario benchmarks
// ────────────────────────────────────────────

// benchmarkLookupMissVsLoadFactor measures miss-lookup time at different load factors.
// This is the key benchmark for showing Robin Hood's early-termination advantage.
func benchmarkLookupMissVsLoadFactor(loadFactors []float64, capacity int, lookups int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(loadFactors))
	missKeys := generateKeys(lookups)

	for i, lf := range loadFactors {
		numElements := int(float64(capacity) * lf)
		keys := generateKeys(numElements)

		bestLP := time.Duration(1<<63 - 1)
		bestRH := time.Duration(1<<63 - 1)

		for round := 0; round < benchRounds; round++ {
			lpHT := lp.NewWithCapacity[string, int](uint64(capacity), lp.StringHasher)
			rhHT := rh.NewWithCapacity[string, int](uint64(capacity), rh.StringHasher)
			for j, key := range keys {
				lpHT.Put(key, j)
				rhHT.Put(key, j)
			}

			start := time.Now()
			for j := 0; j < lookups; j++ {
				lpHT.Lookup(missKeys[j%len(missKeys)])
			}
			bestLP = minDur(bestLP, time.Since(start))

			start = time.Now()
			for j := 0; j < lookups; j++ {
				rhHT.Lookup(missKeys[j%len(missKeys)])
			}
			bestRH = minDur(bestRH, time.Since(start))
		}

		results[i] = BenchmarkResult{
			Name: fmt.Sprintf("%.0f%%", lf*100),
			LP:   float64(bestLP.Nanoseconds()) / float64(lookups),
			RH:   float64(bestRH.Nanoseconds()) / float64(lookups),
		}
		fmt.Printf("  MissVsLF %.0f%%: LP=%.1f ns/op, RH=%.1f ns/op\n",
			lf*100, results[i].LP, results[i].RH)
	}
	return results
}

// benchmarkInsertVsLoadFactor measures the time to insert the LAST batch of elements
// that brings the table from (lf-5%) to lf. This shows insert cost at each fill level.
func benchmarkInsertVsLoadFactor(loadFactors []float64, capacity int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(loadFactors))

	for i, lf := range loadFactors {
		totalElements := int(float64(capacity) * lf)
		prefill := totalElements - int(float64(capacity)*0.05)
		if prefill < 0 {
			prefill = 0
		}
		allKeys := generateKeys(totalElements)
		batchKeys := allKeys[prefill:]
		batchSize := len(batchKeys)

		bestLP := time.Duration(1<<63 - 1)
		bestRH := time.Duration(1<<63 - 1)

		for round := 0; round < benchRounds; round++ {
			lpHT := lp.NewWithCapacity[string, int](uint64(capacity), lp.StringHasher)
			rhHT := rh.NewWithCapacity[string, int](uint64(capacity), rh.StringHasher)
			for j := 0; j < prefill; j++ {
				lpHT.Put(allKeys[j], j)
				rhHT.Put(allKeys[j], j)
			}

			start := time.Now()
			for j, key := range batchKeys {
				lpHT.Put(key, prefill+j)
			}
			bestLP = minDur(bestLP, time.Since(start))

			start = time.Now()
			for j, key := range batchKeys {
				rhHT.Put(key, prefill+j)
			}
			bestRH = minDur(bestRH, time.Since(start))
		}

		results[i] = BenchmarkResult{
			Name: fmt.Sprintf("%.0f%%", lf*100),
			LP:   float64(bestLP.Nanoseconds()) / float64(batchSize),
			RH:   float64(bestRH.Nanoseconds()) / float64(batchSize),
		}
		fmt.Printf("  InsertVsLF %.0f%%: LP=%.1f ns/op, RH=%.1f ns/op (batch=%d)\n",
			lf*100, results[i].LP, results[i].RH, batchSize)
	}
	return results
}

// benchmarkMixedHitRate measures lookup at varying hit rates at a fixed high load factor.
// hitRates are 0.0-1.0; 0.0 = all misses, 1.0 = all hits.
func benchmarkMixedHitRate(hitRates []float64, capacity int, lookups int) []BenchmarkResult {
	const loadFactor = 0.45
	numElements := int(float64(capacity) * loadFactor)
	keys := generateKeys(numElements)
	missKeys := generateKeys(lookups)

	results := make([]BenchmarkResult, len(hitRates))
	for i, hr := range hitRates {
		hitCount := int(float64(lookups) * hr)
		seq := make([]string, lookups)
		for j := 0; j < lookups; j++ {
			if j < hitCount {
				seq[j] = keys[j%len(keys)]
			} else {
				seq[j] = missKeys[j%len(missKeys)]
			}
		}
		rand.Shuffle(len(seq), func(a, b int) { seq[a], seq[b] = seq[b], seq[a] })

		bestLP := time.Duration(1<<63 - 1)
		bestRH := time.Duration(1<<63 - 1)

		for round := 0; round < benchRounds; round++ {
			lpHT := lp.NewWithCapacity[string, int](uint64(capacity), lp.StringHasher)
			rhHT := rh.NewWithCapacity[string, int](uint64(capacity), rh.StringHasher)
			for j, key := range keys {
				lpHT.Put(key, j)
				rhHT.Put(key, j)
			}

			start := time.Now()
			for _, k := range seq {
				lpHT.Lookup(k)
			}
			bestLP = minDur(bestLP, time.Since(start))

			start = time.Now()
			for _, k := range seq {
				rhHT.Lookup(k)
			}
			bestRH = minDur(bestRH, time.Since(start))
		}

		results[i] = BenchmarkResult{
			Name: fmt.Sprintf("%.0f%%", hr*100),
			LP:   float64(bestLP.Nanoseconds()) / float64(lookups),
			RH:   float64(bestRH.Nanoseconds()) / float64(lookups),
		}
		fmt.Printf("  HitRate %.0f%%: LP=%.1f ns/op, RH=%.1f ns/op\n",
			hr*100, results[i].LP, results[i].RH)
	}
	return results
}

// ────────────────────────────────────────────
// Plot creation helpers
// ────────────────────────────────────────────

func newLinePlot(results []BenchmarkResult, title, yLabel, xLabel string, useSize bool) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = title
	p.Y.Label.Text = yLabel
	p.X.Label.Text = xLabel

	lpPts := make(plotter.XYs, len(results))
	rhPts := make(plotter.XYs, len(results))
	for i, r := range results {
		x := float64(r.Size)
		if !useSize {
			x = float64(i)
		}
		lpPts[i] = plotter.XY{X: x, Y: r.LP}
		rhPts[i] = plotter.XY{X: x, Y: r.RH}
	}

	lpLine, lpSc, err := plotter.NewLinePoints(lpPts)
	if err != nil {
		return nil, err
	}
	lpLine.Color = colorLP
	lpLine.Width = vg.Points(2)
	lpSc.Color = colorLP
	lpSc.Radius = vg.Points(3)

	rhLine, rhSc, err := plotter.NewLinePoints(rhPts)
	if err != nil {
		return nil, err
	}
	rhLine.Color = colorRH
	rhLine.Width = vg.Points(2)
	rhSc.Color = colorRH
	rhSc.Radius = vg.Points(3)

	p.Add(lpLine, lpSc, rhLine, rhSc)
	p.Legend.Add("Linear Probing", lpLine)
	p.Legend.Add("Robin Hood", rhLine)
	p.Legend.Top = true
	return p, nil
}

func newBarPlot(results []BenchmarkResult, title, yLabel, xLabel string) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = title
	p.Y.Label.Text = yLabel
	p.X.Label.Text = xLabel

	lpData := make(plotter.Values, len(results))
	rhData := make(plotter.Values, len(results))
	labels := make([]string, len(results))
	for i, r := range results {
		lpData[i] = r.LP
		rhData[i] = r.RH
		labels[i] = r.Name
	}

	w := vg.Points(20)
	lpBars, err := plotter.NewBarChart(lpData, w)
	if err != nil {
		return nil, err
	}
	lpBars.LineStyle.Width = vg.Length(0)
	lpBars.Color = colorLP
	lpBars.Offset = -w / 2

	rhBars, err := plotter.NewBarChart(rhData, w)
	if err != nil {
		return nil, err
	}
	rhBars.LineStyle.Width = vg.Length(0)
	rhBars.Color = colorRH
	rhBars.Offset = w / 2

	p.Add(lpBars, rhBars)
	p.Legend.Add("Linear Probing", lpBars)
	p.Legend.Add("Robin Hood", rhBars)
	p.Legend.Top = true
	p.NominalX(labels...)
	return p, nil
}

func newPSLBarPlot(stats []PSLStats) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = "PSL Statistics"
	p.Y.Label.Text = "PSL"
	p.X.Label.Text = "Load Factor"

	maxData := make(plotter.Values, len(stats))
	avgData := make(plotter.Values, len(stats))
	labels := make([]string, len(stats))
	for i, s := range stats {
		maxData[i] = float64(s.MaxPSL)
		avgData[i] = s.AvgPSL
		labels[i] = fmt.Sprintf("%.0f%%", s.LoadFactor*100)
	}

	w := vg.Points(18)
	maxBars, err := plotter.NewBarChart(maxData, w)
	if err != nil {
		return nil, err
	}
	maxBars.LineStyle.Width = vg.Length(0)
	maxBars.Color = colorRH
	maxBars.Offset = -w / 2

	avgBars, err := plotter.NewBarChart(avgData, w)
	if err != nil {
		return nil, err
	}
	avgBars.LineStyle.Width = vg.Length(0)
	avgBars.Color = colorGreen
	avgBars.Offset = w / 2

	p.Add(maxBars, avgBars)
	p.Legend.Add("Max PSL", maxBars)
	p.Legend.Add("Avg PSL", avgBars)
	p.Legend.Top = true
	p.NominalX(labels...)
	return p, nil
}

func newPSLGrowthPlot(stats []PSLStats) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = "PSL Growth"
	p.Y.Label.Text = "Probe Sequence Length"
	p.X.Label.Text = "Load Factor (%)"

	maxPts := make(plotter.XYs, len(stats))
	avgPts := make(plotter.XYs, len(stats))
	for i, s := range stats {
		maxPts[i] = plotter.XY{X: s.LoadFactor * 100, Y: float64(s.MaxPSL)}
		avgPts[i] = plotter.XY{X: s.LoadFactor * 100, Y: s.AvgPSL}
	}

	maxL, maxS, err := plotter.NewLinePoints(maxPts)
	if err != nil {
		return nil, err
	}
	maxL.Color = colorRH
	maxL.Width = vg.Points(2)
	maxS.Color = colorRH
	maxS.Radius = vg.Points(3)

	avgL, avgS, err := plotter.NewLinePoints(avgPts)
	if err != nil {
		return nil, err
	}
	avgL.Color = colorGreen
	avgL.Width = vg.Points(2)
	avgS.Color = colorGreen
	avgS.Radius = vg.Points(3)

	p.Add(maxL, maxS, avgL, avgS)
	p.Legend.Add("Max PSL", maxL)
	p.Legend.Add("Avg PSL", avgL)
	p.Legend.Top = true
	p.Legend.Left = true
	return p, nil
}

// ────────────────────────────────────────────
// Grid composer
// ────────────────────────────────────────────

func saveGrid(plots []*plot.Plot, cols int, cellW, cellH vg.Length, filename string) error {
	rows := (len(plots) + cols - 1) / cols
	totalW := vg.Length(cols) * cellW
	totalH := vg.Length(rows) * cellH

	imgW := int(totalW.Dots(vgimg.DefaultDPI))
	imgH := int(totalH.Dots(vgimg.DefaultDPI))
	combined := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(combined, combined.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	cw := int(cellW.Dots(vgimg.DefaultDPI))
	ch := int(cellH.Dots(vgimg.DefaultDPI))

	for i, p := range plots {
		r := i / cols
		c := i % cols
		canvas := vgimg.New(cellW, cellH)
		p.Draw(vgdraw.New(canvas))
		x, y := c*cw, r*ch
		draw.Draw(combined, image.Rect(x, y, x+cw, y+ch), canvas.Image(), image.Point{}, draw.Over)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, combined)
}

// ────────────────────────────────────────────
// Main
// ────────────────────────────────────────────

func main() {
	rand.Seed(time.Now().UnixNano())

	outputDir := "plots"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Hash Table Benchmark Plots ===")
	fmt.Println()

	fmt.Println("Warming up caches and runtime...")
	warmup()
	fmt.Println()

	// ── Shared parameters ──
	sizes := []int{100, 500, 1000, 5000, 10000, 50000}
	lookups := 100000
	capacity := 100000
	loadFactors := []float64{0.10, 0.20, 0.30, 0.40, 0.45}
	pslLoadFactors := []float64{0.05, 0.10, 0.15, 0.20, 0.25, 0.30, 0.35, 0.40, 0.45}

	// ── Original benchmarks ──
	fmt.Println("[1/8] Insert benchmarks...")
	insertResults := benchmarkInsert(sizes)
	fmt.Println()

	fmt.Println("[2/8] Lookup (hit) benchmarks...")
	lookupHitResults := benchmarkLookupHit(sizes, lookups)
	fmt.Println()

	fmt.Println("[3/8] Lookup (miss) benchmarks...")
	lookupMissResults := benchmarkLookupMiss(sizes, lookups)
	fmt.Println()

	fmt.Println("[4/8] Load factor (hit lookup) benchmarks...")
	loadFactorResults := benchmarkLoadFactor(loadFactors, capacity, lookups)
	fmt.Println()

	fmt.Println("[5/8] PSL distribution...")
	pslStats := collectPSLDistribution(pslLoadFactors, capacity)
	fmt.Println()

	// ── NEW scenario benchmarks ──
	fmt.Println("[6/8] Lookup MISS vs load factor...")
	missVsLFResults := benchmarkLookupMissVsLoadFactor(loadFactors, capacity, lookups)
	fmt.Println()

	fmt.Println("[7/8] Insert vs load factor...")
	insertVsLFResults := benchmarkInsertVsLoadFactor(loadFactors, capacity)
	fmt.Println()

	fmt.Println("[8/8] Mixed hit-rate at 45% load...")
	hitRates := []float64{0.0, 0.25, 0.50, 0.75, 1.0}
	hitRateResults := benchmarkMixedHitRate(hitRates, capacity, lookups)
	fmt.Println()

	// ──────────────────────────────────────────
	// Plot 1: benchmark_grid.png (2x3 — fixed)
	// ──────────────────────────────────────────
	fmt.Println("Creating benchmark_grid.png ...")

	insertPlot, _ := newLinePlot(insertResults, "Insert Performance", "Time (ns/op)", "Number of Elements", true)
	lookupHitPlot, _ := newLinePlot(lookupHitResults, "Lookup (100% Hit)", "Time (ns/op)", "Table Size", true)
	lookupMissPlot, _ := newLinePlot(lookupMissResults, "Lookup (0% Hit — Miss)", "Time (ns/op)", "Table Size", true)
	loadFactorPlot, _ := newBarPlot(loadFactorResults, "Lookup Hit vs Load Factor", "Time (ns/op)", "Load Factor")
	pslStatsPlot, _ := newPSLBarPlot(pslStats)
	pslGrowthPlot, _ := newPSLGrowthPlot(pslStats)

	if err := saveGrid(
		[]*plot.Plot{insertPlot, lookupHitPlot, lookupMissPlot, loadFactorPlot, pslStatsPlot, pslGrowthPlot},
		3, 6*vg.Inch, 4*vg.Inch,
		filepath.Join(outputDir, "benchmark_grid.png"),
	); err != nil {
		log.Printf("benchmark_grid: %v", err)
	}

	// ──────────────────────────────────────────
	// Plot 2: summary_comparison.png
	// ──────────────────────────────────────────
	fmt.Println("Creating summary_comparison.png ...")

	summaryResults := []BenchmarkResult{
		{Name: "Insert\n(10K)", LP: insertResults[4].LP, RH: insertResults[4].RH},
		{Name: "Lookup\nHit", LP: lookupHitResults[4].LP, RH: lookupHitResults[4].RH},
		{Name: "Lookup\nMiss", LP: lookupMissResults[4].LP, RH: lookupMissResults[4].RH},
		{Name: "High\nLoad", LP: loadFactorResults[4].LP, RH: loadFactorResults[4].RH},
	}
	summaryPlot, _ := newBarPlot(summaryResults, "Performance Summary: Linear Probing vs Robin Hood", "Time (ns/op)", "")
	if err := summaryPlot.Save(8*vg.Inch, 5*vg.Inch, filepath.Join(outputDir, "summary_comparison.png")); err != nil {
		log.Printf("summary_comparison: %v", err)
	}

	// ──────────────────────────────────────────
	// Plot 3: robin_hood_advantage.png (1x3)
	//   Scenario 1: lookup misses common / high load
	// ──────────────────────────────────────────
	fmt.Println("Creating robin_hood_advantage.png ...")

	// 3a. Lookup miss time vs load factor — the key chart
	missVsLFPlot, _ := newBarPlot(missVsLFResults,
		"Lookup Miss vs Load Factor",
		"Time (ns/op)", "Load Factor")

	// 3b. Lookup time by hit rate at 45% load — shows crossover
	hitRatePlot, _ := newBarPlot(hitRateResults,
		"Lookup by Hit Rate (at 45% Load)",
		"Time (ns/op)", "Hit Rate")

	// 3c. PSL growth — explains WHY RH wins (low avg PSL → early termination)
	pslGrowthPlot2, _ := newPSLGrowthPlot(pslStats)
	pslGrowthPlot2.Title.Text = "Why: PSL Stays Low (Early Termination)"

	if err := saveGrid(
		[]*plot.Plot{missVsLFPlot, hitRatePlot, pslGrowthPlot2},
		3, 6*vg.Inch, 4*vg.Inch,
		filepath.Join(outputDir, "robin_hood_advantage.png"),
	); err != nil {
		log.Printf("robin_hood_advantage: %v", err)
	}

	// ──────────────────────────────────────────
	// Plot 4: linear_probing_advantage.png (1x3)
	//   Scenario 2: insert-heavy, load factor < 50%
	// ──────────────────────────────────────────
	fmt.Println("Creating linear_probing_advantage.png ...")

	// 4a. Insert time vs table size — LP consistently faster
	insertSizePlot, _ := newLinePlot(insertResults,
		"Insert: LP Faster at Every Size",
		"Time (ns/op)", "Number of Elements", true)

	// 4b. Insert cost vs load factor — LP advantage holds across fill levels
	insertVsLFPlot, _ := newBarPlot(insertVsLFResults,
		"Insert Cost vs Load Factor",
		"Time (ns/op)", "Load Factor")

	// 4c. Lookup hit vs load factor — LP competitive, no downside
	loadFactorPlot2, _ := newBarPlot(loadFactorResults,
		"Lookup Hit: Comparable Below 50%",
		"Time (ns/op)", "Load Factor")

	if err := saveGrid(
		[]*plot.Plot{insertSizePlot, insertVsLFPlot, loadFactorPlot2},
		3, 6*vg.Inch, 4*vg.Inch,
		filepath.Join(outputDir, "linear_probing_advantage.png"),
	); err != nil {
		log.Printf("linear_probing_advantage: %v", err)
	}

	fmt.Println()
	fmt.Printf("Plots saved to: %s/\n", outputDir)
	fmt.Println("  - benchmark_grid.png           (6 plots: 2 rows x 3 columns)")
	fmt.Println("  - summary_comparison.png        (overall bar chart)")
	fmt.Println("  - robin_hood_advantage.png      (scenario: misses & high load)")
	fmt.Println("  - linear_probing_advantage.png  (scenario: inserts & low load)")
}
