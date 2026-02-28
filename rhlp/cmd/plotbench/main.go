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

// BenchmarkResult holds timing data for a single benchmark
type BenchmarkResult struct {
	Name       string
	LP         float64 // nanoseconds per operation
	RH         float64 // nanoseconds per operation
	Iterations int
}

// generateKeys creates random string keys
func generateKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("key_%d_%d", i, rand.Int63())
	}
	return keys
}

// benchmarkInsert measures insert performance at various sizes
func benchmarkInsert(sizes []int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(sizes))

	for i, size := range sizes {
		keys := generateKeys(size)

		// Benchmark LP
		lpHT := lp.NewStringIntHashtable()
		start := time.Now()
		for j, key := range keys {
			lpHT.Put(key, j)
		}
		lpDuration := time.Since(start)

		// Benchmark RH
		rhHT := rh.NewStringIntHashtable()
		start = time.Now()
		for j, key := range keys {
			rhHT.Put(key, j)
		}
		rhDuration := time.Since(start)

		results[i] = BenchmarkResult{
			Name:       fmt.Sprintf("%d", size),
			LP:         float64(lpDuration.Nanoseconds()) / float64(size),
			RH:         float64(rhDuration.Nanoseconds()) / float64(size),
			Iterations: size,
		}
		fmt.Printf("Insert %d elements: LP=%.2f ns/op, RH=%.2f ns/op\n", size, results[i].LP, results[i].RH)
	}
	return results
}

// benchmarkLookupHit measures lookup performance with 100% hit rate
func benchmarkLookupHit(sizes []int, lookups int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(sizes))

	for i, size := range sizes {
		keys := generateKeys(size)

		// Setup LP
		lpHT := lp.NewStringIntHashtable()
		for j, key := range keys {
			lpHT.Put(key, j)
		}

		// Setup RH
		rhHT := rh.NewStringIntHashtable()
		for j, key := range keys {
			rhHT.Put(key, j)
		}

		// Benchmark LP lookups
		start := time.Now()
		for j := 0; j < lookups; j++ {
			lpHT.Lookup(keys[j%len(keys)])
		}
		lpDuration := time.Since(start)

		// Benchmark RH lookups
		start = time.Now()
		for j := 0; j < lookups; j++ {
			rhHT.Lookup(keys[j%len(keys)])
		}
		rhDuration := time.Since(start)

		results[i] = BenchmarkResult{
			Name:       fmt.Sprintf("%d", size),
			LP:         float64(lpDuration.Nanoseconds()) / float64(lookups),
			RH:         float64(rhDuration.Nanoseconds()) / float64(lookups),
			Iterations: lookups,
		}
		fmt.Printf("Lookup (hit) %d elements: LP=%.2f ns/op, RH=%.2f ns/op\n", size, results[i].LP, results[i].RH)
	}
	return results
}

// benchmarkLookupMiss measures lookup performance with 0% hit rate
func benchmarkLookupMiss(sizes []int, lookups int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(sizes))

	for i, size := range sizes {
		keys := generateKeys(size)
		missKeys := generateKeys(lookups)

		// Setup LP
		lpHT := lp.NewStringIntHashtable()
		for j, key := range keys {
			lpHT.Put(key, j)
		}

		// Setup RH
		rhHT := rh.NewStringIntHashtable()
		for j, key := range keys {
			rhHT.Put(key, j)
		}

		// Benchmark LP lookups
		start := time.Now()
		for j := 0; j < lookups; j++ {
			lpHT.Lookup(missKeys[j%len(missKeys)])
		}
		lpDuration := time.Since(start)

		// Benchmark RH lookups
		start = time.Now()
		for j := 0; j < lookups; j++ {
			rhHT.Lookup(missKeys[j%len(missKeys)])
		}
		rhDuration := time.Since(start)

		results[i] = BenchmarkResult{
			Name:       fmt.Sprintf("%d", size),
			LP:         float64(lpDuration.Nanoseconds()) / float64(lookups),
			RH:         float64(rhDuration.Nanoseconds()) / float64(lookups),
			Iterations: lookups,
		}
		fmt.Printf("Lookup (miss) %d elements: LP=%.2f ns/op, RH=%.2f ns/op\n", size, results[i].LP, results[i].RH)
	}
	return results
}

