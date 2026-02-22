package parse

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
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
	Position
}

func NewScript(body []Expr) Expr {
	return Script{
		Body: body,
	}
}

func (Script) String() string {
	return "<script>"
}

type Push struct {
	readOnly bool
	Position
}

func (Push) String() string {
	return "<push>"
}

type Pop struct {
	Position
}

func (Pop) String() string {
	return "<pop>"
}

type LockRef struct {
	ident string
	Position
}

func (k LockRef) String() string {
	return fmt.Sprintf("lock(%s)", k.ident)
}

type UnlockRef struct {
	ident string
	Position
}

func (k UnlockRef) String() string {
	return fmt.Sprintf("unlock(%s)", k.ident)
}

type UseRef struct {
	ident    string
	readOnly bool
	Position
}

func (u UseRef) Identifier() string {
	return u.ident
}

func (u UseRef) ReadOnly() bool {
	return u.readOnly
}

func (u UseRef) String() string {
	return fmt.Sprintf("use(%s, ro: %t)", u.ident, u.readOnly)
}

type ImportFile struct {
	file string

	format    string // using
	specifier string // with
	options   map[string]string

	alias       string // as
	defaultFile bool   // default
	readOnly    bool   // ro

	Position
}

func (i ImportFile) File() string {
	return i.file
}

func (i ImportFile) Alias() string {
	return i.alias
}

func (i ImportFile) Default() bool {
	return i.defaultFile
}

func (i ImportFile) ReadOnly() bool {
	return i.readOnly
}

func (i ImportFile) String() string {
	return fmt.Sprintf("import(%s, default: %t, ro: %t)", i.file, i.defaultFile, i.readOnly)
}

func (ImportFile) Kind() Kind {
	return KindImport
}

type PrintRef struct {
	expr    Expr
	pattern string
	Position
}

func (p PrintRef) Expr() Expr {
	return p.expr
}

func (p PrintRef) Pattern() string {
	return p.pattern
}

func (p PrintRef) String() string {
	return fmt.Sprintf("print %s", p.expr.String())
}

type ExportRef struct {
	expr   Expr
	file   Expr
	format Expr
	Position
}

func (p ExportRef) String() string {
	return fmt.Sprintf("export %s", p.expr.String())
}

type SaveRef struct {
	expr Expr
	Position
}

func (p SaveRef) String() string {
	return fmt.Sprintf("save %s", p.expr.String())
}

type MacroDef struct {
	ident Expr
	args  []Expr
	body  Expr
	Position
}

func (d MacroDef) String() string {
	return fmt.Sprintf("macro")
}

type Access struct {
	expr Expr
	prop string
	Position
}

func NewAccess(expr Expr, prop string) Expr {
	return Access{
		expr: expr,
		prop: prop,
	}
}

func (a Access) Object() Expr {
	return a.expr
}

func (a Access) Property() string {
	return a.prop
}

func (a Access) String() string {
	return fmt.Sprintf("%s.%s", a.expr.String(), a.prop)
}

func (Access) KindOf() string {
	return "access"
}

type Template struct {
	expr []Expr
	Position
}

func NewTemplate(list []Expr) Expr {
	return Template{
		expr: list,
	}
}

func (t Template) Parts() []Expr {
	return t.expr
}

func (t Template) String() string {
	return "<template>"
}

func (Template) KindOf() string {
	return "template"
}

type Deferred struct {
	expr Expr
	Position
}

func NewDeferred(expr Expr) Expr {
	return Deferred{
		expr: expr,
	}
}

func (d Deferred) Expr() Expr {
	return d.expr
}

func (d Deferred) Type() string {
	return d.KindOf()
}

func (Deferred) KindOf() string {
	return "deferred"
}

func (d Deferred) String() string {
	return fmt.Sprintf("=%s", d.expr.String())
}

func (d Deferred) Kind() value.ValueKind {
	return 0
}

type Assignment struct {
	ident Expr
	expr  Expr
	Position
}

func NewAssignment(ident, expr Expr) Expr {
	return Assignment{
		ident: ident,
		expr:  expr,
	}
}

func (a Assignment) Ident() Expr {
	return a.ident
}

func (a Assignment) Expr() Expr {
	return a.expr
}

func (a Assignment) String() string {
	return fmt.Sprintf("%s := %s", a.ident.String(), a.expr.String())
}

type Binary struct {
	left  Expr
	right Expr
	op    op.Op
	Position
}

func NewBinary(left, right Expr, oper op.Op) Expr {
	return Binary{
		left:  left,
		right: right,
		op:    oper,
	}
}

func (b Binary) Left() Expr {
	return b.left
}

func (b Binary) Right() Expr {
	return b.right
}

func (b Binary) Op() op.Op {
	return b.op
}

func (b Binary) String() string {
	oper := op.Symbol(b.op)
	return fmt.Sprintf("%s %s %s", b.left.String(), oper, b.right.String())
}

