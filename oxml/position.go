package oxml

import (
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

func isLetter(r byte) bool {
	return isUpper(r) || isLower(r)
}

func isLower(r byte) bool {
	return (r >= 'a' && r <= 'z')
}

func isUpper(r byte) bool {
	return (r >= 'A' && r <= 'Z')
}