// benchmarkLoadFactor measures performance at different load factors
func benchmarkLoadFactor(loadFactors []float64, capacity int, lookups int) []BenchmarkResult {
	results := make([]BenchmarkResult, len(loadFactors))

	for i, lf := range loadFactors {
		numElements := int(float64(capacity) * lf)
		keys := generateKeys(numElements)

		// Setup LP with fixed capacity
		lpHT := lp.NewWithCapacity[string, int](uint64(capacity), lp.StringHasher)
		for j, key := range keys {
			lpHT.Put(key, j)
		}

		// Setup RH with fixed capacity
		rhHT := rh.NewWithCapacity[string, int](uint64(capacity), rh.StringHasher)
		for j, key := range keys {
			rhHT.Put(key, j)
		}

		// Benchmark LP lookups
		start := time.Now()
		for j := 0; j < lookups; j++ {
			lpHT.Lookup(keys[j%len(keys)])
		}
		lpDuration := time.Since(start)

		// Benchmark RH lookups
		start = time.Now()
		for j := 0; j < lookups; j++ {
			rhHT.Lookup(keys[j%len(keys)])
		}
		rhDuration := time.Since(start)

		results[i] = BenchmarkResult{
			Name:       fmt.Sprintf("%.0f%%", lf*100),
			LP:         float64(lpDuration.Nanoseconds()) / float64(lookups),
			RH:         float64(rhDuration.Nanoseconds()) / float64(lookups),
			Iterations: lookups,
		}
		fmt.Printf("Load Factor %.0f%%: LP=%.2f ns/op, RH=%.2f ns/op (MaxPSL=%d, AvgPSL=%.2f)\n",
			lf*100, results[i].LP, results[i].RH, rhHT.MaxPSL(), rhHT.AveragePSL())
	}
	return results
}

// createBarChart creates a grouped bar chart comparing LP vs RH
func createBarChart(results []BenchmarkResult, title, yLabel, filename string) error {
	p := plot.New()
	p.Title.Text = title
	p.Y.Label.Text = yLabel
	p.X.Label.Text = "Elements"

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
		return err
	}
	lpBars.LineStyle.Width = vg.Length(0)
	lpBars.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255} // Blue
	lpBars.Offset = -w / 2

	rhBars, err := plotter.NewBarChart(rhData, w)
	if err != nil {
		return err
	}
	rhBars.LineStyle.Width = vg.Length(0)
	rhBars.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255} // Red
	rhBars.Offset = w / 2

	p.Add(lpBars, rhBars)
	p.Legend.Add("Linear Probing", lpBars)
	p.Legend.Add("Robin Hood", rhBars)
	p.Legend.Top = true
	p.NominalX(labels...)

	return p.Save(8*vg.Inch, 5*vg.Inch, filename)
}

// createLineChart creates a line chart comparing LP vs RH over sizes
func createLineChart(results []BenchmarkResult, title, yLabel, filename string) error {
	p := plot.New()
	p.Title.Text = title
	p.Y.Label.Text = yLabel
	p.X.Label.Text = "Number of Elements"

	lpPoints := make(plotter.XYs, len(results))
	rhPoints := make(plotter.XYs, len(results))

	for i, r := range results {
		lpPoints[i].X = float64(r.Iterations)
		lpPoints[i].Y = r.LP
		rhPoints[i].X = float64(r.Iterations)
		rhPoints[i].Y = r.RH
	}

	lpLine, lpScatter, err := plotter.NewLinePoints(lpPoints)
	if err != nil {
		return err
	}
	lpLine.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}
	lpScatter.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}

	rhLine, rhScatter, err := plotter.NewLinePoints(rhPoints)
	if err != nil {
		return err
	}
	rhLine.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	rhScatter.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}

	p.Add(lpLine, lpScatter, rhLine, rhScatter)
	p.Legend.Add("Linear Probing", lpLine)
	p.Legend.Add("Robin Hood", rhLine)
	p.Legend.Top = true

	return p.Save(8*vg.Inch, 5*vg.Inch, filename)
}

