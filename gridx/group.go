package gridx

import (
	"math"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Aggregator interface {
	Reset()
	Aggr(value.Value)
	Result() value.Value
}

func Group(view grid.View, keys layout.Selection, aggr []Aggregator) (grid.View, error) {
	return nil, nil
}

type minv struct {
	result float64
}

func Min() Aggregator {
	return new(minv)
}

func (a *minv) Reset() {
	a.result = 0
}

func (a *minv) Aggr(val value.Value) {
	if math.IsNaN(a.result) {
		return
	}
	f, err := value.CastToFloat(val)
	if err != nil {
		a.result = math.NaN()
		return
	}
	a.result = min(a.result, float64(f))
}

func (a *minv) Result() value.Value {
	if math.IsNaN(a.result) {
		return value.ErrValue
	}
	return value.Float(a.result)
}

type maxv struct {
	result float64
}

func Max() Aggregator {
	return new(maxv)
}

func (a *maxv) Reset() {
	a.result = 0
}

func (a *maxv) Aggr(val value.Value) {
	if math.IsNaN(a.result) {
		return
	}
	f, err := value.CastToFloat(val)
	if err != nil {
		a.result = math.NaN()
		return
	}
	a.result = max(a.result, float64(f))
}

func (a *maxv) Result() value.Value {
	if math.IsNaN(a.result) {
		return value.ErrValue
	}
	return value.Float(a.result)
}

type count struct {
	result int64
}

func Count() Aggregator {
	return new(count)
}

func (a *count) Reset() {
	a.result = 0
}

func (a *count) Aggr(_ value.Value) {
	a.result++
}

func (a *count) Result() value.Value {
	return value.Float(a.result)
}

type avg struct {
	result float64
	count  int
}

func Avg() Aggregator {
	return new(avg)
}

func (a *avg) Reset() {
	a.result = 0
	a.count = 0
}

func (a *avg) Aggr(val value.Value) {
	if math.IsNaN(a.result) {
		return
	}
	f, err := value.CastToFloat(val)
	if err != nil {
		a.result = math.NaN()
		return
	}
	a.count++
	a.result += float64(f)
}

func (a *avg) Result() value.Value {
	if math.IsNaN(a.result) {
		return value.ErrValue
	}
	if a.count == 0 {
		return value.ErrDiv0
	}
	res := a.result / float64(a.count)
	return value.Float(res)
}

type sum struct {
	result float64
}

func Sum() Aggregator {
	return new(sum)
}

func (a *sum) Reset() {
	a.result = 0
}

func (a *sum) Aggr(val value.Value) {
	if math.IsNaN(a.result) {
		return
	}
	f, err := value.CastToFloat(val)
	if err != nil {
		a.result = math.NaN()
		return
	}
	a.result += float64(f)
}

func (a *sum) Result() value.Value {
	if math.IsNaN(a.result) {
		return value.ErrValue
	}
	return value.Float(a.result)
}
