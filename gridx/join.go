package gridx

import (
	"fmt"
	"iter"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type joinView struct {
	left  grid.View
	right grid.View

	// index
}

func Join(left, right grid.View) grid.View {
	return nil
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
