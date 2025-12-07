package oxml

import (
	"fmt"
	"strconv"
)

type Position struct {
	Line   int64
	Column int64
}

func parsePosition(addr string) Position {
	var (
		pos    Position
		offset int
	)
	for offset < len(addr) && isLetter(addr[offset]) {
		delta := byte('A')
		if isLower(addr[offset]) {
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
		result = string('A'+(column%26)) + result
		column /= 26
	}
	return fmt.Sprintf("%s%d", result, p.Line)
}

func isLetter(r byte) bool {
	return isUpper(r) || isLower(r)
}

func isLower(r byte) bool {
	return (r >= 'a' && r <= 'z')
}

func isUpper(r byte) bool {
	return (r >= 'A' && r <= 'Z')
}