// createLoadFactorChart creates a chart showing performance vs load factor
func createLoadFactorChart(results []BenchmarkResult, title, filename string) error {
	p := plot.New()
	p.Title.Text = title
	p.Y.Label.Text = "Time (ns/op)"
	p.X.Label.Text = "Load Factor"

	lpData := make(plotter.Values, len(results))
	rhData := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		lpData[i] = r.LP
		rhData[i] = r.RH
		labels[i] = r.Name
	}

	w := vg.Points(25)

	lpBars, err := plotter.NewBarChart(lpData, w)
	if err != nil {
		return err
	}
	lpBars.LineStyle.Width = vg.Length(0)
	lpBars.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}
	lpBars.Offset = -w / 2

	rhBars, err := plotter.NewBarChart(rhData, w)
	if err != nil {
		return err
	}
	rhBars.LineStyle.Width = vg.Length(0)
	rhBars.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	rhBars.Offset = w / 2

	p.Add(lpBars, rhBars)
	p.Legend.Add("Linear Probing", lpBars)
	p.Legend.Add("Robin Hood", rhBars)
	p.Legend.Top = true
	p.NominalX(labels...)

	return p.Save(8*vg.Inch, 5*vg.Inch, filename)
}

// PSLStats holds PSL distribution data
type PSLStats struct {
	LoadFactor float64
	MaxPSL     uint32
	AvgPSL     float64
	Distribution map[uint32]int
}

// collectPSLDistribution collects PSL distribution at various load factors
func collectPSLDistribution(loadFactors []float64, capacity int) []PSLStats {
	results := make([]PSLStats, len(loadFactors))

	for i, lf := range loadFactors {
		numElements := int(float64(capacity) * lf)
		keys := generateKeys(numElements)

		rhHT := rh.NewWithCapacity[string, int](uint64(capacity), rh.StringHasher)
		for j, key := range keys {
			rhHT.Put(key, j)
		}

		// We need to access the internal cells to get PSL distribution
		// For now, we'll use the stats we can get
		results[i] = PSLStats{
			LoadFactor: lf,
			MaxPSL:     rhHT.MaxPSL(),
			AvgPSL:     rhHT.AveragePSL(),
		}
		fmt.Printf("PSL at %.0f%% load: MaxPSL=%d, AvgPSL=%.2f\n", lf*100, results[i].MaxPSL, results[i].AvgPSL)
	}
	return results
}

// createPSLChart creates a chart showing PSL statistics at different load factors
func createPSLChart(stats []PSLStats, filename string) error {
	p := plot.New()
	p.Title.Text = "Robin Hood PSL Statistics vs Load Factor"
	p.Y.Label.Text = "PSL"
	p.X.Label.Text = "Load Factor"

	maxPSLData := make(plotter.Values, len(stats))
	avgPSLData := make(plotter.Values, len(stats))
	labels := make([]string, len(stats))

	for i, s := range stats {
		maxPSLData[i] = float64(s.MaxPSL)
		avgPSLData[i] = s.AvgPSL
		labels[i] = fmt.Sprintf("%.0f%%", s.LoadFactor*100)
	}

	w := vg.Points(25)

	maxBars, err := plotter.NewBarChart(maxPSLData, w)
	if err != nil {
		return err
	}
	maxBars.LineStyle.Width = vg.Length(0)
	maxBars.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255} // Red
	maxBars.Offset = -w / 2

	avgBars, err := plotter.NewBarChart(avgPSLData, w)
	if err != nil {
		return err
	}
	avgBars.LineStyle.Width = vg.Length(0)
	avgBars.Color = color.RGBA{R: 52, G: 168, B: 83, A: 255} // Green
	avgBars.Offset = w / 2

	p.Add(maxBars, avgBars)
	p.Legend.Add("Max PSL", maxBars)
	p.Legend.Add("Avg PSL", avgBars)
	p.Legend.Top = true
	p.NominalX(labels...)

	return p.Save(8*vg.Inch, 5*vg.Inch, filename)
}

