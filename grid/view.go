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

type Sheet struct {
	Size layout.Dimension

	rows  []*Row
	cells map[layout.Position]*Cell
}
