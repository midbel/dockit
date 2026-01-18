package formula

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/midbel/dockit/layout"
)

type Expr interface {
	fmt.Stringer
	CloneWithOffset(layout.Position) Expr
}

type importFile struct {
	file  string
	sheet string
}

func (i importFile) String() string {
	return fmt.Sprintf("import(%s:%s)", i.file, i.sheet)
}

func (i importFile) CloneWithOffset(_ layout.Position) Expr {
	return i
}

type printRef struct {
	expr Expr
}

func (p printRef) String() string {
	return fmt.Sprintf("print %s", p.expr.String())
}

func (p printRef) CloneWithOffset(_ layout.Position) Expr {
	return p
}

type exportRef struct {
	expr   Expr
	file   Expr
	Format Expr
}

func (p exportRef) String() string {
	return fmt.Sprintf("export %s", p.expr.String())
}

func (p exportRef) CloneWithOffset(_ layout.Position) Expr {
	return p
}

type saveRef struct {
	expr Expr
}

func (p saveRef) String() string {
	return fmt.Sprintf("save %s", p.expr.String())
}

func (p saveRef) CloneWithOffset(_ layout.Position) Expr {
	return p
}

type macroDef struct {
	ident identifier
	args  []Expr
	body  Expr
}

func (d macroDef) String() string {
	return fmt.Sprintf("macro")
}

func (d macroDef) CloneWithOffset(pos layout.Position) Expr {
	return d
}

type assignment struct {
	ident identifier
	expr  Expr
}

func (a assignment) String() string {
	return fmt.Sprintf("%s := %s", a.ident.String(), a.expr.String())
}

func (a assignment) CloneWithOffset(pos layout.Position) Expr {
	return a
}

type binary struct {
	left  Expr
	right Expr
	op    rune
}

func (b binary) String() string {
	var op string
	switch b.op {
	case Add:
		op = "+"
	case Sub:
		op = "-"
	case Mul:
		op = "*"
	case Div:
		op = "/"
	case Pow:
		op = "^"
	case Concat:
		op = "&"
	case Eq:
		op = "="
	case Ne:
		op = "<>"
	case Lt:
		op = "<"
	case Le:
		op = "<="
	case Gt:
		op = ">"
	case Ge:
		op = ">="
	}
	return fmt.Sprintf("%s %s %s", b.left.String(), op, b.right.String())
}

func (b binary) CloneWithOffset(pos layout.Position) Expr {
	x := binary{
		left:  b.left.CloneWithOffset(pos),
		right: b.right.CloneWithOffset(pos),
		op:    b.op,
	}
	return x
}

type unary struct {
	right Expr
	op    rune
}

func (u unary) String() string {
	var op string
	switch u.op {
	case Add:
		op = "+"
	case Sub:
		op = "-"
	}
	return fmt.Sprintf("%s%s", op, u.right.String())
}

func (u unary) CloneWithOffset(pos layout.Position) Expr {
	x := unary{
		right: u.right.CloneWithOffset(pos),
		op:    u.op,
	}
	return x
}

type literal struct {
	value string
}

func (i literal) String() string {
	return fmt.Sprintf("\"%s\"", i.value)
}

func (i literal) CloneWithOffset(_ layout.Position) Expr {
	return i
}

type number struct {
	value float64
}

func (n number) String() string {
	return strconv.FormatFloat(n.value, 'f', -1, 64)
}

func (n number) CloneWithOffset(_ layout.Position) Expr {
	return n
}

type call struct {
	ident Expr
	args  []Expr
}

func (c call) String() string {
	var args []string
	for i := range c.args {
		args = append(args, c.args[i].String())
	}
	return fmt.Sprintf("%s(%s)", c.ident.String(), strings.Join(args, ", "))
}

func (c call) CloneWithOffset(pos layout.Position) Expr {
	x := call{
		ident: c.ident,
	}
	for i := range c.args {
		a := c.args[i].CloneWithOffset(pos)
		x.args = append(x.args, a)
	}
	return x
}

type identifier struct {
	name string
}

func (i identifier) String() string {
	return i.name
}

func (i identifier) CloneWithOffset(_ layout.Position) Expr {
	return i
}

type cellAddr struct {
	layout.Position
	AbsCols bool
	AbsLine bool
}

func (a cellAddr) String() string {
	return formatCellAddr(a)
}

func (a cellAddr) CloneWithOffset(pos layout.Position) Expr {
	x := a
	if !x.AbsLine {
		x.Line += pos.Line
	}
	if !x.AbsCols {
		x.Column += pos.Column
	}
	return x
}

type rangeAddr struct {
	startAddr cellAddr
	endAddr   cellAddr
}

func (a rangeAddr) String() string {
	return fmt.Sprintf("%s:%s", a.startAddr.String(), a.endAddr.String())
}

func (a rangeAddr) CloneWithOffset(pos layout.Position) Expr {
	x := rangeAddr{
		startAddr: a.startAddr.CloneWithOffset(pos).(cellAddr),
		endAddr:   a.endAddr.CloneWithOffset(pos).(cellAddr),
	}
	return x
}

func formatCellAddr(addr cellAddr) string {
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
	if addr.AbsCols {
		parts = append(parts, "$")
	}
	parts = append(parts, result)
	if addr.AbsLine {
		parts = append(parts, "$")
	}
	parts = append(parts, strconv.FormatInt(addr.Line, 10))
	return strings.Join(parts, "")
}

func parseCellAddr(addr string) (cellAddr, error) {
	var (
		pos    cellAddr
		offset int
		size   int
	)
	if offset < len(addr) && addr[offset] == dollar {
		pos.AbsCols = true
		offset++
	}
	pos.Column, size = parseIndex(addr[offset:])
	offset += size

	if offset < len(addr) && addr[offset] == dollar {
		pos.AbsLine = true
		offset++
	}
	if offset < len(addr) {
		pos.Line, _ = strconv.ParseInt(addr[offset:], 10, 64)
	}
	return pos, nil
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