// createPSLGrowthChart shows how PSL grows with load factor
func createPSLGrowthChart(stats []PSLStats, filename string) error {
	p := plot.New()
	p.Title.Text = "Robin Hood: PSL Growth vs Load Factor"
	p.Y.Label.Text = "Probe Sequence Length"
	p.X.Label.Text = "Load Factor (%)"

	maxPoints := make(plotter.XYs, len(stats))
	avgPoints := make(plotter.XYs, len(stats))

	for i, s := range stats {
		maxPoints[i].X = s.LoadFactor * 100
		maxPoints[i].Y = float64(s.MaxPSL)
		avgPoints[i].X = s.LoadFactor * 100
		avgPoints[i].Y = s.AvgPSL
	}

	maxLine, maxScatter, err := plotter.NewLinePoints(maxPoints)
	if err != nil {
		return err
	}
	maxLine.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	maxLine.Width = vg.Points(2)
	maxScatter.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	maxScatter.Radius = vg.Points(4)

	avgLine, avgScatter, err := plotter.NewLinePoints(avgPoints)
	if err != nil {
		return err
	}
	avgLine.Color = color.RGBA{R: 52, G: 168, B: 83, A: 255}
	avgLine.Width = vg.Points(2)
	avgScatter.Color = color.RGBA{R: 52, G: 168, B: 83, A: 255}
	avgScatter.Radius = vg.Points(4)

	p.Add(maxLine, maxScatter, avgLine, avgScatter)
	p.Legend.Add("Max PSL", maxLine)
	p.Legend.Add("Avg PSL", avgLine)
	p.Legend.Top = true
	p.Legend.Left = true

	return p.Save(8*vg.Inch, 5*vg.Inch, filename)
}

// PlotData holds a plot and its metadata for grid composition
type PlotData struct {
	Plot  *plot.Plot
	Title string
}

// createInsertPlot creates a plot for insert performance (returns plot instead of saving)
func createInsertPlot(results []BenchmarkResult) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = "Insert Performance"
	p.Y.Label.Text = "Time (ns/op)"
	p.X.Label.Text = "Number of Elements"

	lpPoints := make(plotter.XYs, len(results))
	rhPoints := make(plotter.XYs, len(results))

	for i, r := range results {
		lpPoints[i].X = float64(r.Iterations)
		lpPoints[i].Y = r.LP
		rhPoints[i].X = float64(r.Iterations)
		rhPoints[i].Y = r.RH
	}

	lpLine, lpScatter, err := plotter.NewLinePoints(lpPoints)
	if err != nil {
		return nil, err
	}
	lpLine.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}
	lpScatter.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}

	rhLine, rhScatter, err := plotter.NewLinePoints(rhPoints)
	if err != nil {
		return nil, err
	}
	rhLine.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	rhScatter.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}

	p.Add(lpLine, lpScatter, rhLine, rhScatter)
	p.Legend.Add("Linear Probing", lpLine)
	p.Legend.Add("Robin Hood", rhLine)
	p.Legend.Top = true

	return p, nil
}

