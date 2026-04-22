package repr

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/internal/ds"
)

const Version1 = "1.0"

const (
	TypeScript = "script"
	TypeStmt   = "statement"
	TypeExpr   = "expression"
	TypeValue  = "primitive"
)

type Param struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type Envelop struct {
	Version string
	Mode    string
	Meta    []Param
	Root    *Node
}

type Node struct {
	Id       string  `json:"id"`
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Value    any     `json:"value,omitempty"`
	Params   []Param `json:"params,omitempty"`
	Children []*Node `json:"nodes,omitempty"`

	expr parse.Expr
}

func (n Node) Raw() string {
	return n.expr.String()
}

func InspectFile(file string) (*Envelop, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return Inspect(r)
}

func Inspect(r io.Reader) (*Envelop, error) {
	scan, err := parse.ScanScript(r)
	if err != nil {
		return nil, err
	}
	ps, err := parse.NewParser(scan)
	if err != nil {
		return nil, err
	}
	var meta []Param
	if list, err := ps.ExtractConfigEntries(); err == nil {
		for _, e := range list {
			p := Param{
				Name:  strings.Join(e.Path, "."),
				Value: e.Value,
			}
			meta = append(meta, p)
		}
	}
	expr, err := ps.Parse()
	if err != nil {
		return nil, err
	}
	v := astVisitor{
		stack:    ds.NewStack[*Node](),
		counters: make(map[string]int),
	}
	if err := v.visitExpr(expr); err != nil {
		return nil, err
	}
	node, _ := v.stack.Peek()
	node.Params = meta

	e := Envelop{
		Version: Version1,
		Mode:    string(ps.Mode()),
		Root:    node,
	}
	return &e, err
}

type astVisitor struct {
	stack    *ds.Stack[*Node]
	counters map[string]int
}

