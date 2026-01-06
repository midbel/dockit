package oxml

import (
	"fmt"
	"strings"
)

type Dimension struct {
	Lines   int64
	Columns int64
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

func (p Position) Update(other Position) Position {
	if p.Line == 0 {
		p.Line = other.Line
	}
	if p.Column == 0 {
		p.Column = other.Column
	}
	return p
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

func (r *Range) Open() bool {
	var zero Position
	return r.Starts == zero || r.Ends == zero
}

func (r *Range) Contains(pos Position) bool {
	ok := pos.Line >= r.Starts.Line && pos.Line <= r.Ends.Line
	if !ok {
		return false
	}
	return pos.Column >= r.Starts.Column && pos.Column <= r.Ends.Column
}

func (r *Range) Width() int64 {
	return r.Ends.Line - r.Starts.Line
}

func (r *Range) Height() int64 {
	return r.Ends.Column - r.Starts.Column
}

func (r *Range) String() string {
	if r.Starts.Equal(r.Ends) {
		return r.Starts.Addr()
	}
	return fmt.Sprintf("%s:%s", r.Starts.Addr(), r.Ends.Addr())
}

type RangeSet struct {
	list []*Range
}

func RangeSetFromString(str string) (*RangeSet, error) {
	var (
		list  []*Range
		parts = strings.Split(str, ";")
	)
	for _, str := range parts {
		fst, lst, ok := strings.Cut(str, ":")
		var (
			starts Position
			ends   Position
		)
		starts = parsePosition(fst)
		if ok {
			ends = parsePosition(lst)
		}
		list = append(list, NewRange(starts, ends))
	}
	set := RangeSet{
		list: list,
	}
	return &set, nil
}