// createLookupHitPlot creates a plot for lookup hit performance
func createLookupHitPlot(results []BenchmarkResult) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = "Lookup (100% Hit)"
	p.Y.Label.Text = "Time (ns/op)"
	p.X.Label.Text = "Number of Elements"

	lpPoints := make(plotter.XYs, len(results))
	rhPoints := make(plotter.XYs, len(results))

	for i, r := range results {
		lpPoints[i].X = float64(r.Iterations)
		lpPoints[i].Y = r.LP
		rhPoints[i].X = float64(r.Iterations)
		rhPoints[i].Y = r.RH
	}

	lpLine, lpScatter, err := plotter.NewLinePoints(lpPoints)
	if err != nil {
		return nil, err
	}
	lpLine.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}
	lpScatter.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}

	rhLine, rhScatter, err := plotter.NewLinePoints(rhPoints)
	if err != nil {
		return nil, err
	}
	rhLine.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	rhScatter.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}

	p.Add(lpLine, lpScatter, rhLine, rhScatter)
	p.Legend.Add("Linear Probing", lpLine)
	p.Legend.Add("Robin Hood", rhLine)
	p.Legend.Top = true

	return p, nil
}

// createLookupMissPlot creates a plot for lookup miss performance
func createLookupMissPlot(results []BenchmarkResult) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = "Lookup (0% Hit)"
	p.Y.Label.Text = "Time (ns/op)"
	p.X.Label.Text = "Number of Elements"

	lpPoints := make(plotter.XYs, len(results))
	rhPoints := make(plotter.XYs, len(results))

	for i, r := range results {
		lpPoints[i].X = float64(r.Iterations)
		lpPoints[i].Y = r.LP
		rhPoints[i].X = float64(r.Iterations)
		rhPoints[i].Y = r.RH
	}

	lpLine, lpScatter, err := plotter.NewLinePoints(lpPoints)
	if err != nil {
		return nil, err
	}
	lpLine.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}
	lpScatter.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}

	rhLine, rhScatter, err := plotter.NewLinePoints(rhPoints)
	if err != nil {
		return nil, err
	}
	rhLine.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	rhScatter.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}

	p.Add(lpLine, lpScatter, rhLine, rhScatter)
	p.Legend.Add("Linear Probing", lpLine)
	p.Legend.Add("Robin Hood", rhLine)
	p.Legend.Top = true

	return p, nil
}

// createLoadFactorPlot creates a plot for load factor comparison
func createLoadFactorPlot(results []BenchmarkResult) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = "Load Factor Impact"
	p.Y.Label.Text = "Time (ns/op)"
	p.X.Label.Text = "Load Factor"

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
	lpBars.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255}
	lpBars.Offset = -w / 2

	rhBars, err := plotter.NewBarChart(rhData, w)
	if err != nil {
		return nil, err
	}
	rhBars.LineStyle.Width = vg.Length(0)
	rhBars.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	rhBars.Offset = w / 2

	p.Add(lpBars, rhBars)
	p.Legend.Add("Linear Probing", lpBars)
	p.Legend.Add("Robin Hood", rhBars)
	p.Legend.Top = true
	p.NominalX(labels...)

	return p, nil
}

// createPSLStatisticsPlot creates a plot for PSL statistics
func createPSLStatisticsPlot(stats []PSLStats) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = "PSL Statistics"
	p.Y.Label.Text = "PSL"
	p.X.Label.Text = "Load Factor"

	maxPSLData := make(plotter.Values, len(stats))
	avgPSLData := make(plotter.Values, len(stats))
	labels := make([]string, len(stats))

	for i, s := range stats {
		maxPSLData[i] = float64(s.MaxPSL)
		avgPSLData[i] = s.AvgPSL
		labels[i] = fmt.Sprintf("%.0f%%", s.LoadFactor*100)
	}

	w := vg.Points(18)

	maxBars, err := plotter.NewBarChart(maxPSLData, w)
	if err != nil {
		return nil, err
	}
	maxBars.LineStyle.Width = vg.Length(0)
	maxBars.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	maxBars.Offset = -w / 2

	avgBars, err := plotter.NewBarChart(avgPSLData, w)
	if err != nil {
		return nil, err
	}
	avgBars.LineStyle.Width = vg.Length(0)
	avgBars.Color = color.RGBA{R: 52, G: 168, B: 83, A: 255}
	avgBars.Offset = w / 2

	p.Add(maxBars, avgBars)
	p.Legend.Add("Max PSL", maxBars)
	p.Legend.Add("Avg PSL", avgBars)
	p.Legend.Top = true
	p.NominalX(labels...)

	return p, nil
}

