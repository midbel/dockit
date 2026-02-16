package eval

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/midbel/dockit/formula/builtins"
	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
)

var (
	ErrEval     = errors.New("expression can not be evaluated")
	ErrCallable = errors.New("expression is not callable")
)

func Eval(expr Expr, ctx value.Context) (value.Value, error) {
	switch e := expr.(type) {
	case binary:
		return evalBinary(e, ctx)
	case unary:
		return evalUnary(e, ctx)
	case literal:
		return value.Text(e.value), nil
	case number:
		return value.Float(e.value), nil
	case call:
		return evalCall(e, ctx)
	case cellAddr:
		return evalCellAddr(e, ctx)
	case rangeAddr:
		return evalRangeAddr(e, ctx)
	default:
		return nil, ErrEval
	}
}

type scriptPhase int8

const (
	phaseStmt scriptPhase = 1 << iota
	phaseImport
	phaseBinary
	phaseCall
	phaseAssign
)

func (p scriptPhase) Allows(k Kind) bool {
	switch p {
	case phaseStmt:
		return k == KindStmt
	case phaseImport:
		return k == KindStmt || k == KindImport
	default:
		return false
	}
}

func (p scriptPhase) Next(k Kind) scriptPhase {
	switch {
	case p == phaseImport && k == KindImport:
		return p
	case p == phaseImport && k == KindStmt:
		return phaseStmt
	case p == phaseStmt && k == KindStmt:
		return p
	default:
		return p
	}
}

func execPhase(expr Expr, phase scriptPhase) (scriptPhase, error) {
	currKind := KindStmt
	if ek, ok := expr.(ExprKind); ok {
		currKind = ek.Kind()
	}
	if !phase.Allows(currKind) {
		return phase, fmt.Errorf("unknown script phase!")
	}
	return phase.Next(currKind), nil
}

type Engine struct {
	Loader
	Config EngineConfig
	Stdout io.Writer
	Stderr io.Writer

	printMode PrintMode
	phases    []scriptPhase
}

func NewEngine(loader Loader) *Engine {
	e := Engine{
		Loader:    loader,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		printMode: PrintDefault,
	}
	e.Config.Print.Cols = maxCols
	e.Config.Print.Rows = maxRows
	return &e
}

func (e *Engine) SetPrintMode(mode PrintMode) {
	e.Config.Print.Debug = mode == PrintDebug
}

