package eval

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Expr interface {
	fmt.Stringer
}

type Clonable interface {
	CloneWithOffset(layout.Position) Expr
}

func CloneWithOffset(expr value.Formula, pos layout.Position) value.Formula {
	e, ok := expr.(*deferredFormula)
	if !ok {
		return expr
	}
	if c, ok := e.expr.(Clonable); ok {
		e.expr = c.CloneWithOffset(pos)
	}
	return e
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

type push struct {
	readOnly bool
}

func (push) String() string {
	return "<push>"
}

type pop struct{}

func (pop) String() string {
	return "<pop>"
}

type lockRef struct {
	ident string
}

func (k lockRef) String() string {
	return fmt.Sprintf("lock(%s)", k.ident)
}

type unlockRef struct {
	ident string
}

func (k unlockRef) String() string {
	return fmt.Sprintf("unlock(%s)", k.ident)
}

type useRef struct {
	ident    string
	readOnly bool
}

func (i useRef) String() string {
	return fmt.Sprintf("use(%s, ro: %t)", i.ident, i.readOnly)
}

type importFile struct {
	file string

	format    string
	specifier string
	options   map[string]string

	alias       string
	defaultFile bool
	readOnly    bool
}

func (i importFile) String() string {
	return fmt.Sprintf("import(%s, default: %t, ro: %t)", i.file, i.defaultFile, i.readOnly)
}

func (importFile) Kind() Kind {
	return KindImport
}

type printRef struct {
	expr    Expr
	pattern string
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

type macroDef struct {
	ident identifier
	args  []Expr
	body  Expr
}

func (d macroDef) String() string {
	return fmt.Sprintf("macro")
}

type access struct {
	expr Expr
	prop string
}

func (a access) String() string {
	return fmt.Sprintf("%s.%s", a.expr.String(), a.prop)
}

func (access) KindOf() string {
	return "access"
}

type template struct {
	expr []Expr
}

func (t template) String() string {
	return "<template>"
}

func (template) KindOf() string {
	return "template"
}

type deferred struct {
	expr Expr
}

func (d deferred) Type() string {
	return d.KindOf()
}

func (deferred) KindOf() string {
	return "deferred"
}

func (d deferred) String() string {
	return fmt.Sprintf("=%s", d.expr.String())
}

func (d deferred) Kind() value.ValueKind {
	return 0
}

type assignment struct {
	ident Expr
	expr  Expr
}

func (a assignment) String() string {
	return fmt.Sprintf("%s := %s", a.ident.String(), a.expr.String())
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

func (b binary) Predicate() value.Predicate {
	f := deferredFormula{
		expr: b,
	}
	return types.NewExprPredicate(&f)
}

type postfix struct {
	expr Expr
	op   op.Op
}

func (p postfix) String() string {
	oper := op.Symbol(p.op)
	return fmt.Sprintf("%s%s", p.expr.String(), oper)
}

func (p postfix) CloneWithOffset(pos layout.Position) Expr {
	expr := p.expr
	if c, ok := p.expr.(Clonable); ok {
		expr = c.CloneWithOffset(pos)
	}
	x := postfix{
		expr: expr,
		op:   p.op,
	}
	return x
}

type not struct {
	expr Expr
}

func (n not) String() string {
	return fmt.Sprintf("not(%s)", n.expr)
}

func (n not) Predicate() value.Predicate {
	e := unary{
		expr: n.expr,
		op:   op.Not,
	}
	f := deferredFormula{
		expr: e,
	}
	return types.NewExprPredicate(&f)
}

type and struct {
	left  Expr
	right Expr
}

func (a and) String() string {
	return fmt.Sprintf("and(%s, %s)", a.left, a.right)
}

func (a and) Predicate() value.Predicate {
	b := binary{
		op:    op.And,
		left:  a.left,
		right: a.right,
	}
	f := deferredFormula{
		expr: b,
	}
	return types.NewExprPredicate(&f)
}

type or struct {
	left  Expr
	right Expr
}

func (o or) String() string {
	return fmt.Sprintf("or(%s, %s)", o.left, o.right)
}

func (o or) Predicate() value.Predicate {
	b := binary{
		op:    op.Or,
		left:  o.left,
		right: o.right,
	}
	f := deferredFormula{
		expr: b,
	}
	return types.NewExprPredicate(&f)
}

type spread struct {
	expr Expr
}

func (s spread) String() string {
	return fmt.Sprintf("...%s", s.expr)
}

type unary struct {
	expr Expr
	op   op.Op
}

func (u unary) String() string {
	oper := op.Symbol(u.op)
	return fmt.Sprintf("%s%s", oper, u.expr.String())
}

func (u unary) CloneWithOffset(pos layout.Position) Expr {
	expr := u.expr
	if c, ok := u.expr.(Clonable); ok {
		expr = c.CloneWithOffset(pos)
	}
	x := unary{
		expr: expr,
		op:   u.op,
	}
	return x
}

type literal struct {
	value string
}

func (i literal) String() string {
	return fmt.Sprintf("\"%s\"", i.value)
}

func (literal) KindOf() string {
	return "primitive"
}

type number struct {
	value float64
}

func (n number) String() string {
	return strconv.FormatFloat(n.value, 'f', -1, 64)
}

func (number) KindOf() string {
	return "primitive"
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

type clear struct {
	name string
}

func (c clear) String() string {
	return fmt.Sprintf("clear(%s)", c.name)
}

type slice struct {
	view Expr
	expr Expr
}

func (s slice) String() string {
	return fmt.Sprintf("slice(%s, %s)", s.view, s.expr)
}

func (slice) KindOf() string {
	return "slice"
}

type rangeSlice struct {
	startAddr cellAddr
	endAddr   cellAddr
}

func (s rangeSlice) String() string {
	return fmt.Sprintf("range(%s, %s)", s.startAddr, s.endAddr)
}

func (s rangeSlice) Range() *layout.Range {
	rg := layout.NewRange(s.startAddr.Position, s.endAddr.Position)
	return rg
}

type columnsSlice struct {
	columns []columnsRange
}

func (s columnsSlice) String() string {
	return fmt.Sprintf("columns(%v)", s.columns)
}

func (s columnsSlice) Selection() layout.Selection {
	all := make([]layout.Selection, 0, len(s.columns))
	for _, r := range s.columns {
		all = append(all, r.Selection())
	}
	return layout.Combine(all...)
}

type columnsRange struct {
	from int
	to   int
	step int
}

func (c columnsRange) Selection() layout.Selection {
	if c.from == c.to {
		return layout.SelectSingle(int64(c.from))
	}
	return layout.SelectSpan(int64(c.from), int64(c.to), int64(c.step))
}

type filterSlice struct {
	expr Expr
}

func (s filterSlice) String() string {
	return fmt.Sprintf("filter(%s)", s.expr)
}

func (s filterSlice) Predicate() value.Predicate {
	f := deferredFormula{
		expr: s.expr,
	}
	return types.NewExprPredicate(&f)
}

type exprRange struct {
	from Expr
	to   Expr
	step Expr
}

func (e exprRange) String() string {
	return fmt.Sprintf("range(%v, %v)", e.from, e.to)
}

type identifier struct {
	name string
}

func (i identifier) String() string {
	return i.name
}

func (identifier) KindOf() string {
	return "identifier"
}

type qualifiedCellAddr struct {
	path Expr
	addr Expr
}

func (a qualifiedCellAddr) String() string {
	return fmt.Sprintf("qualified(%s.%s)", a.path.String(), a.addr.String())
}

func (qualifiedCellAddr) KindOf() string {
	return "qualified-address"
}

type cellAddr struct {
	layout.Position
	AbsCols bool
	AbsLine bool
}

func (a cellAddr) String() string {
	return formatCellAddr(a)
}

func (cellAddr) KindOf() string {
	return "address"
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

func (rangeAddr) KindOf() string {
	return "range"
}

func (a rangeAddr) CloneWithOffset(pos layout.Position) Expr {
	x := rangeAddr{
		startAddr: a.startAddr.CloneWithOffset(pos).(cellAddr),
		endAddr:   a.endAddr.CloneWithOffset(pos).(cellAddr),
	}
	return x
}

func (a rangeAddr) Range() *layout.Range {
	rg := layout.NewRange(a.startAddr.Position, a.endAddr.Position)
	return rg
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

type deferredFormula struct {
	expr Expr
}

func (*deferredFormula) Type() string {
	return "formula"
}

func (*deferredFormula) Kind() value.ValueKind {
	return value.KindFunction
}

func (f *deferredFormula) String() string {
	return f.expr.String()
}

func (f *deferredFormula) Eval(ctx value.Context) (value.Value, error) {
	return Eval(f.expr, ctx)
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
	case template:
		io.WriteString(w, "template(")
		for i := range e.expr {
			if i > 0 {
				io.WriteString(w, ", ")
			}
			dumpExpr(w, e.expr[i])
		}
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
		dumpExpr(w, e.expr)
		io.WriteString(w, ", ")
		io.WriteString(w, op.Symbol(e.op))
		io.WriteString(w, ")")
	case spread:
		io.WriteString(w, "spread(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case not:
		io.WriteString(w, "not(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case and:
		io.WriteString(w, "and(")
		dumpExpr(w, e.left)
		io.WriteString(w, ", ")
		dumpExpr(w, e.right)
		io.WriteString(w, ")")
	case or:
		io.WriteString(w, "or(")
		dumpExpr(w, e.left)
		io.WriteString(w, ", ")
		dumpExpr(w, e.right)
		io.WriteString(w, ")")
	case postfix:
		io.WriteString(w, "postfix(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ", ")
		io.WriteString(w, op.Symbol(e.op))
		io.WriteString(w, ")")
	case access:
		io.WriteString(w, "access(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ", ")
		io.WriteString(w, e.prop)
		io.WriteString(w, ")")
	case deferred:
		io.WriteString(w, "deferred(")
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
	case assignment:
		io.WriteString(w, "assignment(")
		dumpExpr(w, e.ident)
		io.WriteString(w, ", ")
		dumpExpr(w, e.expr)
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
	case qualifiedCellAddr:
		io.WriteString(w, "qualified(")
		dumpExpr(w, e.path)
		io.WriteString(w, ", ")
		dumpExpr(w, e.addr)
		io.WriteString(w, ")")
	case slice:
		io.WriteString(w, "slice(")
		dumpExpr(w, e.view)
		io.WriteString(w, ", ")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case filterSlice:
		dumpExpr(w, e.expr)
	case columnsSlice:
		io.WriteString(w, "selection(")
		for i := range e.columns {
			if i > 0 {
				io.WriteString(w, ",")
			}
			var fix, tix string
			if e.columns[i].from != 0 {
				fix = strconv.Itoa(e.columns[i].from)
			}
			if e.columns[i].to != 0 {
				tix = strconv.Itoa(e.columns[i].to)
			}
			io.WriteString(w, fix)
			if fix != tix {
				io.WriteString(w, ":")
				io.WriteString(w, tix)
			}
		}
		io.WriteString(w, ")")
	case importFile:
		io.WriteString(w, "import(")
		io.WriteString(w, e.file)
		if e.format != "" {
			io.WriteString(w, ", format: ")
			io.WriteString(w, e.format)
		}
		if e.alias != "" {
			io.WriteString(w, ", alias: ")
			io.WriteString(w, e.alias)
		}
		if e.defaultFile {
			io.WriteString(w, ", default")
		}
		if e.readOnly {
			io.WriteString(w, ", readonly")
		}
		io.WriteString(w, ")")
	case useRef:
		io.WriteString(w, "use(")
		io.WriteString(w, e.ident)
		io.WriteString(w, ")")
	case printRef:
		io.WriteString(w, "print(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case exportRef:
	case saveRef:
		io.WriteString(w, "save(")
		io.WriteString(w, ")")
	case push:
		io.WriteString(w, "push()")
	case pop:
		io.WriteString(w, "pop()")
	case clear:
		io.WriteString(w, "clear(")
		io.WriteString(w, e.name)
		io.WriteString(w, ")")
	default:
		io.WriteString(w, fmt.Sprintf("unknown(%T)", e))
	}
}
