package eval

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/midbel/dockit/formula/builtins"
	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
)

var (
	ErrEval     = errors.New("expression can not be evaluated")
	ErrCallable = errors.New("expression is not callable")
)

func Eval(expr parse.Expr, ctx value.Context) (value.Value, error) {
	switch e := expr.(type) {
	case parse.Binary:
		return evalBinary(e, ctx)
	case parse.Unary:
		return evalUnary(e, ctx)
	case parse.Literal:
		return value.Text(e.Text()), nil
	case parse.Number:
		return value.Float(e.Float()), nil
	case parse.Call:
		return evalCall(e, ctx)
	case parse.CellAddr:
		return evalCellAddr(e, ctx)
	case parse.RangeAddr:
		return evalRangeAddr(e, ctx)
	default:
		return nil, ErrEval
	}
}

type Engine struct {
	Stdout io.Writer
	Stderr io.Writer

	phases  []scriptPhase
	loaders map[string]Loader
}

func NewEngine() *Engine {
	e := Engine{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		loaders: make(map[string]Loader),
	}
	e.RegisterLoader(".csv", CsvLoader())
	e.RegisterLoader(".xlsx", XlsxLoader())
	e.RegisterLoader(".ods", OdsLoader())
	e.RegisterLoader(".csv", CsvLoader())
	e.RegisterLoader(".log", LogLoader())
	return &e
}

func (e *Engine) RegisterLoader(kind string, loader Loader) {
	e.loaders[kind] = loader
}

func (e *Engine) Exec(r io.Reader, environ *env.Environment) (value.Value, error) {
	var (
		val   value.Value
		phase = phaseImport
		ctx   = NewEngineContext()
	)
	ctx.loaders = maps.Clone(e.loaders)
	ctx.PushContext(environ)

	ps, err := e.bootstrap(r, ctx)
	if err != nil {
		return nil, err
	}

	for {
		expr, err := ps.ParseNext()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if phase, err = execPhase(expr, phase); err != nil {
			return nil, err
		}
		if val, err = e.exec(expr, ctx); err != nil {
			return nil, err
		}
	}
	return val, nil
}

func (e *Engine) bootstrap(r io.Reader, ctx *EngineContext) (*parse.Parser, error) {
	scan, err := parse.Scan(r, parse.ModeScript)
	if err != nil {
		return nil, err
	}
	var (
		grammar *parse.Grammar
		mode    string
		cfg     = NewConfig()
	)
	if tok := scan.Peek(); tok.Type == op.Directive {
		mode = tok.Literal
		scan.Scan()
	}
	switch mode {
	case "", "script":
		grammar = parse.ScriptGrammar()
	case "pipeline":
	case "cube":
	case "command":
	default:
		return nil, fmt.Errorf("%s: unsupported mode", mode)
	}
	scan.SkipNL()
	for {
		tok := scan.Peek()
		if tok.Type != op.Pragma {
			break
		}
		scan.Scan()
		var ident []string
		for {
			tok := scan.Scan()
			if tok.Type == op.Dot {
				continue
			}
			if tok.Type == op.Assign {
				break
			}
			ident = append(ident, strings.TrimSpace(tok.Literal))
		}
		val := scan.Value()
		if tok := scan.Scan(); tok.Type != op.Eol {
			return nil, fmt.Errorf("newline expected")
		}
		if err := cfg.Set(ident, val); err != nil {
			return nil, err
		}
	}
	if err := ctx.Configure(cfg); err != nil {
		return nil, err
	}

	ps := parse.NewParser(grammar)
	ps.Attach(scan)
	return ps, nil
}

func (e *Engine) enterPhase(ph scriptPhase) {
	e.phases = append(e.phases, ph)
}

func (e *Engine) leavePhase() {
	n := len(e.phases) - 1
	if n <= 0 {
		return
	}
	e.phases = e.phases[:n]
}

func (e *Engine) inAssignment() bool {
	return e.inPhase(phaseAssign)
}

func (e *Engine) inPhase(ph scriptPhase) bool {
	return slices.Contains(e.phases, ph)
}