func (e *Engine) Exec(r io.Reader, environ *env.Environment) (value.Value, error) {
	var (
		val   value.Value
		phase = phaseImport
		ctx   = NewEngineContext()
	)
	ctx.PushContext(environ)
	ctx.config = e.Config
	ctx.config.Stdout = e.Stdout
	ctx.config.Stderr = e.Stderr

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

func (e *Engine) bootstrap(r io.Reader, ctx *EngineContext) (*Parser, error) {
	scan, err := Scan(r, ModeScript)
	if err != nil {
		return nil, err
	}
	var (
		grammar *Grammar
		mode    string
	)
	if tok := scan.Peek(); tok.Type == op.Directive {
		mode = tok.Literal
		scan.Scan()
	}
	switch mode {
	case "", "script":
		grammar = ScriptGrammar()
	case "pipeline":
	case "cube":
	case "command":
	default:
		return nil, fmt.Errorf("%s: unsupported mode", mode)
	}
	scan.skipNL()
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
		err := directiveTrie.Configure(ident, val, &ctx.config)
		if err != nil {
			return nil, err
		}
	}

	ps := NewParser(grammar)
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

func (e *Engine) exec(expr Expr, ctx *EngineContext) (value.Value, error) {
	switch expr := expr.(type) {
	case importFile:
		return evalImport(e, expr, ctx)
	case printRef:
		return evalPrint(e, expr, ctx)
	case useRef:
		return evalUse(e, expr, ctx)
	case lockRef:
		return evalLock(e, expr, ctx)
	case unlockRef:
		return evalUnlock(e, expr, ctx)
	case assignment:
		e.enterPhase(phaseAssign)
		defer e.leavePhase()
		return evalAssignment(e, expr, ctx)
	case access:
		return evalAccess(e, expr, ctx)
	case literal:
		return value.Text(expr.value), nil
	case template:
		return evalTemplate(e, expr, ctx)
	case number:
		return value.Float(expr.value), nil
	case identifier:
		return evalScriptIdent(e, expr, ctx)
	case binary:
		e.enterPhase(phaseBinary)
		defer e.leavePhase()
		return evalScriptBinary(e, expr, ctx)
	case unary:
		return evalScriptUnary(e, expr, ctx)
	case deferred:
		return evalDeferred(e, expr, ctx)
	case call:
		e.enterPhase(phaseCall)
		defer e.leavePhase()
		return evalScriptCall(e, expr, ctx)
	case qualifiedCellAddr:
		return evalQualifiedCell(e, expr, ctx)
	case cellAddr:
		return evalCell(e, expr, ctx)
	case rangeAddr:
		return evalRange(e, expr, ctx)
	case slice:
		return evalSlice(e, expr, ctx)
	default:
		return nil, ErrEval
	}
}

func (e *Engine) execAndNormalize(expr Expr, ctx *EngineContext) (value.Value, error) {
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

func evalSlice(eg *Engine, expr slice, ctx *EngineContext) (value.Value, error) {
	var (
		val value.Value
		err error
	)
	if expr.view == nil {
		val = ctx.CurrentActiveView()
	} else {
		val, err = eg.exec(expr.view, ctx)
	}
	if err != nil {
		return nil, err
	}
	view, ok := val.(*types.View)
	if !ok {
		return nil, fmt.Errorf("slice can only be used on view")
	}
	switch e := expr.expr.(type) {
	case rangeAddr:
		view.BoundedView(e.Range())
	case rangeSlice:
		view.BoundedView(e.Range())
	case columnsSlice:
		view.ProjectView(e.Selection())
	case filterSlice:
		view.FilterView(e.Predicate())
	case identifier:
	default:
		return nil, fmt.Errorf("invalid slice expression")
	}
	return view, nil
}

func evalScriptCall(eg *Engine, expr call, ctx *EngineContext) (value.Value, error) {
	id, ok := expr.ident.(identifier)
	if !ok {
		return value.ErrName, nil
	}
	if fn, ok := specials[id.name]; ok {
		return fn.Eval(eg, expr.args, ctx)
	}
	if fn, ok := builtins.Registry[id.name]; ok {
		var args []value.Value
		for _, a := range expr.args {
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

func evalRange(eg *Engine, expr rangeAddr, ctx *EngineContext) (value.Value, error) {
	rg := types.NewRangeValue(expr.startAddr.Position, expr.endAddr.Position)
	return rg, nil
}

func evalQualifiedCell(eg *Engine, expr qualifiedCellAddr, ctx *EngineContext) (value.Value, error) {
	var (
		cl  io.Closer
		err error
	)
	switch expr := expr.path.(type) {
	case access:
		val, err := eg.exec(expr, ctx)
		if err != nil {
			return nil, err
		}
		cl, err = ctx.PushValue(val, expr.prop)
	case identifier:
		cl, err = ctx.PushReadable(expr.name)
	default:
		return nil, fmt.Errorf("no view can be found from expr")
	}
	if err != nil {
		return nil, err
	}
	defer cl.Close()

	switch a := expr.addr.(type) {
	case cellAddr:
		return ctx.Context().At(a.Position)
	case rangeAddr:
		return ctx.Context().Range(a.startAddr.Position, a.endAddr.Position)
	default:
		return value.ErrValue, nil
	}
}

func evalCell(eg *Engine, expr cellAddr, ctx *EngineContext) (value.Value, error) {
	cl, err := ctx.PushReadable(expr.Sheet)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	return ctx.Context().At(expr.Position)
}

func evalDeferred(eg *Engine, expr deferred, ctx *EngineContext) (value.Value, error) {
	return expr, nil
}

func evalScriptIdent(eg *Engine, expr identifier, ctx *EngineContext) (value.Value, error) {
	v, err := ctx.Resolve(expr.name)
	if err != nil {
		return nil, err
	}
	if d, ok := v.(deferred); !eg.inAssignment() && ok {
		return eg.exec(d.expr, ctx)
	}
	return v, nil
}

func evalTemplate(eg *Engine, expr template, ctx *EngineContext) (value.Value, error) {
	var str strings.Builder
	for i := range expr.expr {
		v, err := eg.exec(expr.expr[i], ctx)
		if err != nil {
			return nil, err
		}
		str.WriteString(v.String())
	}
	return value.Text(str.String()), nil
}

func evalScriptBinary(eg *Engine, e binary, ctx *EngineContext) (value.Value, error) {
	left, err := eg.execAndNormalize(e.left, ctx)
	if err != nil {
		return nil, err
	}
	right, err := eg.execAndNormalize(e.right, ctx)
	if err != nil {
		return nil, err
	}
	switch {
	case value.IsScalar(left) && value.IsScalar(right):
		return evalScalarBinary(left, right, e.op)
	case (value.IsArray(left) || value.IsObject(left)) && value.IsScalar(right):
		return evalScalarArrayBinary(right, left, e.op)
	case value.IsArray(left) && value.IsArray(right):
		return evalArrayBinary(left, right, e.op)
	case value.IsObject(left) && value.IsObject(right):
		return evalViewBinary(left, right, e.op)
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

func evalScriptUnary(eg *Engine, e unary, ctx *EngineContext) (value.Value, error) {
	val, err := eg.exec(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	n, err := value.CastToFloat(val)
	switch e.op {
	case op.Add:
		return n, nil
	case op.Sub:
		return value.Float(float64(-n)), nil
	default:
		return value.ErrValue, nil
	}
}

func evalQualifiedAssignment(eg *Engine, expr qualifiedCellAddr, ctx *EngineContext) (LValue, io.Closer, error) {
	var (
		lv  LValue
		cl  io.Closer
		err error
	)
	switch expr := expr.path.(type) {
	case identifier:
		cl, err = ctx.PushMutable(expr.name)
	case access:
		val, err1 := eg.exec(expr, ctx)
		if err1 != nil {
			err = err1
			break
		}
		cl, err = ctx.PushValue(val, expr.prop)
	default:
		err = fmt.Errorf("expression can not be assigned to %s", expr)
	}
	if err != nil {
		return nil, nil, err
	}
	lv, err = resolveQualified(ctx, expr.addr)
	return lv, cl, err
}

func evalAssignmentTarget(eg *Engine, expr Expr, ctx *EngineContext) (LValue, io.Closer, error) {
	var (
		lv  LValue
		cl  io.Closer
		err error
	)
	switch expr := expr.(type) {
	case cellAddr:
		cl, err = ctx.PushMutable("")
		if err != nil {
			break
		}
		lv, err = resolveCell(ctx, expr)
	case rangeAddr:
		cl, err = ctx.PushMutable("")
		if err != nil {
			break
		}
		lv, err = resolveRange(ctx, expr)
	case qualifiedCellAddr:
		return evalQualifiedAssignment(eg, expr, ctx)
	case access:
	case identifier:
		lv, err = resolveIdent(ctx, expr)
	default:
		err = fmt.Errorf("value can not be assigned to %s", expr)
	}
	return lv, cl, err
}

func evalAssignment(eg *Engine, e assignment, ctx *EngineContext) (value.Value, error) {
	lv, cl, err := evalAssignmentTarget(eg, e.ident, ctx)
	if err != nil {
		return nil, err
	}
	if cl != nil {
		defer cl.Close()
	}
	value, err := eg.execAndNormalize(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	return nil, lv.Set(value)
}

func evalImport(eg *Engine, e importFile, ctx *EngineContext) (value.Value, error) {
	file, err := eg.Loader.Open(e.file)
	if err != nil {
		return nil, err
	}
	if e.alias == "" {
		var (
			alias string
			file  = e.file
		)
		for {
			ext := filepath.Ext(file)
			if ext == "" {
				break
			}
			alias = strings.TrimSuffix(file, ext)
		}
		e.alias = alias
	}
	book := types.NewFileValue(file, e.readOnly)
	if ev, ok := ctx.Context().(interface{ Define(string, value.Value) }); ok {
		ev.Define(e.alias, book)
	}
	if e.defaultFile {
		ctx.SetDefault(book)
	}
	return value.Empty(), nil
}

func evalPush(eg *Engine, e push, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalPop(eg *Engine, e pop, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalClear(eg *Engine, e clear, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalLock(eg *Engine, e lockRef, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalUnlock(eg *Engine, e unlockRef, ctx *EngineContext) (value.Value, error) {
	return value.Empty(), nil
}

func evalUse(eg *Engine, e useRef, ctx *EngineContext) (value.Value, error) {
	v, err := ctx.Resolve(e.ident)
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

func evalPrint(eg *Engine, e printRef, ctx *EngineContext) (value.Value, error) {
	v, err := eg.execAndNormalize(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	ctx.config.Printer().Print(v)
	return value.Empty(), nil
}

func evalAccess(eg *Engine, e access, ctx *EngineContext) (value.Value, error) {
	obj, err := eg.exec(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	g, ok := obj.(value.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("object expected")
	}
	return g.Get(e.prop)
}

func evalBinary(e binary, ctx value.Context) (value.Value, error) {
	left, err := Eval(e.left, ctx)
	if err != nil {
		return nil, err
	}
	right, err := Eval(e.right, ctx)
	if err != nil {
		return nil, err
	}

	switch e.op {
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

func evalUnary(e unary, ctx value.Context) (value.Value, error) {
	val, err := Eval(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	n, ok := val.(value.Float)
	if !ok {
		return value.ErrValue, nil
	}
	switch e.op {
	case op.Add:
		return n, nil
	case op.Sub:
		return value.Float(float64(-n)), nil
	default:
		return value.ErrValue, nil
	}
}

func evalCall(e call, ctx value.Context) (value.Value, error) {
	id, ok := e.ident.(identifier)
	if !ok {
		return value.ErrName, nil
	}
	var args []value.Arg
	for i := range e.args {
		args = append(args, makeArg(e.args[i]))
	}
	fn, err := ctx.Resolve(id.name)
	if err != nil {
		return nil, err
	}
	if fn.Kind() != value.KindFunction {
		return nil, fmt.Errorf("%s: %w", id.name, ErrCallable)
	}
	call, ok := fn.(value.FunctionValue)
	return call.Call(args, ctx)
}

func evalCellAddr(e cellAddr, ctx value.Context) (value.Value, error) {
	return ctx.At(e.Position)
}

func evalRangeAddr(e rangeAddr, ctx value.Context) (value.Value, error) {
	return ctx.Range(e.startAddr.Position, e.endAddr.Position)
}

type arg struct {
	expr Expr
}

func makeArg(expr Expr) value.Arg {
	return arg{
		expr: expr,
	}
}

func (a arg) Eval(ctx value.Context) (value.Value, error) {
	return Eval(a.expr, ctx)
}
