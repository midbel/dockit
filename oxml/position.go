package oxml

import (
	"fmt"
	"strconv"
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
