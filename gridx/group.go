package gridx

import (
	"fmt"
	"iter"
	"maps"
	"math"
	"slices"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Aggregator interface {
	Reset()
	Aggr(value.Value)
	Result() value.Value

	clone() Aggregator
}

type Aggr struct {
	Aggregator
	Column int64
}

func CreateAggr(col, aggr string) (*Aggr, error) {
	var a Aggr
	a.Column, _ = layout.ParseIndex(col)

	fn, ok := aggrBuilder[aggr]
	if !ok {
		return nil, fmt.Errorf("%s: unknown aggregate function", aggr)
	}
	a.Aggregator = fn()
	return &a, nil
}

func NewAggr(col int64, aggr Aggregator) *Aggr {
	return &Aggr{
		Column:     col,
		Aggregator: aggr,
	}
}

func (a *Aggr) Update(row []value.ScalarValue) {
	col := a.Column - 1
	if col < 0 || int(col) >= len(row) {
		return
	}
	a.Aggregator.Aggr(row[int(col)])
}

func (a *Aggr) Clone() *Aggr {
	x := NewAggr(a.Column, a.Aggregator.clone())
	return x
}

func Group(view grid.View, keys layout.Selection, aggr []Aggr) (grid.View, error) {
	groups, err := createGroups(view, keys, aggr)
	if err != nil {
		return nil, err
	}
	vs := maps.Values(groups)
	return newGroupedView(view, slices.Collect(vs)), nil
}

func createGroups(view grid.View, keys layout.Selection, aggr []Aggr) (map[string]*groupRow, error) {
	var (
		cols   = keys.Indices(view.Bounds())
		groups = make(map[string]*groupRow)
	)
	for lino, rs := range view.Rows() {
		k := keyFromRow(rs, cols)

		gr, ok := groups[k]
		if !ok {
			row := new(groupRow)
			for _, c := range cols {
				row.values = append(row.values, rs[c])
			}
			for _, a := range aggr {
				row.aggr = append(row.aggr, a.Clone())
			}
			gr = row
			groups[k] = gr
		}
		gr.Update(rs)
		gr.indices = append(gr.indices, lino)
	}
	return groups, nil
}

type groupedView struct {
	view   grid.View
	groups []*groupRow
}

func newGroupedView(view grid.View, rows []*groupRow) grid.View {
	v := groupedView{
		view:   view,
		groups: rows,
	}
	return &v
}

func (g *groupedView) Name() string {
	return g.view.Name()
}

func (g *groupedView) Bounds() *layout.Range {
	var (
		start = layout.NewPosition(1, 1)
		end   = layout.NewPosition(int64(len(g.groups)), g.groups[0].Columns())
	)
	return layout.NewRange(start, end)
}

func (g *groupedView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		for lino, r := range g.groups {
			ok := yield(int64(lino)+1, r.Values())
			if !ok {
				return
			}
		}
	}
	return it
}

func (g *groupedView) Cell(pos layout.Position) (grid.Cell, error) {
	return grid.Empty(pos), nil
}

func (g *groupedView) Sync(ctx value.Context) error {
	if err := g.view.Sync(ctx); err != nil {
		return err
	}
	return grid.ErrSupported
}

type groupRow struct {
	values  []value.Value
	aggr    []*Aggr
	indices []int64
}

func (r *groupRow) Update(row []value.ScalarValue) {
	for i := range r.aggr {
		r.aggr[i].Update(row)
	}
}

func (r *groupRow) Columns() int64 {
	t := len(r.values) + len(r.aggr)
	return int64(t)
}

func (r *groupRow) Values() []value.ScalarValue {
	out := make([]value.ScalarValue, 0, int(r.Columns()))
	for _, v := range r.values {
		if value.IsScalar(v) {
			out = append(out, v.(value.ScalarValue))
		} else {
			out = append(out, value.Empty())
		}
	}
	for _, v := range r.aggr {
		res := v.Result()
		if value.IsScalar(res) {
			out = append(out, res.(value.ScalarValue))
		} else {
			out = append(out, value.Empty())
		}
	}
	return out
}

var aggrBuilder = map[string]func() Aggregator{
	"min":      Min,
	"max":      Max,
	"avg":      Avg,
	"avergage": Avg,
	"sum":      Sum,
	"count":    Count,
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

func (*minv) clone() Aggregator {
	return Min()
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

func (*maxv) clone() Aggregator {
	return Max()
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

func (*count) clone() Aggregator {
	return Count()
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

func (*avg) clone() Aggregator {
	return Avg()
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

func (*sum) clone() Aggregator {
	return Sum()
}
