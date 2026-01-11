package grid

import (
	"iter"

	"github.com/midbel/dockit/formula"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Encoder interface {
	EncodeSheet(View) error
}

type View interface {
	Bounds() *layout.Range
	Rows() iter.Seq[[]value.ScalarValue]
	Encode(Encoder) error
}

type Cell struct {
	layout.Position

	Raw     string
	Parsed  value.ScalarValue
	Formula formula.Expr
	Dirty   bool
}

func (c *Cell) Reload(ctx formula.Context) error {
	if c.Formula == nil {
		return nil
	}
	res, err := formula.Eval(c.Formula, ctx)
	if err == nil {
		if !formula.IsScalar(res) {
			c.Parsed = formula.ErrValue
		} else {
			c.Parsed = res.(value.ScalarValue)
		}
		c.Raw = res.String()
	}
	return err
}

type Row struct {
	Line   int64
	Hidden bool
	Cells  []*Cell
}

func (r *Row) Sparse() bool {
	for i, c := range r.Cells {
		if i == 0 {
			return
		}
		if r.Cells[i-1].Column - c.Column > 1 {
			return true
		}
	}
	return false
}

func (r *Row) Values() []value.ScalarValue {
	var list []value.ScalarValue
	for i := range r.Cells {
		list = append(list, r.Cells[i].Parsed)
	}
	return list
}

func (r *Row) cloneCells() []*Cell {
	var cells []*Cell
	for i := range r.Cells {
		c := *r.Cells[i]
		cells = append(cells, &c)
	}
	return cells
}

type Sheet struct {
	Size layout.Dimension

	rows  []*Row
	cells map[layout.Position]*Cell
}
