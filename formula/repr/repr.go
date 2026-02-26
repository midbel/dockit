package repr

import (
	"fmt"
	"io"

	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/internal/ds"
)

const (
	TypeScript = "script"
	TypeStmt   = "statement"
	TypeExpr   = "expression"
	TypeValue  = "primitive"
)

type Param struct {
	Name  string
	Value any
}

type Node struct {
	Id       string
	Type     string
	Name     string
	Params   []Param
	Children []*Node
}

func Inspect(r io.Reader) (*Node, error) {
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
		stack: ds.NewStack[*Node](),
	}
	if a, ok := expr.(parse.VisitableExpr); ok {
		err := a.Accept(v)
		if err != nil {
			return nil, err
		}
		root, _ := v.stack.Peek()
		return root, err
	}
	return nil, fmt.Errorf("expression can not be inspected!")
}

type astVisitor struct {
	stack *ds.Stack[*Node]
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
	node := Node{
		Type: TypeStmt,
		Name: "import",
		Params: []Param{
			{Name: "file", Value: expr.File()},
			{Name: "alias", Value: expr.Alias()},
			{Name: "format", Value: expr.Format()},
			{Name: "default", Value: expr.Default()},
			{Name: "ro", Value: expr.ReadOnly()},
		},
	}
	top, _ := v.stack.Peek()
	top.Children = append(top.Children, &node)
	return nil
}

func (v astVisitor) VisitExportRef(expr parse.ExportRef) error {
	return nil
}

func (v astVisitor) VisitPrintRef(expr parse.PrintRef) error {
	return nil
}

func (v astVisitor) VisitUseRef(expr parse.UseRef) error {
	return nil
}

func (v astVisitor) VisitClear(expr parse.Clear) error {
	return nil
}

func (v astVisitor) VisitIdentifier(expr parse.Identifier) error {
	return nil
}

func (v astVisitor) VisitLiteral(expr parse.Literal) error {
	return nil
}

func (v astVisitor) VisitNumber(expr parse.Number) error {
	return nil
}

func (v astVisitor) VisitQualifiedCellAddr(expr parse.QualifiedCellAddr) error {
	return nil
}

func (v astVisitor) VisitCellAddr(expr parse.CellAddr) error {
	return nil
}

func (v astVisitor) VisitRangeAddr(expr parse.RangeAddr) error {
	return nil
}

func (v astVisitor) VisitTemplate(expr parse.Template) error {
	return nil
}

func (v astVisitor) VisitAccess(expr parse.Access) error {
	return nil
}

func (v astVisitor) VisitDeferred(expr parse.Deferred) error {
	return nil
}

func (v astVisitor) VisitCall(expr parse.Call) error {
	return nil
}

func (v astVisitor) VisitSlice(expr parse.Slice) error {
	return nil
}

func (v astVisitor) VisitBinary(expr parse.Binary) error {
	return nil
}

func (v astVisitor) VisitAssignment(expr parse.Assignment) error {
	return nil
}

func (v astVisitor) VisitPostfix(expr parse.Postfix) error {
	return nil
}

func (v astVisitor) VisitNot(expr parse.Not) error {
	return nil
}

func (v astVisitor) VisitAnd(expr parse.And) error {
	return nil
}

func (v astVisitor) VisitOr(expr parse.Or) error {
	return nil
}

func (v astVisitor) VisitUnary(expr parse.Unary) error {
	return nil
}
