package eval

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/internal/ds"
	"github.com/midbel/dockit/value"
)

type evalVisitor struct {
	ctx    *EngineContext
	stack  *ds.Stack[value.Value]
	phases *ds.Stack[scriptPhase]
}

func evalScript(ctx *EngineContext) *evalVisitor {
	return &evalVisitor{
		ctx:    ctx,
		stack:  ds.NewStack[value.Value](),
		phases: ds.NewStack[scriptPhase](),
	}
}

func (v *evalVisitor) Run(expr parse.Expr) (value.Value, error) {
	if err := v.visitExpr(expr); err != nil {
		return value.ErrValue, err
	}
	return v.popValue(), nil
}

func (v *evalVisitor) VisitScript(expr parse.Script) error {
	for i := range expr.Body {
		if err := v.visitExpr(expr.Body[i]); err != nil {
			return err
		}
	}
	return nil
}

func (v *evalVisitor) VisitUseRef(expr parse.UseRef) error {
	val, err := v.resolve(expr.Identifier())
	if err != nil {
		return err
	}
	switch val.(type) {
	case *types.File, *types.View:
		v.ctx.SetDefault(val)
	default:
		return fmt.Errorf("file/view expected")
	}
	return nil
}

func (v *evalVisitor) VisitImportFile(expr parse.ImportFile) error {
	options := expr.Options()
	switch spec := expr.Specifier(); expr.Format() {
	case "csv":
		options["delimiter"] = spec
	case "log":
		options["pattern"] = spec
	default:
	}
	file, err := v.ctx.Open(expr.File(), options)
	if err != nil {
		return err
	}
	alias := expr.Alias()
	if file := expr.File(); alias == "" {
		for {
			ext := filepath.Ext(file)
			if ext == "" {
				break
			}
			alias = strings.TrimSuffix(file, ext)
		}
	}
	wb := types.NewFileValue(file, expr.ReadOnly())
	if ev, ok := v.ctx.Context().(interface{ Define(string, value.Value) }); ok {
		ev.Define(alias, wb)
	}
	if expr.Default() {
		v.ctx.SetDefault(wb)
	}
	return nil
}

func (v *evalVisitor) VisitPrintRef(expr parse.PrintRef) error {
	val, err := v.visitNormalize(expr.Expr())
	if err != nil {
		return err
	}
	v.ctx.Print(val)
	return err
}

func (v *evalVisitor) VisitExportRef(expr parse.ExportRef) error {
	return nil
}

func (v *evalVisitor) VisitAccess(expr parse.Access) error {
	if err := v.visitExpr(expr.Object()); err != nil {
		return err
	}
	obj, ok := v.popValue().(value.ObjectValue)
	if !ok {
		return fmt.Errorf("object expected")
	}
	val, err := obj.Get(expr.Property())
	if err == nil {
		val = value.ErrValue
	}
	v.pushValue(val)
	return err
}

func (v *evalVisitor) VisitTemplate(expr parse.Template) error {
	var str strings.Builder
	for _, e := range expr.Parts() {
		if err := v.visitExpr(e); err != nil {
			return err
		}
		val := v.popValue()
		str.WriteString(val.String())
	}
	v.pushValue(value.Text(str.String()))
	return nil
}

func (v *evalVisitor) VisitDeferred(expr parse.Deferred) error {
	v.pushValue(expr)
	return nil
}

func (v *evalVisitor) VisitAssignment(expr parse.Assignment) error {
	v.enterPhase(phaseAssign)
	defer v.leavePhase()

	lv, cl, err := v.assignmentTarget(expr.Ident())
	if err != nil {
		return err
	}
	if cl != nil {
		defer cl.Close()
	}
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	val, err := v.normalize(v.popValue())
	if err == nil {
		return err
	}
	return lv.Set(val)
}

func (v *evalVisitor) assignmentTarget(expr parse.Expr) (LValue, io.Closer, error) {
	var (
		lv  LValue
		cl  io.Closer
		err error
	)
	switch expr := expr.(type) {
	case parse.CellAddr:
		cl, err = v.ctx.PushMutable("")
		if err != nil {
			break
		}
		lv, err = resolveCell(v.ctx, expr)
	case parse.RangeAddr:
		cl, err = v.ctx.PushMutable("")
		if err != nil {
			break
		}
		lv, err = resolveRange(v.ctx, expr)
	case parse.QualifiedCellAddr:
		return v.qualifiedAssignment(expr)
	case parse.Access:
	case parse.Identifier:
		lv, err = resolveIdent(v.ctx, expr)
	default:
		err = fmt.Errorf("value can not be assigned to %s", expr)
	}
	return lv, cl, err
}

