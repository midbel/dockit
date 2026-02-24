package eval

import (
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/internal/ds"
	"github.com/midbel/dockit/value"
)

type evalVisitor struct {
	ctx    *EngineContext
	values *ds.Stack[value.Value]
	phase  scriptPhase
}

func evalScript(ctx *EngineContext) *evalVisitor {
	return &evalVisitor{
		ctx:    ctx,
		values: ds.NewStack[value.Value](),
	}
}

func (v *evalVisitor) Run(expr parse.Expr) (value.Value, error) {
	a, ok := expr.(parse.VisitableExpr[value.Value])
	if ok {
		return v.result, a.Accept(v)
	}
	return nil, ErrEval
}

func (v *evalVisitor) VisitUseRef(expr UseRef) error {
	return nil
}

func (v *evalVisitor) VisitImportFile(expr ImportFile) error {
	return nil
}

func (v *evalVisitor) VisitPrintRef(expr PrintRef) error {
	return nil
}

func (v *evalVisitor) VisitExportRef(expr ExportRef) error {
	return nil
}

func (v *evalVisitor) VisitLockRef(expr LockRef) error {
	return nil
}

func (v *evalVisitor) VisitUnlockRef(expr UnlockRef) error {
	return nil
}

func (v *evalVisitor) VisitAccess(expr Access) error {
	return nil
}

func (v *evalVisitor) VisitTemplate(expr Template) error {
	var str strings.Builder
	for _, e := range expr.Parts() {
		a, ok := e.(VisitableExpr)
		if !ok {
			return ErrEval
		}
		err := a.Accept(v)
		if err != nil {
			return err
		}
		val := v.popValue()
		str.WriteString(val.String())
	}
	v.pushValue(value.Text(str.String()))
	return nil
}

func (v *evalVisitor) VisitDeferred(expr Deferred) error {
	v.pushValue(expr)
	return nil
}

func (v *evalVisitor) VisitAssignment(expr Assignment) error {
	return nil
}

func (v *evalVisitor) VisitBinary(expr Binary) error {
	return nil
}

func (v *evalVisitor) VisitPostfix(expr Postfix) error {
	return nil
}

func (v *evalVisitor) VisitNot(expr Not) error {
	return nil
}

func (v *evalVisitor) VisitAnd(expr And) error {
	return nil
}

func (v *evalVisitor) VisitOr(expr Or) error {
	return nil
}

func (v *evalVisitor) VisitUnary(expr Unary) error {
	return nil
}

func (v *evalVisitor) VisitLiteral(expr Literal) error {
	val := value.Text(expr.Text())
	v.pushValue(val)
	return nil
}

func (v *evalVisitor) VisitNumber(expr Number) error {
	val := value.Float(expr.Float())
	v.pushValue(val)
	return nil
}

func (v *evalVisitor) VisitCall(expr Call) error {
	id, ok := expr.Name().(parse.Identifier)
	if !ok {
		return fmt.Errorf("identifier expected")
	}
	if fn, ok := specials[id.Ident()]; ok {
		return nil
	}
	if fn, ok := builtins.Registry[id.Ident()]; ok {
		var args []value.Value
		for _, a := range expr.Args() {
			a, ok := a.(parse.VisitableExpr)
			if !ok {
				return ErrEval
			}
			if err := a.Accept(v); err != nil {
				return err
			}
			args = append(args, v.popValue())
		}
		val, err := fn(args)
		if err == nil {
			v.pushValue(val)
		}
		return err
	}
	return fmt.Errorf("%s: builtin undefined", id.Ident())
}

func (v *evalVisitor) VisitClear(expr Clear) error {
	return nil
}

func (v *evalVisitor) VisitSlice(expr Slice) error {
	return nil
}

func (v *evalVisitor) VisitIdentifier(expr Identifier) error {
	val, err := v.resolve(expr.Ident())
	if err != nil {
		return value.ErrValue
	}
	v.pushValue(val)
	return nil
}

func (v *evalVisitor) VisitQualifiedCellAddr(expr QualifiedCellAddr) error {
	return nil
}

func (v *evalVisitor) VisitCellAddr(expr CellAddr) error {
	return nil
}

func (v *evalVisitor) VisitRangeAddr(expr RangeAddr) error {
	rg := types.NewRangeValue(expr.StartAt().Position, expr.EndAt().Position)
	v.pushValue(rg)
	return nil
}

func (v *evalVisitor) resolve(ident string) (value.Value, error) {
	return v.ctx.Resolve(ident)
}

func (v *evalVisitor) top() value.Value {
	val, ok := v.stack.Peek()
	if !ok {
		return value.ErrValue
	}
	return val
}

func (v *evalVisitor) pushValue(val value.Value) {
	v.stack.Push(val)
}

func (v *evalVisitor) popValue() value.Value {
	val, ok := v.stack.Pop()
	if !ok {
		return value.ErrValue
	}
	return val
}
