package parse

import (
	"fmt"
	"strconv"
	"strings"
)

func formatCellAddr(addr CellAddr) string {
	if addr.Column == 0 {
		return ""
	}
	var (
		column = addr.Column
		result string
	)
	for column > 0 {
		column--
		result = string(rune('A')+rune(column%26)) + result
		column /= 26
	}
	var parts []string
	if addr.Sheet != "" {
		parts = append(parts, addr.Sheet)
		parts = append(parts, "!")
	}
	if addr.AbsCol {
		parts = append(parts, "$")
	}
	parts = append(parts, result)
	if addr.AbsRow {
		parts = append(parts, "$")
	}
	parts = append(parts, strconv.FormatInt(addr.Line, 10))
	return strings.Join(parts, "")
}

func parseCellAddr(addr string) (CellAddr, error) {
	var (
		pos    CellAddr
		err    error
		offset int
		size   int
	)
	if addr == "" {
		return pos, fmt.Errorf("empty cell address")
	}
	if offset < len(addr) && addr[offset] == dollar {
		pos.AbsCol = true
		offset++
	}
	pos.Column, size = parseIndex(addr[offset:])
	offset += size

	if offset < len(addr) && addr[offset] == dollar {
		pos.AbsRow = true
		offset++
	}
	if offset < len(addr) {
		pos.Line, err = strconv.ParseInt(addr[offset:], 10, 64)
		if err != nil {
			return pos, fmt.Errorf("invalid cell address - invalid row number")
		}
	}
	return pos, err
}

func parseIndex(str string) (int64, int) {
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

func parseColumnExpr(expr Expr) (int, error) {
	e, ok := expr.(Identifier)
	if !ok {
		return 0, fmt.Errorf("invalid column identifier")
	}
	ix, size := parseIndex(e.name)
	if size != len(e.name) {
		return 0, fmt.Errorf("invalid column index")
	}
	return int(ix), nil
}
