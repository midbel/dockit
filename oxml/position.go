package oxml

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Selection struct {
	Start int
	End   int
}

func ParseRange(str string) (Select, error) {
	var (
		sel   Select
		parts = strings.Split(str, ",")
	)
	for _, str := range parts {
		p, err := ParseSelection(strings.TrimSpace(str))
		if err != nil {
			return sel, err
		}
		sel.ranges = append(sel.ranges, p)
	}
	return sel, nil
}

func ParseSelection(str string) (Selection, error) {
	var (
		sel Selection
		err error
	)
	first, last, ok := strings.Cut(str, ":")
	if !ok {
		ix, err := parseOffset(str)
		if err == nil {
			sel.Start = ix
			sel.End = ix
		}
		return sel, err
	}
	sel.Start = -1
	sel.End = -1
	if str := strings.TrimSpace(first); str != "" {
		sel.Start, err = parseOffset(str)
		if err != nil {
			return sel, err
		}
	}
	if str := strings.TrimSpace(last); str != "" {
		sel.End, err = parseOffset(str)
		if err != nil {
			return sel, err
		}
	}
	return sel, err
}

func parseOffset(str string) (int, error) {
	var (
		column int
		offset int
	)
	for offset < len(str) && isLetter(rune(str[offset])) {
		delta := byte('A')
		if isLower(rune(str[offset])) {
			delta = 'a'
		}
		column = column*26 + int(str[offset]-delta+1)
		offset++
	}
	if offset < len(str) {
		return 0, fmt.Errorf("invalid column")
	}
	return column - 1, nil
}

func (s Selection) One() bool {
	return s.Start >= 0 && s.Start == s.End
}

func (s Selection) Full() bool {
	return s.Start < 0 && s.Start == s.End
}

func (s Selection) Open() bool {
	return s.End < 0 || s.Start < 0
}

func (s Selection) Select(rs *Row) []string {
	if s.One() {
		return s.selectOne(rs)
	}
	if s.Full() {
		return rs.Data()
	}
	if s.Open() {
		if s.End < 0 {
			s.End = len(rs.Cells) - 1
		}
		if s.Start < 0 {
			s.Start = 0
		}
	}
	var reverse bool
	if s.Start > s.End {
		reverse = true
		s.Start, s.End = s.End, s.Start
	}
	var list []string
	for i := s.Start; i <= s.End && i < len(rs.Cells); i++ {
		list = append(list, rs.Cells[i].Value())
	}
	if reverse {
		slices.Reverse(list)
	}
	return list
}

func (s Selection) selectOne(rs *Row) []string {
	if s.Start < 0 || s.Start >= len(rs.Cells) {
		return nil
	}
	return []string{rs.Cells[s.Start].Value()}
}

type Select struct {
	ranges []Selection
}

func (s Select) Select(rs *Row) []string {
	var list []string
	for _, r := range s.ranges {
		others := r.Select(rs)
		list = slices.Concat(list, others)
	}
	return list
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
	var (
		pos    Position
		offset int
	)
	for offset < len(addr) && isLetter(rune(addr[offset])) {
		delta := byte('A')
		if isLower(rune(addr[offset])) {
			delta = 'a'
		}
		pos.Column = pos.Column*26 + int64(addr[offset]-delta+1)
		offset++
	}
	if offset < len(addr) {
		pos.Line, _ = strconv.ParseInt(addr[offset:], 10, 64)
	}
	return pos
}

func (p Position) Equal(other Position) bool {
	return p.Line == other.Line && p.Column == other.Column
}

func (p Position) Addr() string {
	if p.Column == 0 {
		return ""
	}
	var (
		column = p.Column
		result string
	)
	for column > 0 {
		column--
		result = string(rune('A')+rune(column%26)) + result
		column /= 26
	}
	return fmt.Sprintf("%s%d", result, p.Line)
}
