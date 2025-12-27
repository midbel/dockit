package oxml

import (
	"fmt"
)

type Dimension struct {
	Lines   int64
	Columns int64
}

type Bounds struct {
	Start Position
	End   Position
}

func (b Bounds) String() string {
	if b.Start.Equal(b.End) {
		return b.Start.Addr()
	}
	return fmt.Sprintf("%s:%s", b.Start.Addr(), b.End.Addr())
}

type Position struct {
	Line   int64
	Column int64
}

func parsePosition(addr string) Position {
	cell, err := parseCellAddr(addr)
	if err != nil {
		return Position{}
	}
	return cell.Position
}

func (p Position) Equal(other Position) bool {
	return p.Line == other.Line && p.Column == other.Column
}

func (p Position) Addr() string {
	addr := cellAddr{
		Position: p,
	}
	return addr.String()
}