func (v *evalVisitor) qualifiedAssignment(expr parse.QualifiedCellAddr) (LValue, io.Closer, error) {
	var (
		lv  LValue
		cl  io.Closer
		err error
	)
	switch expr := expr.Path().(type) {
	case parse.Identifier:
		cl, err = v.ctx.PushMutable(expr.Ident())
	case parse.Access:
		if err1 := expr.Accept(v); err1 != nil {
			err = err1
			break
		}
		cl, err = v.ctx.PushValue(v.popValue(), expr.Property())
	default:
		err = fmt.Errorf("expression can not be assigned to %s", expr)
	}
	if err != nil {
		return nil, nil, err
	}
	lv, err = resolveQualified(v.ctx, expr.Addr())
	return lv, cl, err
}

func (v *evalVisitor) VisitBinary(expr parse.Binary) error {
	v.enterPhase(phaseBinary)
	defer v.leavePhase()

	left, err := v.visitNormalize(expr.Left())
	if err != nil {
		v.pushValue(value.ErrValue)
		return err
	}
	right, err := v.visitNormalize(expr.Right())
	if err != nil {
		v.pushValue(value.ErrValue)
		return err
	}
	var val value.Value
	switch {
	case value.IsScalar(left) && value.IsScalar(right):
		val, err = v.evalScalarBinary(left, right, expr.Op())
	case (value.IsArray(left) || value.IsObject(left)) && value.IsScalar(right):
		val, err = v.evalScalarArrayBinary(right, left, expr.Op())
	case value.IsArray(left) && value.IsArray(right):
		val, err = v.evalArrayBinary(left, right, expr.Op())
	case value.IsObject(left) && value.IsObject(right):
		val, err = v.evalViewBinary(left, right, expr.Op())
	default:
		val = value.ErrValue
	}
	v.pushValue(val)
	return err
}

func (v *evalVisitor) evalScalarBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	switch oper {
	case op.Add:
		return value.Add(left, right)
	case op.Sub:
		return value.Sub(left, right)
	case op.Mul:
		return value.Mul(left, right)
	case op.Div:
		return value.Div(left, right)
	case op.Pow:
		return value.Pow(left, right)
	case op.Concat:
		return value.Concat(left, right)
	case op.Eq:
		return value.Eq(left, right)
	case op.Ne:
		return value.Ne(left, right)
	case op.Lt:
		return value.Lt(left, right)
	case op.Le:
		return value.Le(left, right)
	case op.Gt:
		return value.Gt(left, right)
	case op.Ge:
		return value.Ge(left, right)
	default:
		return value.ErrValue, nil
	}
}

func (v *evalVisitor) evalScalarArrayBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	if v, ok := right.(*types.View); ok {
		right = v.AsArray()
	}
	arr, err := value.CastToArray(right)
	if err != nil {
		return value.ErrValue, nil
	}
	err = arr.Apply(func(val value.ScalarValue) (value.ScalarValue, error) {
		ret, err := v.evalScalarBinary(left, val, oper)
		if err != nil {
			return value.ErrValue, err
		}
		scalar, ok := ret.(value.ScalarValue)
		if !ok {
			return value.ErrValue, nil
		}
		return scalar, nil
	})
	if err != nil {
		return value.ErrValue, nil
	}
	return arr, nil
}

func (v *evalVisitor) evalArrayBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	larr, err := value.CastToArray(left)
	if err != nil {
		return value.ErrValue, nil
	}
	rarr, err := value.CastToArray(right)
	if err != nil {
		return value.ErrValue, nil
	}
	res, err := larr.ApplyArray(rarr, func(left, right value.ScalarValue) (value.ScalarValue, error) {
		ret, err := v.evalScalarBinary(left, right, oper)
		if err != nil {
			return value.ErrValue, err
		}
		scalar, ok := ret.(value.ScalarValue)
		if !ok {
			return value.ErrValue, nil
		}
		return scalar, nil
	})
	if err != nil {
		return value.ErrValue, err
	}
	return res, nil
}