func (e *Engine) exec(expr parse.Expr, ctx *EngineContext) (value.Value, error) {
	switch expr := expr.(type) {
	case parse.ImportFile:
		return evalImport(e, expr, ctx)
	case parse.PrintRef:
		return evalPrint(e, expr, ctx)
	case parse.UseRef:
		return evalUse(e, expr, ctx)
	case parse.LockRef:
		return evalLock(e, expr, ctx)
	case parse.UnlockRef:
		return evalUnlock(e, expr, ctx)
	case parse.Assignment:
		e.enterPhase(phaseAssign)
		defer e.leavePhase()
		return evalAssignment(e, expr, ctx)
	case parse.Access:
		return evalAccess(e, expr, ctx)
	case parse.Literal:
		return value.Text(expr.Text()), nil
	case parse.Template:
		return evalTemplate(e, expr, ctx)
	case parse.Number:
		return value.Float(expr.Float()), nil
	case parse.Identifier:
		return evalScriptIdent(e, expr, ctx)
	case parse.Not:
		return evalScriptNot(e, expr, ctx)
	case parse.And:
		return evalScriptAnd(e, expr, ctx)
	case parse.Or:
		return evalScriptOr(e, expr, ctx)
	case parse.Binary:
		e.enterPhase(phaseBinary)
		defer e.leavePhase()
		return evalScriptBinary(e, expr, ctx)
	case parse.Unary:
		return evalScriptUnary(e, expr, ctx)
	case parse.Deferred:
		return evalDeferred(e, expr, ctx)
	case parse.Call:
		e.enterPhase(phaseCall)
		defer e.leavePhase()
		return evalScriptCall(e, expr, ctx)
	case parse.QualifiedCellAddr:
		return evalQualifiedCell(e, expr, ctx)
	case parse.CellAddr:
		return evalCell(e, expr, ctx)
	case parse.RangeAddr:
		return evalRange(e, expr, ctx)
	case parse.Slice:
		return evalSlice(e, expr, ctx)
	default:
		return nil, ErrEval
	}
}

func (e *Engine) execAndNormalize(expr parse.Expr, ctx *EngineContext) (value.Value, error) {
	val, err := e.exec(expr, ctx)
	if err != nil {
		return value.ErrValue, err
	}
	return e.normalizeValue(val, ctx)
}

func (e *Engine) normalizeValue(val value.Value, ctx *EngineContext) (value.Value, error) {
	switch val := val.(type) {
	case *types.Range:
		cl, err := ctx.PushReadable(val.Target())
		if err != nil {
			return value.ErrValue, err
		}
		defer cl.Close()
		rg := val.Range()
		return ctx.Context().Range(rg.Starts, rg.Ends)
	default:
		return val, nil
	}
}

func evalSlice(eg *Engine, expr parse.Slice, ctx *EngineContext) (value.Value, error) {
	var (
		val value.Value
		err error
	)
	if v := expr.View(); v == nil {
		val = ctx.CurrentActiveView()
	} else {
		val, err = eg.exec(v, ctx)
	}
	if err != nil {
		return nil, err
	}
	view, ok := val.(*types.View)
	if !ok {
		return nil, fmt.Errorf("slice can only be used on view")
	}
	switch e := expr.Expr().(type) {
	case parse.RangeAddr:
		view = view.BoundedView(e.Range())
	case parse.RangeSlice:
		view = view.BoundedView(e.Range())
	case parse.ColumnsSlice:
		view = view.ProjectView(e.Selection())
	case parse.Binary:
		p := types.NewExprPredicate(NewFormula(e))
		view = view.FilterView(p)
	case parse.And:
		f := NewFormula(parse.NewBinary(e.Left(), e.Right(), op.And))
		p := types.NewExprPredicate(f)
		view = view.FilterView(p)
	case parse.Or:
		f := NewFormula(parse.NewBinary(e.Left(), e.Right(), op.Or))
		p := types.NewExprPredicate(f)
		view = view.FilterView(p)
	case parse.Not:
		f := NewFormula(parse.NewUnary(e.Expr(), op.Not))
		p := types.NewExprPredicate(f)
		view = view.FilterView(p)
	case parse.Identifier:
	default:
		return nil, fmt.Errorf("invalid slice expression")
	}
	return view, nil
}

