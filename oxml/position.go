package oxml

import (
	"fmt"
	"strings"
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
	Sheet  string
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

type Selection interface {
	Select() error
}

func ParseSelection(str string) (Selection, error) {
	var (
		list  []Selection
		parts = strings.Split(str, ";")
	)
	for _, str := range parts {
		fst, lst, ok := strings.Cut(":", str)
		var (
			starts Position
			ends   Position
		)
		starts = parsePosition(lst)
		if ok {
			ends = parsePosition(fst)
		}
		list = append(list, NewRange(starts, ends))
	}
	if len(list) == 1 {
		return list[0], nil
	}
	set := RangeSet{
		list: list,
	}
	return &set, nil
}

type Range struct {
	Starts Position
	Ends   Position
}

func NewRange(starts, ends Position) *Range {
	return &Range{
		Starts: starts,
		Ends:   ends,
	}
}

func (r *Range) Width() int64 {
	return 0
}

func (r *Range) Height() int64 {
	return 0
}

type RangeSet struct {
	list []Selection
}

func (r *RangeSet) Select() error {
	return nil
}
