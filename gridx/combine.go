package gridx

import (
	"fmt"
	"iter"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type unionRow struct {
	view grid.View
	lino int64
}

func (u unionRow) At(pos layout.Position) (grid.Cell, error) {
	c, _ := u.view.Cell(layout.NewPosition(u.lino, pos.Column))
	return grid.ResetAt(c, pos), nil
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
	if len(v.rows) == 0 {
		pos := layout.NewPosition(1, 1)
		return layout.NewRange(pos, pos)
	}
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
				p := layout.NewPosition(r.lino, int64(i+1))
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
	if pos.Line < 1 || pos.Line > int64(len(v.rows)) {
		return grid.Empty(pos), nil
	}
	return v.rows[int(pos.Line)].At(pos)
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
		delete(rows, key)
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
	view grid.View
	rows []unionRow
}

func (v *intersectView) Name() string {
	return v.view.Name()
}

func (v *intersectView) Bounds() *layout.Range {
	if len(v.rows) == 0 {
		pos := layout.NewPosition(1, 1)
		return layout.NewRange(pos, pos)
	}
	var (
		bd    = v.view.Bounds()
		start = layout.NewPosition(1, 1)
		end   = layout.NewPosition(int64(len(v.rows)), bd.Width())
	)
	return layout.NewRange(start, end)
}

func (v *intersectView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		bd := v.view.Bounds()
		for i, r := range v.rows {
			out := make([]value.ScalarValue, 0, int(bd.Width()))
			for c := range bd.Width() {
				p := layout.NewPosition(r.lino, c+1)
				x, _ := r.view.Cell(p)
				out = append(out, x.Value())
			}
			ok := yield(int64(i+1), out)
			if !ok {
				return
			}
		}
	}
	return it
}

func (v *intersectView) Cell(pos layout.Position) (grid.Cell, error) {
	if pos.Line < 1 || pos.Line > int64(len(v.rows)) {
		return grid.Empty(pos), nil
	}
	return v.rows[int(pos.Line)].At(pos)
}

func (v *intersectView) Sync(ctx value.Context) error {
	return grid.ErrSupported
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
		rows[key] = struct{}{}
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
	view grid.View
	rows []unionRow
}

func (v *exceptView) Name() string {
	return v.view.Name()
}

func (v *exceptView) Bounds() *layout.Range {
	if len(v.rows) == 0 {
		pos := layout.NewPosition(1, 1)
		return layout.NewRange(pos, pos)
	}
	var (
		bd    = v.view.Bounds()
		start = layout.NewPosition(1, 1)
		end   = layout.NewPosition(int64(len(v.rows)), bd.Width())
	)
	return layout.NewRange(start, end)
}

func (v *exceptView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		bd := v.view.Bounds()
		for i, r := range v.rows {
			out := make([]value.ScalarValue, 0, int(bd.Width()))
			for c := range bd.Width() {
				p := layout.NewPosition(r.lino, c+1)
				x, _ := r.view.Cell(p)
				out = append(out, x.Value())
			}
			ok := yield(int64(i+1), out)
			if !ok {
				return
			}
		}
	}
	return it
}

func (v *exceptView) Cell(pos layout.Position) (grid.Cell, error) {
	if pos.Line < 1 || pos.Line > int64(len(v.rows)) {
		return grid.Empty(pos), nil
	}
	return v.rows[int(pos.Line)].At(pos)
}

func (v *exceptView) Sync(ctx value.Context) error {
	return grid.ErrSupported
}