func evalScriptCall(eg *Engine, expr parse.Call, ctx *EngineContext) (value.Value, error) {
	id, ok := expr.Name().(parse.Identifier)
	if !ok {
		return value.ErrName, nil
	}
	if fn, ok := specials[id.Ident()]; ok {
		return fn.Eval(eg, expr.Args(), ctx)
	}
	if fn, ok := builtins.Registry[id.Ident()]; ok {
		var args []value.Value
		for _, a := range expr.Args() {
			v, err := eg.exec(a, ctx)
			if err != nil {
				return nil, err
			}
			args = append(args, v)
		}
		return fn(args)
	}
	return value.ErrName, nil
}

func evalRange(eg *Engine, expr parse.RangeAddr, ctx *EngineContext) (value.Value, error) {
	rg := types.NewRangeValue(expr.StartAt().Position, expr.EndAt().Position)
	return rg, nil
}

func evalQualifiedCell(eg *Engine, expr parse.QualifiedCellAddr, ctx *EngineContext) (value.Value, error) {
	var (
		cl  io.Closer
		err error
	)
	switch expr := expr.Path().(type) {
	case parse.Access:
		val, err := eg.exec(expr, ctx)
		if err != nil {
			return nil, err
		}
		cl, err = ctx.PushValue(val, expr.Property())
	case parse.Identifier:
		cl, err = ctx.PushReadable(expr.Ident())
	default:
		return nil, fmt.Errorf("no view can be found from expr")
	}
	if err != nil {
		return nil, err
	}
	defer cl.Close()

	switch a := expr.Addr().(type) {
	case parse.CellAddr:
		return ctx.Context().At(a.Position)
	case parse.RangeAddr:
		return ctx.Context().Range(a.StartAt().Position, a.EndAt().Position)
	default:
		return value.ErrValue, nil
	}
}

func evalCell(eg *Engine, expr parse.CellAddr, ctx *EngineContext) (value.Value, error) {
	cl, err := ctx.PushReadable(expr.Sheet)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	return ctx.Context().At(expr.Position)
}

func evalDeferred(eg *Engine, expr parse.Deferred, ctx *EngineContext) (value.Value, error) {
	return expr, nil
}

func evalScriptIdent(eg *Engine, expr parse.Identifier, ctx *EngineContext) (value.Value, error) {
	v, err := ctx.Resolve(expr.Ident())
	if err != nil {
		return nil, err
	}
	if d, ok := v.(parse.Deferred); !eg.inAssignment() && ok {
		return eg.exec(d.Expr(), ctx)
	}
	return v, nil
}

func evalTemplate(eg *Engine, expr parse.Template, ctx *EngineContext) (value.Value, error) {
	var str strings.Builder
	for _, e := range expr.Parts() {
		v, err := eg.exec(e, ctx)
		if err != nil {
			return nil, err
		}
		str.WriteString(v.String())
	}
	return value.Text(str.String()), nil
}

func evalScriptNot(eg *Engine, e parse.Not, ctx *EngineContext) (value.Value, error) {
	val, err := eg.exec(e.Expr(), ctx)
	if err != nil {
		return value.ErrValue, err
	}
	ok := value.True(val)
	return value.Boolean(!ok), nil
}

func evalScriptAnd(eg *Engine, e parse.And, ctx *EngineContext) (value.Value, error) {
	left, err := eg.exec(e.Left(), ctx)
	if err != nil {
		return value.ErrValue, err
	}
	right, err := eg.exec(e.Right(), ctx)
	if err != nil {
		return nil, err
	}
	ok := value.True(left) && value.True(right)
	return value.Boolean(ok), nil
}

func evalScriptOr(eg *Engine, e parse.Or, ctx *EngineContext) (value.Value, error) {
	left, err := eg.exec(e.Left(), ctx)
	if err != nil {
		return value.ErrValue, err
	}
	right, err := eg.exec(e.Right(), ctx)
	if err != nil {
		return nil, err
	}
	ok := value.True(left) || value.True(right)
	return value.Boolean(ok), nil
}