// createPSLGrowthPlot creates a plot for PSL growth
func createPSLGrowthPlot(stats []PSLStats) (*plot.Plot, error) {
	p := plot.New()
	p.Title.Text = "PSL Growth"
	p.Y.Label.Text = "Probe Sequence Length"
	p.X.Label.Text = "Load Factor (%)"

	maxPoints := make(plotter.XYs, len(stats))
	avgPoints := make(plotter.XYs, len(stats))

	for i, s := range stats {
		maxPoints[i].X = s.LoadFactor * 100
		maxPoints[i].Y = float64(s.MaxPSL)
		avgPoints[i].X = s.LoadFactor * 100
		avgPoints[i].Y = s.AvgPSL
	}

	maxLine, maxScatter, err := plotter.NewLinePoints(maxPoints)
	if err != nil {
		return nil, err
	}
	maxLine.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	maxLine.Width = vg.Points(2)
	maxScatter.Color = color.RGBA{R: 234, G: 67, B: 53, A: 255}
	maxScatter.Radius = vg.Points(3)

	avgLine, avgScatter, err := plotter.NewLinePoints(avgPoints)
	if err != nil {
		return nil, err
	}
	avgLine.Color = color.RGBA{R: 52, G: 168, B: 83, A: 255}
	avgLine.Width = vg.Points(2)
	avgScatter.Color = color.RGBA{R: 52, G: 168, B: 83, A: 255}
	avgScatter.Radius = vg.Points(3)

	p.Add(maxLine, maxScatter, avgLine, avgScatter)
	p.Legend.Add("Max PSL", maxLine)
	p.Legend.Add("Avg PSL", avgLine)
	p.Legend.Top = true
	p.Legend.Left = true

	return p, nil
}

