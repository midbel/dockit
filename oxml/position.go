package oxml

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

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
		list  []Selection
		parts = strings.Split(str, ";")
	)
	for _, str := range parts {
		parts := strings.Split(strings.TrimSpace(str), ":")
		switch n := len(parts); n {
		case 1:
			ix, _ := parseIndex(parts[0])
			list = append(list, columnRef{
				Index: ix,
			})
		case 2, 3:
			lo, _ := parseIndex(parts[0])
			hi, _ := parseIndex(parts[1])
			ref := columnSpan{
				Starts: lo,
				Ends:   hi,
				Step:   1,
			}
			if n == 3 && parts[2] != "" {
				st, err := strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					return nil, err
				}
				ref.Step = st
			}
			list = append(list, ref)
		default:
			return nil, fmt.Errorf("selection: invalid syntax")
		}
	}
	if len(list) == 1 {
		return list[0], nil
	}
	combined := combinedRef{
		list: list,
	}
	return combined, nil
}

type columnRef struct {
	Index int64
}

func (c columnRef) Indices(rg *Range) []int64 {
	if rg == nil {
		return nil
	}
	if c.Index >= rg.Starts.Column && c.Index <= rg.Ends.Column {
		return []int64{c.Index - 1}
	}
	return nil
}

type columnSpan struct {
	Starts int64
	Ends   int64
	Step   int64
}

func (c columnSpan) Indices(rg *Range) []int64 {
	if rg == nil {
		return nil
	}
	var (
		all     []int64
		step    = c.Step
		starts  = c.Starts
		ends    = c.Ends
		forward bool
	)
	if step == 0 {
		step = 1
	}
	forward = step > 0

	if starts == 0 {
		if forward {
			starts = rg.Starts.Column
		} else {
			starts = rg.Ends.Column
		}
	}
	if ends == 0 {
		if forward {
			ends = rg.Ends.Column
		} else {
			ends = rg.Starts.Column
		}
	}

	if forward {
		starts = max(starts, rg.Starts.Column)
		ends = min(ends, rg.Ends.Column)

		for i := starts; i <= ends; i += step {
			all = append(all, i-1)
		}
	} else {
		starts = min(starts, rg.Ends.Column)
		ends = max(ends, rg.Starts.Column)

		for i := starts; i >= ends; i += step {
			all = append(all, i-1)
		}
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
