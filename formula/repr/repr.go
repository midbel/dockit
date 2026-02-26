package repr

import (
	"fmt"
	"io"

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

type Node struct {
	Id       string  `json:"id"`
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Value    any     `json:"value,omitempty"`
	Params   []Param `json:"params,omitempty"`
	Children []*Node `json:"nodes,omitempty"`
}

func Inspect(r io.Reader) (any, error) {
	scan, err := parse.Scan(r, parse.ScanScript)
	if err != nil {
		return nil, err
	}
	ps, err := parse.NewParser(scan)
	if err != nil {
		return nil, err
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
	root := struct {
		Version string
		Root    *Node
	}{
		Version: Version1,
		Root:    node,
	}
	return root, err
}

type astVisitor struct {
	stack    *ds.Stack[*Node]
	counters map[string]int
}

func (v astVisitor) VisitScript(expr parse.Script) error {
	node := Node{
		Type: TypeScript,
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

func (v astVisitor) VisitImportFile(expr parse.ImportFile) error {
	node := v.newStmt("import")
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
	node := v.newStmt("export")
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitPrintRef(expr parse.PrintRef) error {
	node := v.newStmt("print")
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()

	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitUseRef(expr parse.UseRef) error {
	node := v.newStmt("use")
	node.Params = []Param{
		{Name: "identifier", Value: expr.Identifier()},
		{Name: "ro", Value: expr.ReadOnly()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitClear(expr parse.Clear) error {
	return nil
}

func (v astVisitor) VisitIdentifier(expr parse.Identifier) error {
	node := v.newValue("identifier")
	node.Params = []Param{
		{Name: "name", Value: expr.Ident()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitLiteral(expr parse.Literal) error {
	node := v.newValue("literal")
	node.Value = expr.Text()
	node.Params = []Param{
		{Name: "value", Value: expr.Text()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitNumber(expr parse.Number) error {
	node := v.newValue("number")
	node.Value = expr.Float()
	node.Params = []Param{
		{Name: "value", Value: expr.Float()},
	}
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitQualifiedCellAddr(expr parse.QualifiedCellAddr) error {
	node := v.newValue("address")
	v.stack.Push(node)
	if err := v.visitExpr(expr.Path()); err != nil {
		return err
	}
	if err := v.visitExpr(expr.Addr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitCellAddr(expr parse.CellAddr) error {
	node := v.newValue("address")
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
	node := v.newValue("range")
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
	node := v.newValue("template")
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
	node := v.newValue("access")
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitDeferred(expr parse.Deferred) error {
	node := v.newValue("formula")
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitCall(expr parse.Call) error {
	node := v.newExpr("call")
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
	node := v.newExpr("slice")
	v.stack.Push(node)
	if err := v.visitExpr(expr.View()); err != nil {
		return err
	}
	if _, ok := expr.Expr().(parse.VisitableExpr); ok {
		v.visitExpr(expr.Expr())
	} else {
		switch e := expr.Expr().(type) {
		case parse.RangeSlice:
			if err := e.StartAt().Accept(v); err != nil {
				return err
			}
			if err := e.EndAt().Accept(v); err != nil {
				return err
			}
		case parse.ColumnsSlice:
			//TODO
		default:
			//ignore
		}
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitBinary(expr parse.Binary) error {
	node := v.newExpr("binary")
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
	node := v.newExpr("assignment")
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
	node := v.newExpr("postfix")
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
	node := v.newExpr("not")
	v.stack.Push(node)
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	v.stack.Pop()
	v.pushNode(node)
	return nil
}

func (v astVisitor) VisitAnd(expr parse.And) error {
	node := v.newExpr("and")
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
	node := v.newExpr("or")
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
	node := v.newExpr("unary")
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
	a, ok := expr.(parse.VisitableExpr)
	if !ok {
		return fmt.Errorf("expression can not inspected")
	}
	return a.Accept(v)
}

func (v astVisitor) pushNode(node *Node) {
	top, _ := v.stack.Peek()
	top.Children = append(top.Children, node)
}

func (v astVisitor) newValue(name string) *Node {
	return &Node{
		Id:   v.nextID(TypeValue),
		Type: TypeValue,
		Name: name,
	}
}

func (v astVisitor) newStmt(name string) *Node {
	return &Node{
		Id:   v.nextID(TypeStmt),
		Type: TypeExpr,
		Name: name,
	}
}

func (v astVisitor) newExpr(name string) *Node {
	return &Node{
		Id:   v.nextID(TypeExpr),
		Type: TypeExpr,
		Name: name,
	}
}

func (v astVisitor) nextID(kind string) string {
	id := v.counters[kind]
	id++
	v.counters[kind] = id
	return fmt.Sprintf("%s-%d", kind, id)
}