func (v *evalVisitor) evalViewBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	lv, ok := left.(*types.View)
	if !ok {
		return value.ErrValue, nil
	}
	rv, ok := right.(*types.View)
	if !ok {
		return value.ErrValue, nil
	}
	var (
		view grid.View
		v1   = lv.View()
		v2   = rv.View()
		d1   = v1.Bounds()
		d2   = v2.Bounds()
	)
	switch oper {
	case op.Add, op.Sub, op.Mul, op.Div, op.Pow:
		return v.evalArrayBinary(lv.AsArray(), rv.AsArray(), oper)
	case op.Union:
		if d1.Width() != d2.Width() {
			return value.ErrValue, fmt.Errorf("view can not be combined - number of columns mismatched")
		}
		view = grid.VerticalView(v1, v2)
	case op.Concat:
		if d1.Height() != d2.Height() {
			return value.ErrValue, fmt.Errorf("view can not be combined - number of lines mismatched")
		}
		view = grid.HorizontalView(v1, v2)
	default:
		return value.ErrValue, nil
	}
	return types.NewViewValue(view), nil
}

func (v *evalVisitor) VisitPostfix(expr parse.Postfix) error {
	return nil
}

func (v *evalVisitor) VisitNot(expr parse.Not) error {
	val, err := v.visitNormalize(expr.Expr())
	if err != nil {
		v.pushValue(value.ErrValue)
		return err
	}
	ok := value.True(val)
	v.pushValue(value.Boolean(!ok))
	return nil
}

func (v *evalVisitor) VisitAnd(expr parse.And) error {
	left, err := v.visitNormalize(expr.Left())
	if err != nil {
		v.pushValue(value.ErrValue)
		return err
	}
	right, err := v.visitNormalize(expr.Right())
	if err != nil {
		v.pushValue(value.ErrValue)
		return err
	}
	ok := value.True(left) && value.True(right)
	v.pushValue(value.Boolean(ok))
	return nil
}

func (v *evalVisitor) VisitOr(expr parse.Or) error {
	left, err := v.visitNormalize(expr.Left())
	if err != nil {
		v.pushValue(value.ErrValue)
		return err
	}
	right, err := v.visitNormalize(expr.Right())
	if err != nil {
		v.pushValue(value.ErrValue)
		return err
	}
	ok := value.True(left) || value.True(right)
	v.pushValue(value.Boolean(ok))
	return nil
}

func (v *evalVisitor) VisitUnary(expr parse.Unary) error {
	val, err := v.visitNormalize(expr.Expr())
	if err != nil {
		return err
	}
	x, err := value.CastToFloat(val)
	switch expr.Op() {
	case op.Add:
		val = x
	case op.Sub:
		val = value.Float(float64(-x))
	default:
		val = value.ErrValue
	}
	if err != nil {
		val = value.ErrValue
	}
	v.pushValue(val)
	return nil
}

func (v *evalVisitor) VisitLiteral(expr parse.Literal) error {
	val := value.Text(expr.Text())
	v.pushValue(val)
	return nil
}

func (v *evalVisitor) VisitNumber(expr parse.Number) error {
	val := value.Float(expr.Float())
	v.pushValue(val)
	return nil
}

