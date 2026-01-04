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
	Contains(Position) bool
}

func ParseSelection(str string) (Selection, error) {
	var (
		list  []Selection
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

type RangeSet struct {
	list []Selection
}

func (r *RangeSet) Contains(pos Position) bool {
	for _, s := range r.list {
		if ok := s.Contains(pos); ok {
			return true
		}
	}
	return false
}
