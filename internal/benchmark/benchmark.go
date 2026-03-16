package benchmark

import (
	"fmt"
	"time"
)

type Result struct {
	OperationCount  int
	TotalDuration   time.Duration
	AvgLatencyNs    int64
	AvgLatencyUs    float64
	OpsPerSecond    float64
	MinLatencyNs    int64
	MaxLatencyNs    int64
}

type PutFunc func(key []byte, value []byte) error
type GetFunc func(key []byte) ([]byte, error)

func RunWriteBenchmark(put PutFunc, numOps int) (*Result, error) {

	fmt.Println("Running write benchmark with", numOps, "operations...")

	latencies := make([]int64, numOps)
	start := time.Now()

	for i := 0; i < numOps; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i))
		value := []byte(fmt.Sprintf("bench_value_%d", i))

		opStart := time.Now()
		if err := put(key, value); err != nil {
			return nil, err
		}
		latencies[i] = time.Since(opStart).Nanoseconds()
	}

	total := time.Since(start)

	return calculateResult(latencies, numOps, total), nil
}

func RunReadBenchmark(get GetFunc, numOps int) (*Result, error) {

	fmt.Println("Running read benchmark with", numOps, "operations...")

	latencies := make([]int64, numOps)
	start := time.Now()

	for i := 0; i < numOps; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i))

		opStart := time.Now()
		get(key)
		latencies[i] = time.Since(opStart).Nanoseconds()
	}

	total := time.Since(start)

	return calculateResult(latencies, numOps, total), nil
}

func RunMapWriteBenchmark(numOps int) *Result {

	fmt.Println("Running map(btree approx) write benchmark with", numOps, "operations...")

	m := make(map[string]string)
	latencies := make([]int64, numOps)
	start := time.Now()

	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := fmt.Sprintf("bench_value_%d", i)

		opStart := time.Now()
		m[key] = value
		latencies[i] = time.Since(opStart).Nanoseconds()
	}

	total := time.Since(start)

	return calculateResult(latencies, numOps, total)
}

func RunMapReadBenchmark(numOps int) *Result {

	fmt.Println("Running map(btree approx) read benchmark with", numOps, "operations...")

	m := make(map[string]string)
	for i := 0; i < numOps; i++ {
		m[fmt.Sprintf("bench_key_%d", i)] = fmt.Sprintf("bench_value_%d", i)
	}

	latencies := make([]int64, numOps)
	start := time.Now()

	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		opStart := time.Now()
		_ = m[key]
		latencies[i] = time.Since(opStart).Nanoseconds()
	}

	total := time.Since(start)

	return calculateResult(latencies, numOps, total)
}

func calculateResult(latencies []int64, numOps int, total time.Duration) *Result {

	var sum int64
	minL := latencies[0]
	maxL := latencies[0]

	for _, l := range latencies {
		sum += l
		if l < minL {
			minL = l
		}
		if l > maxL {
			maxL = l
		}
	}

	avg := sum / int64(numOps)
	ops := float64(numOps) / total.Seconds()

	return &Result{
		OperationCount: numOps,
		TotalDuration:  total,
		AvgLatencyNs:   avg,
		AvgLatencyUs:   float64(avg) / 1000.0,
		OpsPerSecond:   ops,
		MinLatencyNs:   minL,
		MaxLatencyNs:   maxL,
	}
}

func FormatResult(name string, r *Result) string {
	return fmt.Sprintf(
		"%s | ops: %d | avg: %.2fµs | min: %dns | max: %dns | throughput: %.0f ops/sec",
		name,
		r.OperationCount,
		r.AvgLatencyUs,
		r.MinLatencyNs,
		r.MaxLatencyNs,
		r.OpsPerSecond,
	)
}