func (v *evalVisitor) VisitCall(expr parse.Call) error {
	v.enterPhase(phaseCall)
	defer v.leavePhase()

	id, ok := expr.Name().(parse.Identifier)
	if !ok {
		return fmt.Errorf("identifier expected")
	}
	if fn, ok := specials[id.Ident()]; ok {
		val, err := fn.Run(v, expr.Args(), v.ctx)
		if err != nil {
			val = value.ErrValue
		}
		v.pushValue(val)
		return err
	}
	if fn, ok := builtins.Registry[id.Ident()]; ok {
		var args []value.Value
		for _, a := range expr.Args() {
			if err := v.visitExpr(a); err != nil {
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

func (v *evalVisitor) VisitClear(expr parse.Clear) error {
	return nil
}

func (v *evalVisitor) VisitSlice(expr parse.Slice) error {
	var (
		val value.Value
		err error
	)
	if view := expr.View(); view == nil {
		val = v.ctx.CurrentActiveView()
	} else {
		val, err = v.visitNormalize(view)
	}
	if err != nil {
		return err
	}
	view, ok := val.(*types.View)
	if !ok {
		return fmt.Errorf("slice can only be used on view")
	}
	switch e := expr.Expr().(type) {
	case parse.RangeAddr:
		view = view.BoundedView(e.Range())
	case parse.ColumnsSlice:
		view = view.ProjectView(e.Selection())
	case parse.Binary:
		p := types.NewExprPredicate(grid.NewFormula(e))
		view = view.FilterView(p)
	case parse.And:
		f := grid.NewFormula(parse.NewBinary(e.Left(), e.Right(), op.And))
		p := types.NewExprPredicate(f)
		view = view.FilterView(p)
	case parse.Or:
		f := grid.NewFormula(parse.NewBinary(e.Left(), e.Right(), op.Or))
		p := types.NewExprPredicate(f)
		view = view.FilterView(p)
	case parse.Not:
		f := grid.NewFormula(parse.NewUnary(e.Expr(), op.Not))
		p := types.NewExprPredicate(f)
		view = view.FilterView(p)
	case parse.Identifier:
	default:
		return fmt.Errorf("invalid slice expression")
	}
	v.pushValue(view)
	return nil
}

func (v *evalVisitor) VisitIdentifier(expr parse.Identifier) error {
	val, err := v.resolve(expr.Ident())
	if err != nil {
		return value.ErrValue
	}
	v.pushValue(val)
	return nil
}

func (v *evalVisitor) VisitQualifiedCellAddr(expr parse.QualifiedCellAddr) error {
	var (
		cl  io.Closer
		err error
		val value.Value
	)
	switch expr := expr.Path().(type) {
	case parse.Access:
		val, err = v.visitNormalize(expr)
		if err != nil {
			return err
		}
		cl, err = v.ctx.PushValue(val, expr.Property())
	case parse.Identifier:
		cl, err = v.ctx.PushReadable(expr.Ident())
	default:
		return fmt.Errorf("no view can be found from expr")
	}
	if err != nil {
		return err
	}
	defer cl.Close()

	switch a := expr.Addr().(type) {
	case parse.CellAddr:
		val, err = v.ctx.Context().At(a.Position)
	case parse.RangeAddr:
		val, err = v.ctx.Context().Range(a.StartAt().Position, a.EndAt().Position)
	default:
		val = value.ErrValue
	}
	v.pushValue(val)
	return err
}

func (v *evalVisitor) VisitCellAddr(expr parse.CellAddr) error {
	cl, err := v.ctx.PushReadable(expr.Sheet)
	if err != nil {
		return err
	}
	defer cl.Close()

	val, err := v.ctx.Context().At(expr.Position)
	if err == nil {
		v.pushValue(val)
	} else {
		v.pushValue(value.ErrValue)
	}
	return nil
}

func (v *evalVisitor) VisitRangeAddr(expr parse.RangeAddr) error {
	rg := types.NewRangeValue(expr.StartAt().Position, expr.EndAt().Position)
	v.pushValue(rg)
	return nil
}

func (v *evalVisitor) visitExpr(expr parse.Expr) error {
	a, ok := expr.(parse.VisitableExpr)
	if !ok {
		return ErrEval
	}
	return a.Accept(v)
}

func (v *evalVisitor) visitNormalize(expr parse.Expr) (value.Value, error) {
	if err := v.visitExpr(expr); err != nil {
		return value.ErrValue, err
	}
	return v.normalize(v.popValue())
}

func (v *evalVisitor) normalize(val value.Value) (value.Value, error) {
	switch val := val.(type) {
	case *types.Range:
		cl, err := v.ctx.PushReadable(val.Target())
		if err != nil {
			return value.ErrValue, err
		}
		defer cl.Close()
		rg := val.Range()

		return v.ctx.Context().Range(rg.Starts, rg.Ends)
	default:
		return val, nil
	}
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

func (v *evalVisitor) enterPhase(ph scriptPhase) {
	v.phases.Push(ph)
}

func (v *evalVisitor) leavePhase() {
	v.phases.Pop()
}

func (v *evalVisitor) inAssignment() bool {
	return v.inPhase(phaseAssign)
}

func (v *evalVisitor) inPhase(ph scriptPhase) bool {
	top, _ := v.phases.Peek()
	return top == ph
}
