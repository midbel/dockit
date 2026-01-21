package formula

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/midbel/dockit/layout"
)

type Expr interface {
	fmt.Stringer
	CloneWithOffset(layout.Position) Expr
}

type Kind int8

const (
	KindStmt Kind = 1 << iota
	KindImport
	KindUse
)

type ExprKind interface {
	Kind() Kind
}

type useFile struct {
	file    Expr
	symbols []Expr
}

func (i useFile) String() string {
	return fmt.Sprintf("use(%s)", i.file.String())
}

func (i useFile) CloneWithOffset(_ layout.Position) Expr {
	return i
}

func (useFile) Kind() Kind {
	return KindUse
}

type importFile struct {
	file  Expr
	alias Expr
}

func (i importFile) String() string {
	return fmt.Sprintf("import(%s)", i.file.String())
}

func (i importFile) CloneWithOffset(_ layout.Position) Expr {
	return i
}

func (importFile) Kind() Kind {
	return KindImport
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

type chain struct {
	expr   Expr
	member Expr
}

func (c chain) String() string {
	return fmt.Sprintf("%s.%s", c.expr.String(), c.member.String())
}

func (c chain) CloneWithOffset(pos layout.Position) Expr {
	return c
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
	op := binaryOpString[b.op]
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
	op := unaryOpString[u.op]
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
		err    error
		offset int
		size   int
	)
	if addr == "" {
		return pos, fmt.Errorf("empty cell address")
	}
	if offset < len(addr) && addr[offset] == dollar {
		pos.AbsCols = true
		offset++
	}
	pos.Column, size = parseIndex(addr[offset:])
	if size == 0 {
		return pos, fmt.Errorf("invalid cell address - missing column")
	}
	offset += size
	if offset >= len(addr) {
		return pos, fmt.Errorf("invalid cell address - missing row")
	}

	if offset < len(addr) && addr[offset] == dollar {
		pos.AbsLine = true
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

var binaryOpString = map[rune]string{
	Add:     "+",
	Sub:     "-",
	Mul:     "*",
	Pow:     "^",
	Div:     "/",
	Percent: "%",
	Concat:  "&",
	Eq:      "=",
	Ne:      "<>",
	Lt:      "<",
	Le:      "<=",
	Gt:      ">",
	Ge:      ">=",
}

var unaryOpString = map[rune]string{
	Add: "+",
	Sub: "-",
}

func DumpExpr(expr Expr) string {
	var buf bytes.Buffer
	dumpExpr(&buf, expr)
	return buf.String()
}

func dumpExpr(w io.Writer, expr Expr) {
	switch e := expr.(type) {
	case identifier:
		io.WriteString(w, "identifier(")
		io.WriteString(w, e.name)
		io.WriteString(w, ")")
	case literal:
		io.WriteString(w, "literal(")
		io.WriteString(w, e.value)
		io.WriteString(w, ")")
	case number:
		io.WriteString(w, "number(")
		io.WriteString(w, strconv.FormatFloat(e.value, 'f', -1, 64))
		io.WriteString(w, ")")
	case binary:
		io.WriteString(w, "binary(")
		dumpExpr(w, e.left)
		io.WriteString(w, ", ")
		dumpExpr(w, e.right)
		io.WriteString(w, ", ")
		io.WriteString(w, binaryOpString[e.op])
		io.WriteString(w, ")")
	case unary:
		io.WriteString(w, "unary(")
		dumpExpr(w, e.right)
		io.WriteString(w, ", ")
		io.WriteString(w, unaryOpString[e.op])
		io.WriteString(w, ")")
	case call:
		io.WriteString(w, "call(")
		dumpExpr(w, e.ident)
		io.WriteString(w, ", args: ")
		for i := range e.args {
			if i > 0 {
				io.WriteString(w, ", ")
			}
			dumpExpr(w, e.args[i])
		}
		io.WriteString(w, ")")
	case cellAddr:
		io.WriteString(w, "cell(")
		io.WriteString(w, e.Position.String())
		io.WriteString(w, ", ")
		io.WriteString(w, strconv.FormatBool(e.AbsCols))
		io.WriteString(w, ", ")
		io.WriteString(w, strconv.FormatBool(e.AbsLine))
		io.WriteString(w, ")")
	case rangeAddr:
		io.WriteString(w, "range(")
		dumpExpr(w, e.startAddr)
		io.WriteString(w, ", ")
		dumpExpr(w, e.endAddr)
		io.WriteString(w, ")")
	default:
		io.WriteString(w, "unknown")
	}
}
