package parse

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Selectable interface {
	Selection() (layout.Selection, error)
}

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
	KindInclude
	KindUse
)

type ExprKind interface {
	Kind() Kind
}

type Script struct {
	Body     []Expr
	Includes []Expr
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

func (s Script) Accept(v Visitor) error {
	return v.VisitScript(s)
}

type Macro struct {
	name string
	args []Expr
	body []Expr
}

func NewMacro(name string, args, body []Expr) Expr {
	return Macro{
		name: name,
		args: args,
		body: body,
	}
}

func (m Macro) String() string {
	return m.name
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

func (u UseRef) Accept(v Visitor) error {
	return v.VisitUseRef(u)
}

type IncludeFile struct {
	file  string
	alias string
}

func NewInclude(file, alias string) Expr {
	return IncludeFile{
		file:  file,
		alias: alias,
	}
}

func (i IncludeFile) File() string {
	return i.file
}

func (i IncludeFile) Alias() string {
	if i.alias != "" {
		return i.alias
	}
	alias := filepath.Base(i.file)
	for {
		ext := filepath.Ext(alias)
		if ext == "" {
			break
		}
		alias = strings.TrimSuffix(alias, ext)
	}
	return alias
}

func (i IncludeFile) String() string {
	return fmt.Sprintf("include(%s, alias: %s)", i.file, i.alias)
}

func (IncludeFile) Kind() Kind {
	return KindInclude
}

func (i IncludeFile) Accept(v Visitor) error {
	return v.VisitIncludeFile(i)
}

type ImportFile struct {
	file string

	format    string // using
	specifier string // with
	options   map[string]any

	alias       string // as
	defaultFile bool   // default
	readOnly    bool   // ro

	Position
}

func (i ImportFile) Options() map[string]any {
	if i.options == nil {
		return make(map[string]any)
	}
	return i.options
}

func (i ImportFile) Specifier() string {
	return i.specifier
}

func (i ImportFile) File() string {
	return i.file
}

func (i ImportFile) Format() string {
	return i.format
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

func (i ImportFile) Accept(v Visitor) error {
	return v.VisitImportFile(i)
}

type AssertType int8

const (
	AssertFail AssertType = iota
	AssertWarn
	AssertIgnore
	AssertUnknown
)

type Assert struct {
	expr Expr
	msg  string
	mode AssertType
}

func NewAssert(expr Expr, msg string, mode AssertType) Expr {
	return Assert{
		expr: expr,
		msg:  msg,
		mode: mode,
	}
}

func (a Assert) Type() AssertType {
	return a.mode
}

func (a Assert) Expr() Expr {
	return a.expr
}

func (a Assert) Failure() string {
	return a.msg
}

func (a Assert) String() string {
	return fmt.Sprintf("assert %s else %s", a.expr, a.msg)
}

func (a Assert) Accept(v Visitor) error {
	return v.VisitAssert(a)
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

func (p PrintRef) Accept(v Visitor) error {
	return v.VisitPrintRef(p)
}

type ExportRef struct {
	expr Expr

	file string

	format    string // using
	specifier string // with
	options   map[string]any

	Position
}

func (e ExportRef) Expr() Expr {
	return e.expr
}

func (e ExportRef) File() string {
	return e.file
}

func (e ExportRef) Format() string {
	return e.format
}

func (e ExportRef) String() string {
	return fmt.Sprintf("export %s", e.expr.String())
}

func (e ExportRef) Accept(v Visitor) error {
	return v.VisitExportRef(e)
}

type CellAccess struct {
	expr Expr
	addr Expr
	Position
}

func NewCellAccess(expr, addr Expr) Expr {
	return CellAccess{
		expr: expr,
		addr: addr,
	}
}

func (a CellAccess) Expr() Expr {
	return a.expr
}

func (a CellAccess) Addr() Expr {
	return a.addr
}

func (a CellAccess) String() string {
	return fmt.Sprintf("%s!%s", a.expr, a.addr)
}

func (CellAccess) KindOf() string {
	return "access"
}

func (a CellAccess) Accept(v Visitor) error {
	return v.VisitCellAccess(a)
}

type SpecialAccess struct {
	expr Expr
	prop Expr
	Position
}

func NewSpecial(expr, prop Expr) Expr {
	return SpecialAccess{
		expr: expr,
		prop: prop,
	}
}

func (a SpecialAccess) Object() Expr {
	return a.expr
}

func (a SpecialAccess) Property() Expr {
	return a.prop
}

func (a SpecialAccess) String() string {
	return fmt.Sprintf("%s@%s", a.expr.String(), a.prop)
}

func (SpecialAccess) KindOf() string {
	return "access"
}

func (a SpecialAccess) Accept(v Visitor) error {
	return v.VisitSpecial(a)
}

type Access struct {
	expr Expr
	prop Expr
	Position
}

func NewAccess(expr, prop Expr) Expr {
	return Access{
		expr: expr,
		prop: prop,
	}
}

func (a Access) Object() Expr {
	return a.expr
}

func (a Access) Property() Expr {
	return a.prop
}

func (a Access) String() string {
	return fmt.Sprintf("%s.%s", a.expr.String(), a.prop)
}

func (Access) KindOf() string {
	return "access"
}

func (a Access) Accept(v Visitor) error {
	return v.VisitAccess(a)
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

func (t Template) Accept(v Visitor) error {
	return v.VisitTemplate(t)
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

func (d Deferred) Accept(v Visitor) error {
	return v.VisitDeferred(d)
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

func (a Assignment) Accept(v Visitor) error {
	return v.VisitAssignment(a)
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

func (b Binary) Accept(v Visitor) error {
	return v.VisitBinary(b)
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

func (p Postfix) Accept(v Visitor) error {
	return v.VisitPostfix(p)
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

func (n Not) Accept(v Visitor) error {
	return v.VisitNot(n)
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

func (a And) Accept(v Visitor) error {
	return v.VisitAnd(a)
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

func (o Or) Accept(v Visitor) error {
	return v.VisitOr(o)
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

func (u Unary) Accept(v Visitor) error {
	return v.VisitUnary(u)
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

func (i Literal) Accept(v Visitor) error {
	return v.VisitLiteral(i)
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

func (n Number) Accept(v Visitor) error {
	return v.VisitNumber(n)
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

func (c Call) Accept(v Visitor) error {
	return v.VisitCall(c)
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

func (s Slice) Accept(v Visitor) error {
	return v.VisitSlice(s)
}

type IntervalList struct {
	items []Expr
	Position
}

func NewIntervalList(expr []Expr) Expr {
	return IntervalList{
		items: expr,
	}
}

func (i IntervalList) Count() int {
	return len(i.items)
}

func (i IntervalList) String() string {
	return fmt.Sprintf("internval(%v)", i.items)
}

func (i IntervalList) Selection() (layout.Selection, error) {
	all := make([]layout.Selection, 0, len(i.items))
	for _, r := range i.items {
		s, ok := r.(Selectable)
		if !ok {
			continue
		}
		e, err := s.Selection()
		if err != nil {
			return nil, err
		}
		all = append(all, e)
	}
	return layout.Combine(all...), nil
}

type IntervalExpr struct {
	from Expr
	to   Expr
	step Expr
	Position
}

func NewInterval(from, to, step Expr) Expr {
	return IntervalExpr{
		from: from,
		to:   to,
		step: step,
	}
}

func (e IntervalExpr) String() string {
	return fmt.Sprintf("interval(%v, %v)", e.from, e.to)
}

func (e IntervalExpr) Selection() (layout.Selection, error) {
	from, err := parseColumnExpr(e.from)
	if err != nil && e.from != nil {
		return nil, err
	}
	to, err := parseColumnExpr(e.to)
	if err != nil && e.to != nil {
		return nil, err
	}
	var step int64
	if e.step != nil {
		x, ok := e.step.(Number)
		if !ok {
			return nil, fmt.Errorf("step should be a number")
		}
		step = int64(x.Float())
	}
	return layout.SelectSpan(int64(from), int64(to), step), nil
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

func (i Identifier) Selection() (layout.Selection, error) {
	ix, err := parseColumnExpr(i)
	if err != nil {
		return nil, err
	}
	return layout.SelectSingle(int64(ix)), nil
}

func (i Identifier) Accept(v Visitor) error {
	return v.VisitIdentifier(i)
}

type ColumnAddr struct {
	layout.Position
	Absolute bool
}

func NewColumnAddr(pos layout.Position, abs bool) Expr {
	return ColumnAddr{
		Position: pos,
		Absolute: abs,
	}
}

func (a ColumnAddr) String() string {
	c := NewCellAddr(a.Position, false, false)
	return formatCellAddr(c.(CellAddr))
}

func (ColumnAddr) KindOf() string {
	return "address"
}

func (a ColumnAddr) CloneWithOffset(pos layout.Position) Expr {
	return a
}

func (a ColumnAddr) Accept(v Visitor) error {
	return v.VisitColumnAddr(a)
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

func (a CellAddr) Accept(v Visitor) error {
	return v.VisitCellAddr(a)
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

func (a RangeAddr) Accept(v Visitor) error {
	return v.VisitRangeAddr(a)
}