func (b Binary) CloneWithOffset(pos layout.Position) Expr {
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
	x := Binary{
		left:  left,
		right: right,
		op:    b.op,
	}
	return x
}

type Postfix struct {
	expr Expr
	op   op.Op
	Position
}

func NewPostfix(expr Expr, oper op.Op) Expr {
	return Postfix{
		expr: expr,
		op:   oper,
	}
}

func (p Postfix) Expr() Expr {
	return p.expr
}

func (p Postfix) Op() op.Op {
	return p.op
}

func (p Postfix) String() string {
	oper := op.Symbol(p.op)
	return fmt.Sprintf("%s%s", p.expr.String(), oper)
}

func (p Postfix) CloneWithOffset(pos layout.Position) Expr {
	expr := p.expr
	if c, ok := p.expr.(Clonable); ok {
		expr = c.CloneWithOffset(pos)
	}
	x := Postfix{
		expr: expr,
		op:   p.op,
	}
	return x
}

type Not struct {
	expr Expr
	Position
}

func NewNot(expr Expr) Expr {
	return Not{
		expr: expr,
	}
}

func (n Not) Expr() Expr {
	return n.expr
}

func (n Not) String() string {
	return fmt.Sprintf("not(%s)", n.expr)
}

type And struct {
	left  Expr
	right Expr
	Position
}

func NewAnd(left, right Expr) Expr {
	return And{
		left:  left,
		right: right,
	}
}

func (a And) Left() Expr {
	return a.left
}

func (a And) Right() Expr {
	return a.right
}

func (a And) String() string {
	return fmt.Sprintf("and(%s, %s)", a.left, a.right)
}

type Or struct {
	left  Expr
	right Expr
	Position
}

func NewOr(left, right Expr) Expr {
	return Or{
		left:  left,
		right: right,
	}
}

func (o Or) Left() Expr {
	return o.left
}

func (o Or) Right() Expr {
	return o.right
}

func (o Or) String() string {
	return fmt.Sprintf("or(%s, %s)", o.left, o.right)
}

type Spread struct {
	expr Expr
	Position
}

func NewSpread(expr Expr) Expr {
	return Spread{
		expr: expr,
	}
}

func (s Spread) String() string {
	return fmt.Sprintf("...%s", s.expr)
}

type Unary struct {
	expr Expr
	op   op.Op
	Position
}

func NewUnary(expr Expr, oper op.Op) Expr {
	return Unary{
		expr: expr,
		op:   oper,
	}
}

func (u Unary) Expr() Expr {
	return u.expr
}

func (u Unary) Op() op.Op {
	return u.op
}

func (u Unary) String() string {
	oper := op.Symbol(u.op)
	return fmt.Sprintf("%s%s", oper, u.expr.String())
}

func (u Unary) CloneWithOffset(pos layout.Position) Expr {
	expr := u.expr
	if c, ok := u.expr.(Clonable); ok {
		expr = c.CloneWithOffset(pos)
	}
	x := Unary{
		expr: expr,
		op:   u.op,
	}
	return x
}

type Literal struct {
	value string
	Position
}

func NewLiteral(value string) Expr {
	return Literal{
		value: value,
	}
}

func (i Literal) Text() string {
	return i.value
}

func (i Literal) String() string {
	return fmt.Sprintf("\"%s\"", i.value)
}

func (Literal) KindOf() string {
	return "primitive"
}

type Number struct {
	value float64
	Position
}

func NewNumber(value float64) Expr {
	return Number{
		value: value,
	}
}

func (n Number) Float() float64 {
	return n.value
}

func (n Number) String() string {
	return strconv.FormatFloat(n.value, 'f', -1, 64)
}

func (Number) KindOf() string {
	return "primitive"
}

type Call struct {
	ident Expr
	args  []Expr
	Position
}

func NewCall(id Expr, args []Expr) Expr {
	return Call{
		ident: id,
		args:  args,
	}
}

func (c Call) Name() Expr {
	return c.ident
}

func (c Call) Args() []Expr {
	return c.args
}

func (c Call) String() string {
	var args []string
	for i := range c.args {
		args = append(args, c.args[i].String())
	}
	return fmt.Sprintf("%s(%s)", c.ident.String(), strings.Join(args, ", "))
}

