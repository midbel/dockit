package gridx

import (
	"fmt"
	"iter"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type joinRow struct {
	Left  int64
	Right int64
}

type joinView struct {
	left  grid.View
	right grid.View

	rows []joinRow
}

func Join(left, right grid.View, leftcols, rightcols layout.Selection) grid.View {
	var (
		index = createIndex(right, rightcols)
		rows  = createLinks(left, leftcols, index)
	)

	j := joinView{
		left:  left,
		right: right,
		rows:  rows,
	}
	return &j
}

func (v *joinView) Name() string {
	return fmt.Sprintf("%s:%s", v.left.Name(), v.right.Name())
}

func (v *joinView) Bounds() *layout.Range {
	return nil
}

func (v *joinView) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue)) {

	}
	return it
}

func (v *joinView) Cell(layout.Position) (Cell, error) {
	return nil
}

func (v *joinView) Sync(value.Context) error {
	return grid.ErrSupported
}

func keyFromRow(row []value.ScalarValue, cols []int64) string {
	var ks []string
	var b strings.Builder
	for i, c := range cols {
		if i > 0 {
			b.WriteRune('|')
		}
		k := createKey(row[c])
		b.WriteString(k)
	}
	return b.String()
}

func createLinks(view grid.View, keys layout.Selection, index map[string][]int64) []joinRow {
	var (
		lino int64
		rows []joinRow
		cols = keys.Indices(view.Bounds())
	)

	for rs := range view.Rows() {
		lino++
		k := keyFromRow(rs, cols)

		matches := index[k]
		for _, m := range matches {
			r := joinRow{
				Left:  lino,
				Right: m,
			}
			rows = append(rows, r)
		}
	}
	return rows
}

func createIndex(view grid.View, keys layout.Selection) map[string][]int64 {
	var (
		lino  int
		cols  = keys.Indices(view.Bounds())
		index = make(map[string][]int64)
	)
	for rs := range right.Rows() {
		lino++
		k := keyFromRow(rs, cols)
		index[k] = append(index[k], lino)
	}
	return index
}

func createKey(v value.Value) string {
	var prefix string
	switch v.Type() {
	case value.TypeNumber:
		prefix = "n"
	case value.TypeText:
		prefix = "s"
	case value.TypeBool:
		prefix = "b"
	case value.TypeDate:
		prefix = "d"
	case value.TypeError:
		prefix = "e"
	case value.TypeBlank:
		prefix = "b"
	default:
		prefix = "?"
	}
	return fmt.Sprintf("%s:%s", prefix, v.String())
}
