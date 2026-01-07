package oxml

import (
	"fmt"
	"slices"
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

func RangeFromString(str string) *Range {
	fst, lst, ok := strings.Cut(str, ":")
	var (
		starts Position
		ends   Position
	)
	starts = parsePosition(fst)
	if ok {
		ends = parsePosition(lst)
	}
	return NewRange(starts, ends)
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

func (r *Range) normalize() *Range {
	x := NewRange(r.Starts, r.Ends)
	x.Starts.Line = min(r.Starts.Line, r.Ends.Line)
	x.Starts.Column = min(r.Starts.Column, r.Ends.Column)
	x.Ends.Line = max(r.Starts.Line, r.Ends.Line)
	x.Ends.Column = max(r.Starts.Column, r.Ends.Column)
	return x
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
		rg := RangeFromString(str)
		list = append(list, rg)
	}
	set := RangeSet{
		list: list,
	}
	return &set, nil
}

type Selection interface {
	Indices(*Range) []int64
}

func SelectionFromString(str string) (Selection, error) {
	var (
		list []Selection
		parts = strings.Split(str, ";")
	)
	for _, str := range parts {
		str = strings.TrimSpace(str)
		fst, lst, ok := strings.Cut(str, ":")
		if ok {

		} else {
			ix, err := strconv.ParseInt(fst, 10, 64)
			if err != nil {
				return nil, err
			}
			c := columnRef{
				Index: ix,
			}
			list = append(list, c)
		}
	}
	return list, nil
}

type columnRef struct {
	Index int64
}

func (c columnRef) Indices(rg *Range) []int64 {
	if rg == nil || rg.Open() {
		return nil
	}
	if c.Index >= rg.Starts.Column && c.Index <= rg.Ends.Column {
		return []int64{c.Index}
	}
	return nil
}

type columnSpan struct {
	Starts int64
	Ends   int64
	Step   int64
}

func (c columnSpan) Indices(rg *Range) []int64 {
	if rg == nil || rg.Open() {
		return nil
	}
	var (
		all     []int64
		starts  = c.Starts
		ends    = c.Ends
		reverse = c.Starts > c.Ends
	)
	if c.Step == 0 {
		c.Step = 1
	}
	starts = max(starts, rg.Starts.Column)
	starts = min(starts, rg.Ends.Column)
	ends = max(ends, rg.Starts.Column)
	ends = min(ends, rg.Ends.Column)

	for i := starts; i <= ends; i += c.Step {
		all = append(all, i)
	}
	if reverse {
		slices.Reverse(all)
	}
	return all
}

type combinedRef struct {
	list []Selection
}

func (r combinedRef) Indices(rg *Range) []int64 {
	var all []int64
	for i := range r.list {
		all = slices.Concat(all, r.list[i].Indices(rg))
	}
	return all
}
