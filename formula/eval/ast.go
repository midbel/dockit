package eval

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/layout"
)

type Expr interface {
	fmt.Stringer
}

type Clonable interface {
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

type Script struct {
	Body []Expr
}

func (Script) String() string {
	return "<script>"
}

type useFile struct {
	file  string
	alias string
}

func (i useFile) String() string {
	return fmt.Sprintf("use(%s)", i.file)
}

func (useFile) Kind() Kind {
	return KindUse
}

type importFile struct {
	file  string
	alias string
}

func (i importFile) String() string {
	return fmt.Sprintf("import(%s)", i.file)
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

type exportRef struct {
	expr   Expr
	file   Expr
	format Expr
}

func (p exportRef) String() string {
	return fmt.Sprintf("export %s", p.expr.String())
}

type saveRef struct {
	expr Expr
}

func (p saveRef) String() string {
	return fmt.Sprintf("save %s", p.expr.String())
}

type defaultRef struct {
	expr Expr
}

func (d defaultRef) String() string {
	return fmt.Sprintf("default %s", d.expr.String())
}

type macroDef struct {
	ident identifier
	args  []Expr
	body  Expr
}

func (d macroDef) String() string {
	return fmt.Sprintf("macro")
}

type access struct {
	expr   Expr
	member Expr
}

func (a access) String() string {
	return fmt.Sprintf("%s.%s", a.expr.String(), a.member.String())
}

type lambda struct {
	expr Expr
}

func (f lambda) String() string {
	return fmt.Sprintf("=%s", f.expr.String())
}

type assignment struct {
	ident identifier
	expr  Expr
}

func (a assignment) String() string {
	return fmt.Sprintf("%s := %s", a.ident.String(), a.expr.String())
}

type pivotExpr struct {
	body []Expr
}

func makePivotExpr(body []Expr) Expr {
	return pivotExpr{
		body: body,
	}
}

func (e pivotExpr) String() string {
	return "<pivot>"
}

type chartExpr struct {
	body []Expr
}

func makeChartExpr(body []Expr) Expr {
	return chartExpr{
		body: body,
	}
}

func (e chartExpr) String() string {
	return "<chart>"
}

type sheetExpr struct {
	body []Expr
}

func makeSheetExpr(body []Expr) Expr {
	return sheetExpr{
		body: body,
	}
}

func (e sheetExpr) String() string {
	return "<sheet>"
}

type filterExpr struct {
	body []Expr
}

func makeFilterExpr(body []Expr) Expr {
	return filterExpr{
		body: body,
	}
}

func (e filterExpr) String() string {
	return "<filter>"
}

type binary struct {
	left  Expr
	right Expr
	op    op.Op
}

func (b binary) String() string {
	oper := op.Symbol(b.op)
	return fmt.Sprintf("%s %s %s", b.left.String(), oper, b.right.String())
}

func (b binary) CloneWithOffset(pos layout.Position) Expr {
	var (
		left  = b.left
		right = b.right
	)
	if c, ok := b.left.(Clonable); ok {
		left = c.CloneWithOffset(pos)
	}
	if c, ok := b.right.(Clonable); ok {
		right = c.CloneWithOffset(pos)
	}
	x := binary{
		left:  left,
		right: right,
		op:    b.op,
	}
	return x
}

type unary struct {
	right Expr
	op    op.Op
}

func (u unary) String() string {
	oper := op.Symbol(u.op)
	return fmt.Sprintf("%s%s", oper, u.right.String())
}

func (u unary) CloneWithOffset(pos layout.Position) Expr {
	right := u.right
	if c, ok := u.right.(Clonable); ok {
		right = c.CloneWithOffset(pos)
	}
	x := unary{
		right: right,
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

type number struct {
	value float64
}

func (n number) String() string {
	return strconv.FormatFloat(n.value, 'f', -1, 64)
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
		a := c.args[i]
		if c, ok := a.(Clonable); ok {
			a = c.CloneWithOffset(pos)
		}
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
		io.WriteString(w, op.Symbol(e.op))
		io.WriteString(w, ")")
	case unary:
		io.WriteString(w, "unary(")
		dumpExpr(w, e.right)
		io.WriteString(w, ", ")
		io.WriteString(w, op.Symbol(e.op))
		io.WriteString(w, ")")
	case access:
		io.WriteString(w, "access(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ", ")
		dumpExpr(w, e.member)
		io.WriteString(w, ")")
	case lambda:
		io.WriteString(w, "formula(")
		dumpExpr(w, e.expr)
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
	case importFile:
	case useFile:
	case printRef:
	case exportRef:
	case defaultRef:
	case saveRef:
	case pivotExpr:
	case sheetExpr:
	case chartExpr:
	case filterExpr:
	default:
		io.WriteString(w, "unknown")
	}
}
