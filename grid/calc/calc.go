package calc

import (
	"math"
	"math/rand"
	"sort"
)

func Sign(value float64) float64 {
	if value < 0 {
		return -1
	}
	return 1
}

func Odd(value float64) bool {
	return math.Mod(value, 2) == 1
}

func Even(value float64) bool {
	return math.Mod(value, 2) == 0
}

func Min(values []float64) float64 {
	var ret float64
	for i := 0; i < len(values); i++ {
		if i == 0 {
			ret = values[i]
			continue
		}
		ret = min(ret, values[i])
	}
	return ret
}

func Max(values []float64) float64 {
	var ret float64
	for i := 0; i < len(values); i++ {
		ret = max(ret, values[i])
	}
	return ret
}

func Sum(values []float64) float64 {
	var ret float64
	for i := 0; i < len(values); i++ {
		ret += values[i]
	}
	return ret
}

func Median(values []float64) float64 {
	sort.Float64s(values)

	size := len(values)
	switch size {
	case 0:
		return 0
	case 1:
		return values[0]
	default:
	}
	ix := len(values) / 2
	if ix%2 == 0 {
		return values[ix]
	}
	fd := values[(size/2)-1]
	td := values[size/2]
	return (td - fd) / 2
}

func Avg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := Sum(values)
	return total / float64(len(values))
}

func Mode(values []float64) float64 {
	vs := make(map[float64]int)
	for _, f := range values {
		vs[f]++
	}
	if len(vs) == 0 {
		return 0
	}
	var (
		ms = make(map[int][]float64)
		mx int
	)
	for v, c := range vs {
		mx = max(c, mx)
		ms[c] = append(ms[c], v)
	}
	rs := ms[mx]
	return rs[0]
}

func Stdev(values []float64) float64 {
	v := Var(values)
	return math.Sqrt(v)
}

func Var(values []float64) float64 {
	z := len(values)
	if z < 2 {
		return 0
	}
	var (
		avg = Avg(values)
		sum float64
	)
	for _, f := range values {
		x := (f - avg) * (f - avg)
		sum += x
	}
	return sum / float64(z)
}

func Deg(value float64) float64 {
	return value * (180 / math.Pi)
}

func Rad(value float64) float64 {
	return value * (math.Pi / 180)
}

func Pi() float64 {
	return math.Pi
}

func Rand() float64 {
	return rand.Float64()
}
