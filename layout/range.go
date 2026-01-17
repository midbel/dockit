package layout

import (
	"fmt"
	"strings"
)

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
	starts = ParsePosition(fst)
	if ok {
		ends = ParsePosition(lst)
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
	return r.Ends.Column - r.Starts.Column
}

func (r *Range) Height() int64 {
	return r.Ends.Line - r.Starts.Line
}

func (r *Range) String() string {
	if r.Starts.Equal(r.Ends) {
		return r.Starts.Addr()
	}
	return fmt.Sprintf("%s:%s", r.Starts.Addr(), r.Ends.Addr())
}

func (r *Range) Normalize() *Range {
	x := NewRange(r.Starts, r.Ends)
	x.Starts.Line = min(r.Starts.Line, r.Ends.Line)
	x.Starts.Column = min(r.Starts.Column, r.Ends.Column)
	x.Ends.Line = max(r.Starts.Line, r.Ends.Line)
	x.Ends.Column = max(r.Starts.Column, r.Ends.Column)
	return x
}

func (r *Range) Range() *Range {
	start := Position{
		Line:   1,
		Column: 1,
	}
	if r.Width() == 0 && r.Height() == 0 {
		return NewRange(start, start)
	}
	end := Position{
		Line:   r.Height(),
		Column: r.Width(),
	}
	return NewRange(start, end)
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
