package aggregator

import (
	"math"
	"sort"
	"time"
)

type Percentiles struct {
	P50 time.Duration
	P95 time.Duration
	P99 time.Duration
}

func ComputePercentiles(samples []time.Duration) Percentiles {
	if len(samples) == 0 {
		return Percentiles{}
	}
	values := append([]time.Duration(nil), samples...)
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return Percentiles{
		P50: pick(values, 0.50),
		P95: pick(values, 0.95),
		P99: pick(values, 0.99),
	}
}

func pick(sorted []time.Duration, q float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if q <= 0 {
		return sorted[0]
	}
	if q >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := int(math.Ceil(q*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