func (c Call) CloneWithOffset(pos layout.Position) Expr {
	x := Call{
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

type Clear struct {
	name string
	Position
}

func (c Clear) String() string {
	return fmt.Sprintf("clear(%s)", c.name)
}

type Slice struct {
	view Expr
	expr Expr
	Position
}

func NewSlice(view, expr Expr) Expr {
	return Slice{
		view: view,
		expr: expr,
	}
}

func (s Slice) View() Expr {
	return s.view
}

func (s Slice) Expr() Expr {
	return s.expr
}

func (s Slice) String() string {
	return fmt.Sprintf("slice(%s, %s)", s.view, s.expr)
}

func (Slice) KindOf() string {
	return "slice"
}

type RangeSlice struct {
	startAddr CellAddr
	endAddr   CellAddr
	Position
}

func (s RangeSlice) StartAt() CellAddr {
	return s.startAddr
}

func (s RangeSlice) EndAt() CellAddr {
	return s.endAddr
}

func (s RangeSlice) String() string {
	return fmt.Sprintf("range(%s, %s)", s.startAddr, s.endAddr)
}

func (s RangeSlice) Range() *layout.Range {
	rg := layout.NewRange(s.startAddr.Position, s.endAddr.Position)
	return rg
}

type ColumnsSlice struct {
	columns []ColumnsRange
	Position
}

func NewColumnsSlice(cols []Expr) Expr {
	var columns []ColumnsRange
	for i := range cols {
		c := cols[i].(ColumnsRange)
		columns = append(columns, c)
	}
	return ColumnsSlice{
		columns: columns,
	}
}

func (s ColumnsSlice) Count() int {
	return len(s.columns)
}

func (s ColumnsSlice) String() string {
	return fmt.Sprintf("columns(%v)", s.columns)
}

func (s ColumnsSlice) Selection() layout.Selection {
	all := make([]layout.Selection, 0, len(s.columns))
	for _, r := range s.columns {
		all = append(all, r.Selection())
	}
	return layout.Combine(all...)
}

type ColumnsRange struct {
	from int
	to   int
	step int
}

func SelectRange(from, to, step int) Expr {
	return ColumnsRange{
		from: from,
		to:   to,
		step: step,
	}
}

func (c ColumnsRange) String() string {
	return fmt.Sprintf("columns(%d, %d, %d)", c.from, c.to, c.step)
}

func (c ColumnsRange) Selection() layout.Selection {
	if c.from == c.to {
		return layout.SelectSingle(int64(c.from))
	}
	return layout.SelectSpan(int64(c.from), int64(c.to), int64(c.step))
}

type ExprRange struct {
	from Expr
	to   Expr
	step Expr
	Position
}

func (e ExprRange) String() string {
	return fmt.Sprintf("range(%v, %v)", e.from, e.to)
}

type Identifier struct {
	name string
	Position
}

func NewIdentifier(id string) Expr {
	return Identifier{
		name: id,
	}
}

func (i Identifier) Ident() string {
	return i.name
}

func (i Identifier) String() string {
	return i.name
}

func (Identifier) KindOf() string {
	return "identifier"
}

type QualifiedCellAddr struct {
	path Expr
	addr Expr
	Position
}

func NewQualifiedAddr(path, addr Expr) Expr {
	return QualifiedCellAddr{
		path: path,
		addr: addr,
	}
}

func (a QualifiedCellAddr) Path() Expr {
	return a.path
}

func (a QualifiedCellAddr) Addr() Expr {
	return a.addr
}

func (a QualifiedCellAddr) String() string {
	return fmt.Sprintf("qualified(%s.%s)", a.path.String(), a.addr.String())
}

func (QualifiedCellAddr) KindOf() string {
	return "qualified-address"
}

type CellAddr struct {
	layout.Position
	AbsCol bool
	AbsRow bool
}

func NewCellAddr(pos layout.Position, col, row bool) Expr {
	return CellAddr{
		Position: pos,
		AbsCol:   col,
		AbsRow:   row,
	}
}

func (a CellAddr) String() string {
	return formatCellAddr(a)
}

func (CellAddr) KindOf() string {
	return "address"
}

func (a CellAddr) CloneWithOffset(pos layout.Position) Expr {
	x := a
	if !x.AbsRow {
		x.Line += pos.Line
	}
	if !x.AbsCol {
		x.Column += pos.Column
	}
	return x
}

type RangeAddr struct {
	startAddr CellAddr
	endAddr   CellAddr
	Position
}

func NewRangeAddr(start, end Expr) Expr {
	return RangeAddr{
		startAddr: start.(CellAddr),
		endAddr:   end.(CellAddr),
	}
}

func (a RangeAddr) StartAt() CellAddr {
	return a.startAddr
}

func (a RangeAddr) EndAt() CellAddr {
	return a.endAddr
}

func (a RangeAddr) String() string {
	return fmt.Sprintf("%s:%s", a.startAddr.String(), a.endAddr.String())
}

func (RangeAddr) KindOf() string {
	return "range"
}

func (a RangeAddr) CloneWithOffset(pos layout.Position) Expr {
	x := RangeAddr{
		startAddr: a.startAddr.CloneWithOffset(pos).(CellAddr),
		endAddr:   a.endAddr.CloneWithOffset(pos).(CellAddr),
	}
	return x
}

func (a RangeAddr) Range() *layout.Range {
	rg := layout.NewRange(a.startAddr.Position, a.endAddr.Position)
	return rg
}

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
	if size == 0 {
		return pos, fmt.Errorf("invalid cell address - missing column")
	}
	offset += size
	if offset >= len(addr) {
		return pos, fmt.Errorf("invalid cell address - missing row")
	}

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