func (v astVisitor) VisitScript(expr parse.Script) error {
	node := Node{
		Type: TypeScript,
		expr: expr,
	}
	v.stack.Push(&node)
	for _, e := range expr.Body {
		a, ok := e.(parse.VisitableExpr)
		if !ok {
			return fmt.Errorf("expression can not be inspected")
		}
		if err := a.Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v astVisitor) VisitIncludeFile(expr parse.IncludeFile) error {
	node := v.newStmt("include", expr)
	node.Params = []Param{
		{Name: "file", Value: expr.File()},
		{Name: "alias", Value: expr.Alias()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitImportFile(expr parse.ImportFile) error {
	node := v.newStmt("import", expr)
	node.Params = []Param{
		{Name: "file", Value: expr.File()},
		{Name: "alias", Value: expr.Alias()},
		{Name: "format", Value: expr.Format()},
		{Name: "default", Value: expr.Default()},
		{Name: "ro", Value: expr.ReadOnly()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitExportRef(expr parse.ExportRef) error {
	node := v.newStmt("export", expr)
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitPrintRef(expr parse.PrintRef) error {
	node := v.newStmt("print", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()

	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitUseRef(expr parse.UseRef) error {
	node := v.newStmt("use", expr)
	node.Params = []Param{
		{Name: "identifier", Value: expr.Identifier()},
		{Name: "ro", Value: expr.ReadOnly()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitIdentifier(expr parse.Identifier) error {
	node := v.newValue("identifier", expr)
	node.Params = []Param{
		{Name: "name", Value: expr.Ident()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitLiteral(expr parse.Literal) error {
	node := v.newValue("literal", expr)
	node.Value = expr.Text()
	node.Params = []Param{
		{Name: "value", Value: expr.Text()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitNumber(expr parse.Number) error {
	node := v.newValue("number", expr)
	node.Value = expr.Float()
	node.Params = []Param{
		{Name: "value", Value: expr.Float()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitCellAddr(expr parse.CellAddr) error {
	node := v.newValue("address", expr)
	node.Value = expr.String()
	node.Params = []Param{
		{Name: "sheet", Value: expr.Sheet},
		{Name: "row", Value: expr.Line},
		{Name: "col", Value: expr.Column},
		{Name: "absRow", Value: expr.AbsRow},
		{Name: "absCol", Value: expr.AbsCol},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitRangeAddr(expr parse.RangeAddr) error {
	node := v.newValue("range", expr)
	node.Value = expr.String()
	v.stack.Push(node)
	if err := expr.StartAt().Accept(v); err != nil {
		return err
	}
	if err := expr.EndAt().Accept(v); err != nil {
		return err
	}
	v.stack.Pop()

	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitTemplate(expr parse.Template) error {
	node := v.newValue("template", expr)
	v.stack.Push(node)
	for _, e := range expr.Parts() {
		if err := v.visitExpr(e); err != nil {
			return err
		}
	}
	v.stack.Pop()

	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitAccess(expr parse.Access) error {
	node := v.newValue("access", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Object()); err != nil {
		return err
	}
	if err := v.visitExpr(expr.Property()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitCellAccess(expr parse.CellAccess) error {
	node := v.newValue("cell", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	if err := v.visitExpr(expr.Addr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitSpecial(expr parse.SpecialAccess) error {
	node := v.newValue("special", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Object()); err != nil {
		return err
	}
	if err := v.visitExpr(expr.Property()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitDeferred(expr parse.Deferred) error {
	node := v.newValue("formula", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitCall(expr parse.Call) error {
	node := v.newExpr("call", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Name()); err != nil {
		return err
	}
	for _, e := range expr.Args() {
		if err := v.visitExpr(e); err != nil {
			return err
		}
	}
	v.stack.Pop()

	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitSlice(expr parse.Slice) error {
	node := v.newExpr("slice", expr)
	v.stack.Push(node)
	if view := expr.View(); view != nil {
		if err := v.visitExpr(expr.View()); err != nil {
			return err
		}
	}
	if _, ok := expr.Expr().(parse.VisitableExpr); ok {
		v.visitExpr(expr.Expr())
	} else {
		switch e := expr.Expr().(type) {
		case parse.RangeAddr:
			if err := e.StartAt().Accept(v); err != nil {
				return err
			}
			if err := e.EndAt().Accept(v); err != nil {
				return err
			}
		case parse.IntervalList:
		case parse.IntervalExpr:
			//TODO
		default:
			//ignore
		}
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitAssert(expr parse.Assert) error {
	node := v.newExpr("assert", expr)
	node.Params = []Param{
		{Name: "message", Value: expr.Failure()},
		{Name: "mode", Value: "fail"},
	}
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitBinary(expr parse.Binary) error {
	node := v.newExpr("binary", expr)
	node.Params = []Param{
		{Name: "operator", Value: op.Symbol(expr.Op())},
	}
	v.stack.Push(node)
	if err := v.visitExpr(expr.Left()); err != nil {
		return err
	}
	if err := v.visitExpr(expr.Right()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitAssignment(expr parse.Assignment) error {
	node := v.newExpr("assignment", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Ident()); err != nil {
		return err
	}
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitPostfix(expr parse.Postfix) error {
	node := v.newExpr("postfix", expr)
	node.Params = []Param{
		{Name: "operator", Value: op.Symbol(expr.Op())},
	}
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitNot(expr parse.Not) error {
	node := v.newExpr("not", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitAnd(expr parse.And) error {
	node := v.newExpr("and", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Left()); err != nil {
		return err
	}
	if err := v.visitExpr(expr.Right()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitOr(expr parse.Or) error {
	node := v.newExpr("or", expr)
	v.stack.Push(node)
	if err := v.visitExpr(expr.Left()); err != nil {
		return err
	}
	if err := v.visitExpr(expr.Right()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitUnary(expr parse.Unary) error {
	node := v.newExpr("unary", expr)
	node.Params = []Param{
		{Name: "operator", Value: op.Symbol(expr.Op())},
	}
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) visitExpr(expr parse.Expr) error {
	if expr == nil {
		return nil
	}
	a, ok := expr.(parse.VisitableExpr)
	if !ok {
		return fmt.Errorf("expression can not be inspected")
	}
	return a.Accept(v)
}

func (v astVisitor) pushNode(node *Node) {
	top, _ := v.stack.Peek()
	top.Children = append(top.Children, node)
}

func (v astVisitor) newValue(name string, expr parse.Expr) *Node {
	return &Node{
		Id:   v.nextID(TypeValue),
		Type: TypeValue,
		Name: name,
		expr: expr,
	}
}

func (v astVisitor) newStmt(name string, expr parse.Expr) *Node {
	return &Node{
		Id:   v.nextID(TypeStmt),
		Type: TypeStmt,
		Name: name,
		expr: expr,
	}
}

func (v astVisitor) newExpr(name string, expr parse.Expr) *Node {
	return &Node{
		Id:   v.nextID(TypeExpr),
		Type: TypeExpr,
		Name: name,
		expr: expr,
	}
}

func (v astVisitor) nextID(kind string) string {
	id := v.counters[kind]
	id++
	v.counters[kind] = id
	return fmt.Sprintf("%s-%d", kind, id)
}