// createCombinedGrid creates a 2x3 grid of plots and saves to a single image
func createCombinedGrid(plots []*plot.Plot, filename string) error {
	if len(plots) != 6 {
		return fmt.Errorf("expected 6 plots, got %d", len(plots))
	}

	// Each cell dimensions
	cellWidth := 6 * vg.Inch
	cellHeight := 4 * vg.Inch

	// Total grid dimensions: 3 columns x 2 rows
	totalWidth := 3 * cellWidth
	totalHeight := 2 * cellHeight

	// Create the combined image canvas
	imgWidth := int(totalWidth.Dots(vgimg.DefaultDPI))
	imgHeight := int(totalHeight.Dots(vgimg.DefaultDPI))
	combined := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	// Fill with white background
	draw.Draw(combined, combined.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Render each plot to its position in the grid
	for i, p := range plots {
		row := i / 3
		col := i % 3

		// Create a canvas for this plot
		c := vgimg.New(cellWidth, cellHeight)

		// Draw the plot onto the canvas
		p.Draw(vgdraw.New(c))

		// Calculate position in combined image
		x := col * int(cellWidth.Dots(vgimg.DefaultDPI))
		y := row * int(cellHeight.Dots(vgimg.DefaultDPI))

		// Copy to combined image
		draw.Draw(combined, image.Rect(x, y, x+int(cellWidth.Dots(vgimg.DefaultDPI)), y+int(cellHeight.Dots(vgimg.DefaultDPI))),
			c.Image(), image.Point{}, draw.Over)
	}

	// Save the combined image
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, combined)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Create output directory
	outputDir := "plots"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Hash Table Benchmark Plots ===")
	fmt.Println()

	// 1. Insert performance at various sizes
	fmt.Println("Running insert benchmarks...")
	insertSizes := []int{100, 500, 1000, 5000, 10000, 50000}
	insertResults := benchmarkInsert(insertSizes)
	fmt.Println()

	// 2. Lookup hit performance
	fmt.Println("Running lookup (hit) benchmarks...")
	lookupSizes := []int{100, 500, 1000, 5000, 10000, 50000}
	lookupHitResults := benchmarkLookupHit(lookupSizes, 100000)
	fmt.Println()

	// 3. Lookup miss performance
	fmt.Println("Running lookup (miss) benchmarks...")
	lookupMissResults := benchmarkLookupMiss(lookupSizes, 100000)
	fmt.Println()

	// 4. Load factor comparison
	fmt.Println("Running load factor benchmarks...")
	loadFactors := []float64{0.10, 0.20, 0.30, 0.40, 0.45}
	loadFactorResults := benchmarkLoadFactor(loadFactors, 100000, 100000)
	fmt.Println()

	// 5. PSL distribution
	fmt.Println("Collecting PSL statistics...")
	pslLoadFactors := []float64{0.05, 0.10, 0.15, 0.20, 0.25, 0.30, 0.35, 0.40, 0.45}
	pslStats := collectPSLDistribution(pslLoadFactors, 100000)
	fmt.Println()

	// Create individual plots for the combined grid
	fmt.Println("Creating combined grid plot (2 rows x 3 columns)...")

	insertPlot, err := createInsertPlot(insertResults)
	if err != nil {
		log.Printf("Error creating insert plot: %v", err)
	}

	lookupHitPlot, err := createLookupHitPlot(lookupHitResults)
	if err != nil {
		log.Printf("Error creating lookup hit plot: %v", err)
	}

	lookupMissPlot, err := createLookupMissPlot(lookupMissResults)
	if err != nil {
		log.Printf("Error creating lookup miss plot: %v", err)
	}

	loadFactorPlot, err := createLoadFactorPlot(loadFactorResults)
	if err != nil {
		log.Printf("Error creating load factor plot: %v", err)
	}

	pslStatsPlot, err := createPSLStatisticsPlot(pslStats)
	if err != nil {
		log.Printf("Error creating PSL statistics plot: %v", err)
	}

	pslGrowthPlot, err := createPSLGrowthPlot(pslStats)
	if err != nil {
		log.Printf("Error creating PSL growth plot: %v", err)
	}

	// Create combined grid: 2 rows x 3 columns
	// Row 1: Insert, Lookup Hit, Lookup Miss
	// Row 2: Load Factor, PSL Statistics, PSL Growth
	plots := []*plot.Plot{
		insertPlot, lookupHitPlot, lookupMissPlot,
		loadFactorPlot, pslStatsPlot, pslGrowthPlot,
	}

	if err := createCombinedGrid(plots, filepath.Join(outputDir, "benchmark_grid.png")); err != nil {
		log.Printf("Error creating combined grid: %v", err)
	}

	// Also create summary bar chart separately (useful as standalone)
	fmt.Println("Creating summary chart...")
	summaryResults := []BenchmarkResult{
		{Name: "Insert\n(10K)", LP: insertResults[4].LP, RH: insertResults[4].RH},
		{Name: "Lookup\nHit", LP: lookupHitResults[4].LP, RH: lookupHitResults[4].RH},
		{Name: "Lookup\nMiss", LP: lookupMissResults[4].LP, RH: lookupMissResults[4].RH},
		{Name: "High\nLoad", LP: loadFactorResults[4].LP, RH: loadFactorResults[4].RH},
	}
	if err := createBarChart(summaryResults, "Performance Summary: Linear Probing vs Robin Hood",
		"Time (ns/op)", filepath.Join(outputDir, "summary_comparison.png")); err != nil {
		log.Printf("Error creating summary chart: %v", err)
	}

	fmt.Println()
	fmt.Printf("Plots saved to: %s/\n", outputDir)
	fmt.Println("  - benchmark_grid.png (6 plots in 2 rows x 3 columns)")
	fmt.Println("  - summary_comparison.png")
}
