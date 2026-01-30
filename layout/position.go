package layout

import (
	"strconv"
	"strings"
)

type Position struct {
	Sheet  string
	Line   int64
	Column int64
}

func ParsePosition(addr string) Position {
	var (
		pos    Position
		offset int
	)
	pos.Column, offset = ParseIndex(addr)
	pos.Line, _ = strconv.ParseInt(addr[offset:], 10, 64)
	return pos
}

func (p Position) Equal(other Position) bool {
	return p.Line == other.Line && p.Column == other.Column
}

func (p Position) Addr() string {
	var parts []string
	if p.Sheet != "" {
		parts = append(parts, p.Sheet)
		parts = append(parts, "!")
	}
	parts = append(parts, indexToString(p.Column))
	parts = append(parts, strconv.FormatInt(p.Line, 10))
	return strings.Join(parts, "")
}

func (p Position) String() string {
	return p.Addr()
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

func IsAddress(addr string) bool {
	size := len(addr)
	if size < 2 {
		return false
	}
	var offset int
	for offset < size {
		c := addr[offset]
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		if c < 'A' || c > 'Z' {
			break
		}
		offset++
	}
	if offset == 0 || offset >= size || addr[offset] == '0' {
		return false
	}
	for offset < size {
		c := addr[offset]
		if c < '0' || c > '9' {
			return false
		}
		offset++
	}
	return offset == size
}

func ParseIndex(str string) (int64, int) {
	if len(str) == 0 {
		return 0, 0
	}
	var (
		offset int
		index  int
	)
	for offset < len(str) && isLetter(rune(str[offset])) {
		delta := byte('A')
		if isLower(rune(str[offset])) {
			delta = 'a'
		}
		index = index*26 + int(str[offset]-delta+1)
		offset++
	}
	return int64(index), offset
}

func indexToString(ix int64) string {
	var result string
	for ix > 0 {
		ix--
		result = string(rune('A')+rune(ix%26)) + result
		ix /= 26
	}
	return result
}

func isLower(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func isLetter(c rune) bool {
	return isLower(c) || isUpper(c)
}
