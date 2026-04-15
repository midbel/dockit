package grid

import (
	"fmt"
	
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type CopyMode int

func CopyModeFromString(str string) (CopyMode, error) {
	var mode CopyMode
	switch str {
	case "value":
		mode |= CopyValue
	case "formula":
		mode |= CopyFormula
	case "style":
		mode |= CopyStyle
	case "", "all":
		mode |= CopyAll
	default:
		return mode, fmt.Errorf("%s invalid value for copy mode", str)
	}
	return mode, nil
}

const (
	CopyValue = iota << 1
	CopyFormula
	CopyStyle
	CopyAll = CopyValue | CopyFormula | CopyStyle
)

type Row interface {
	Values() []value.ScalarValue
	Sparse() bool
}

type Cell interface {
	At() layout.Position
	Value() value.ScalarValue
	Formula() value.Formula
	Dirty() bool
}

func ResetAt(cell Cell, pos layout.Position) Cell {
	if a, ok := cell.(interface{ SetAt(layout.Position) }); ok {
		a.SetAt(pos)
	}
	return cell
}

type empty struct {
	pos   layout.Position
	value value.ScalarValue
}

func Single(val value.ScalarValue, pos layout.Position) Cell {
	return empty{
		pos:   pos,
		value: val,
	}
}

func Empty(pos layout.Position) Cell {
	return empty{
		pos: pos,
	}
}

func (c empty) At() layout.Position {
	return c.pos
}

func (c empty) Value() value.ScalarValue {
	if c.value == nil {
		return value.Empty()
	}
	return c.value
}

func (empty) Formula() value.Formula {
	return nil
}

func (empty) Dirty() bool {
	return false
}
