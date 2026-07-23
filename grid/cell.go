package grid

import (
	"fmt"

	"github.com/midbel/dockit/internal/id"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

func NoCell(pos layout.Position) error {
	return fmt.Errorf("%s no cell at given position", pos)
}

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

func (c CopyMode) Valid() bool {
	return c != 0
}

func (c CopyMode) Value() bool {
	return c == CopyAll || (c&CopyValue != 0)
}

func (c CopyMode) Formula() bool {
	return c == CopyAll || (c&CopyFormula != 0)
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
	Id() uint64
	At() layout.Position
	Value() value.Value
	Formula() value.Formula
	Dirty() bool

	// SetValue(value.Value)
	// SetFormula(value.Formula)
}

type proxyCell struct {
	Cell
	layout.Position
}

func ResetAt(cell Cell, pos layout.Position) Cell {
	c := proxyCell{
		Position: pos,
		Cell:     cell,
	}
	return c
}

type empty struct {
	pos   layout.Position
	id    uint64
	value value.Value
}

func Single(val value.Value, pos layout.Position) Cell {
	return empty{
		id:    id.Next(),
		pos:   pos,
		value: val,
	}
}

func Empty(pos layout.Position) Cell {
	return Single(nil, pos)
}

func (c empty) Id() uint64 {
	return c.id
}

func (c empty) At() layout.Position {
	return c.pos
}

func (c empty) Value() value.Value {
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
