package gridx

import (
	"fmt"
	"iter"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
	"github.com/midbel/dockit/layout"
)

type unionRow struct {
	view grid.View
	lino int64
}

func Union(left, right grid.View) (grid.View, error) {
	var (
		lbd = left.Bounds()
		rbd = right.Bounds()
	)
	if lbd.Width() != rbd.Width() {
		return nil, fmt.Errorf("columns count mismatched")
	}
	var (
		rows = make(map[string]struct{})
		list []unionRow
	)
	for lino, rs := range left.Rows() {
		key := keyFromValues(rs)
		if _, ok := rows[key]; ok {
			continue
		}
		rows[key] = struct{}{}
		r := unionRow{
			view: left,
			lino: lino,
		}
		list = append(list, r)
	}
	for lino, rs := range right.Rows() {
		key := keyFromValues(rs)
		if _, ok := rows[key]; ok {
			continue
		}
		rows[key] = struct{}{}
		r := unionRow{
			view: right,
			lino: lino,
		}
		list = append(list, r)
	}
	v := unionView{
		left:  left,
		right: right,
		rows:  list,
	}
	return &v, nil
}

type unionView struct {
	left  grid.View
	right grid.View
	rows  []unionRow
}

func (v *unionView) Name() string {
	return v.left.Name()
}

func (v *unionView) Bounds() *layout.Range {
	var (
		lbd   = v.left.Bounds()
		rbd   = v.right.Bounds()
		width = max(lbd.Width(), rbd.Width())
		start = layout.NewPosition(1, 1)
		end   = layout.NewPosition(int64(len(v.rows)), width)
	)
	return layout.NewRange(start, end)
}

func (v *unionView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		for i, r := range v.rows {
			var (
				bd  = r.view.Bounds()
				out = make([]value.ScalarValue, 0, bd.Width())
			)
			for i := range bd.Width() {
				p := layout.NewPosition(r.lino, int64(i))
				c, _ := r.view.Cell(p)

				out = append(out, c.Value())
			}
			if !yield(int64(i+1), out) {
				return
			}
		}
	}
	return it
}

func (v *unionView) Cell(pos layout.Position) (grid.Cell, error) {
	return grid.Empty(pos), nil
}

func (v *unionView) Sync(ctx value.Context) error {
	return grid.ErrSupported
}

func Intersect(left, right grid.View) (grid.View, error) {
	var (
		lbd = left.Bounds()
		rbd = right.Bounds()
	)
	if lbd.Width() != rbd.Width() {
		return nil, fmt.Errorf("columns count mismatched")
	}
	var (
		rows = make(map[string]struct{})
		list []unionRow
	)
	for _, rs := range right.Rows() {
		key := keyFromValues(rs)
		rows[key] = struct{}{}
	}
	for lino, rs := range left.Rows() {
		key := keyFromValues(rs)
		_, ok := rows[key]
		if !ok {
			continue
		}
		r := unionRow{
			lino: lino,
			view: left,
		}
		list = append(list, r)
	}
	v := intersectView{
		view: left,
		rows: list,
	}
	return &v, nil
}

type intersectView struct {
	view  grid.View
	rows []unionRow
}

func (v *intersectView) Name() string {
	return v.view.Name()
}

func (v *intersectView) Bounds() *layout.Range {
	return nil
}

func (v *intersectView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	return nil
}

func (v *intersectView) Cell(pos layout.Position) (grid.Cell, error) {
	return nil, nil
}

func (v *intersectView) Sync(ctx value.Context) error {
	return nil
}

func Except(left, right grid.View) (grid.View, error) {
	var (
		lbd = left.Bounds()
		rbd = right.Bounds()
	)
	if lbd.Width() != rbd.Width() {
		return nil, fmt.Errorf("columns count mismatched")
	}
	var (
		rows = make(map[string]struct{})
		list []unionRow
	)
	for _, rs := range right.Rows() {
		key := keyFromValues(rs)
		rows[key] = struct{}{}
	}
	for lino, rs := range left.Rows() {
		key := keyFromValues(rs)
		_, ok := rows[key]
		if ok {
			continue
		}
		r := unionRow{
			lino: lino,
			view: left,
		}
		list = append(list, r)
	}
	v := exceptView{
		view: left,
		rows: list,
	}
	return &v, nil
}

type exceptView struct {
	view  grid.View
	rows []unionRow
}

func (v *exceptView) Name() string {
	return v.view.Name()
}

func (v *exceptView) Bounds() *layout.Range {
	return nil
}

func (v *exceptView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	return nil
}

func (v *exceptView) Cell(pos layout.Position) (grid.Cell, error) {
	return nil, nil
}

func (v *exceptView) Sync(ctx value.Context) error {
	return nil
}