func evalScriptBinary(eg *Engine, e parse.Binary, ctx *EngineContext) (value.Value, error) {
	left, err := eg.execAndNormalize(e.Left(), ctx)
	if err != nil {
		return nil, err
	}
	right, err := eg.execAndNormalize(e.Right(), ctx)
	if err != nil {
		return nil, err
	}
	switch {
	case value.IsScalar(left) && value.IsScalar(right):
		return evalScalarBinary(left, right, e.Op())
	case (value.IsArray(left) || value.IsObject(left)) && value.IsScalar(right):
		return evalScalarArrayBinary(right, left, e.Op())
	case value.IsArray(left) && value.IsArray(right):
		return evalArrayBinary(left, right, e.Op())
	case value.IsObject(left) && value.IsObject(right):
		return evalViewBinary(left, right, e.Op())
	default:
		return value.ErrValue, nil
	}
}

func evalScalarBinary(left, right value.Value, oper op.Op) (value.Value, error) {
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

func evalScalarArrayBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	if v, ok := right.(*types.View); ok {
		right = v.AsArray()
	}
	arr, err := value.CastToArray(right)
	if err != nil {
		return value.ErrValue, nil
	}
	err = arr.Apply(func(val value.ScalarValue) (value.ScalarValue, error) {
		ret, err := evalScalarBinary(left, val, oper)
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

func evalArrayBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	larr, err := value.CastToArray(left)
	if err != nil {
		return value.ErrValue, nil
	}
	rarr, err := value.CastToArray(right)
	if err != nil {
		return value.ErrValue, nil
	}
	res, err := larr.ApplyArray(rarr, func(left, right value.ScalarValue) (value.ScalarValue, error) {
		ret, err := evalScalarBinary(left, right, oper)
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

func evalViewBinary(left, right value.Value, oper op.Op) (value.Value, error) {
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
		return evalArrayBinary(lv.AsArray(), rv.AsArray(), oper)
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

func evalScriptUnary(eg *Engine, e parse.Unary, ctx *EngineContext) (value.Value, error) {
	val, err := eg.exec(e.Expr(), ctx)
	if err != nil {
		return nil, err
	}
	n, err := value.CastToFloat(val)
	switch e.Op() {
	case op.Add:
		return n, nil
	case op.Sub:
		return value.Float(float64(-n)), nil
	default:
		return value.ErrValue, nil
	}
}

func evalQualifiedAssignment(eg *Engine, expr parse.QualifiedCellAddr, ctx *EngineContext) (LValue, io.Closer, error) {
	var (
		lv  LValue
		cl  io.Closer
		err error
	)
	switch expr := expr.Path().(type) {
	case parse.Identifier:
		cl, err = ctx.PushMutable(expr.Ident())
	case parse.Access:
		val, err1 := eg.exec(expr, ctx)
		if err1 != nil {
			err = err1
			break
		}
		cl, err = ctx.PushValue(val, expr.Property())
	default:
		err = fmt.Errorf("expression can not be assigned to %s", expr)
	}
	if err != nil {
		return nil, nil, err
	}
	lv, err = resolveQualified(ctx, expr.Addr())
	return lv, cl, err
}

func evalAssignmentTarget(eg *Engine, expr parse.Expr, ctx *EngineContext) (LValue, io.Closer, error) {
	var (
		lv  LValue
		cl  io.Closer
		err error
	)
	switch expr := expr.(type) {
	case parse.CellAddr:
		cl, err = ctx.PushMutable("")
		if err != nil {
			break
		}
		lv, err = resolveCell(ctx, expr)
	case parse.RangeAddr:
		cl, err = ctx.PushMutable("")
		if err != nil {
			break
		}
		lv, err = resolveRange(ctx, expr)
	case parse.QualifiedCellAddr:
		return evalQualifiedAssignment(eg, expr, ctx)
	case parse.Access:
	case parse.Identifier:
		lv, err = resolveIdent(ctx, expr)
	default:
		err = fmt.Errorf("value can not be assigned to %s", expr)
	}
	return lv, cl, err
}

func evalAssignment(eg *Engine, e parse.Assignment, ctx *EngineContext) (value.Value, error) {
	lv, cl, err := evalAssignmentTarget(eg, e.Ident(), ctx)
	if err != nil {
		return nil, err
	}
	if cl != nil {
		defer cl.Close()
	}
	value, err := eg.execAndNormalize(e.Expr(), ctx)
	if err != nil {
		return nil, err
	}
	return nil, lv.Set(value)
}

func evalImport(eg *Engine, e parse.ImportFile, ctx *EngineContext) (value.Value, error) {
	file, err := ctx.Open(e.File(), nil)
	if err != nil {
		return nil, err
	}
	alias := e.Alias()
	if file := e.File(); alias == "" {
		for {
			ext := filepath.Ext(file)
			if ext == "" {
				break
			}
			alias = strings.TrimSuffix(file, ext)
		}
	}
	book := types.NewFileValue(file, e.ReadOnly())
	if ev, ok := ctx.Context().(interface{ Define(string, value.Value) }); ok {
		ev.Define(alias, book)
	}
	if e.Default() {
		ctx.SetDefault(book)
	}
	return value.Empty(), nil
}

func evalPush(eg *Engine, e parse.Push, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalPop(eg *Engine, e parse.Pop, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalClear(eg *Engine, e parse.Clear, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalLock(eg *Engine, e parse.LockRef, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalUnlock(eg *Engine, e parse.UnlockRef, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalUse(eg *Engine, e parse.UseRef, ctx *EngineContext) (value.Value, error) {
	v, err := ctx.Resolve(e.Identifier())
	if err != nil {
		return nil, err
	}
	switch v := v.(type) {
	case *types.File, *types.View:
		ctx.SetDefault(v)
	default:
		return nil, fmt.Errorf("default can only be used with file or view")
	}
	return value.Empty(), nil
}

func evalPrint(eg *Engine, e parse.PrintRef, ctx *EngineContext) (value.Value, error) {
	v, err := eg.execAndNormalize(e.Expr(), ctx)
	if err != nil {
		return nil, err
	}
	ctx.Print(v)
	return value.Empty(), nil
}

func evalAccess(eg *Engine, e parse.Access, ctx *EngineContext) (value.Value, error) {
	obj, err := eg.exec(e.Object(), ctx)
	if err != nil {
		return nil, err
	}
	g, ok := obj.(value.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("object expected")
	}
	return g.Get(e.Property())
}

func evalBinary(e parse.Binary, ctx value.Context) (value.Value, error) {
	left, err := Eval(e.Left(), ctx)
	if err != nil {
		return nil, err
	}
	right, err := Eval(e.Right(), ctx)
	if err != nil {
		return nil, err
	}

	switch e.Op() {
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
	case op.And:
		ok := value.True(left) && value.True(right)
		return value.Boolean(ok), nil
	case op.Or:
		ok := value.True(left) || value.True(right)
		return value.Boolean(ok), nil
	default:
		return value.ErrValue, nil
	}
}

func evalUnary(e parse.Unary, ctx value.Context) (value.Value, error) {
	val, err := Eval(e.Expr(), ctx)
	if err != nil {
		return nil, err
	}
	n, ok := val.(value.Float)
	if !ok {
		return value.ErrValue, nil
	}
	switch e.Op() {
	case op.Not:
		ok := value.True(val)
		return value.Boolean(!ok), nil
	case op.Add:
		return n, nil
	case op.Sub:
		return value.Float(float64(-n)), nil
	default:
		return value.ErrValue, nil
	}
}

func evalCall(e parse.Call, ctx value.Context) (value.Value, error) {
	id, ok := e.Name().(parse.Identifier)
	if !ok {
		return value.ErrName, nil
	}
	var args []value.Arg
	for _, a := range e.Args() {
		args = append(args, makeArg(a))
	}
	fn, err := ctx.Resolve(id.Ident())
	if err != nil {
		return nil, err
	}
	if fn.Kind() != value.KindFunction {
		return nil, fmt.Errorf("%s: %w", id.Ident(), ErrCallable)
	}
	call, ok := fn.(value.FunctionValue)
	return call.Call(args, ctx)
}

func evalCellAddr(e parse.CellAddr, ctx value.Context) (value.Value, error) {
	return ctx.At(e.Position)
}

func evalRangeAddr(e parse.RangeAddr, ctx value.Context) (value.Value, error) {
	return ctx.Range(e.StartAt().Position, e.EndAt().Position)
}
