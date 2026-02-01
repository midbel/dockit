package eval

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
		return types.Text(e.value), nil
	case number:
		return types.Float(e.value), nil
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

type Loader interface {
	Open(string) (grid.File, error)
}

type noopLoader struct{}

func (noopLoader) Open(_ string) (grid.File, error) {
	return nil, fmt.Errorf("noop loader can not open file")
}

func defaultLoader() Loader {
	return noopLoader{}
}

type Engine struct {
	Loader
	Stdout io.Writer
	Stderr io.Writer
}

func NewEngine(loader Loader) *Engine {
	e := Engine{
		Loader: loader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return &e
}

func (e *Engine) Exec(r io.Reader, ctx *env.Environment) (value.Value, error) {
	var (
		val   value.Value
		phase = phaseImport
		ps    = NewParser(ScriptGrammar())
	)
	if err := ps.Init(r); err != nil {
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

func (e *Engine) exec(expr Expr, ctx *env.Environment) (value.Value, error) {
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
		return evalAssignment(e, expr, ctx)
	case access:
		return evalAccess(e, expr, ctx)
	case literal:
		return types.Text(expr.value), nil
	case template:
		return evalTemplate(e, expr, ctx)
	case number:
		return types.Float(expr.value), nil
	case identifier:
		return ctx.Resolve(expr.name)
	case binary:
		return evalScriptBinary(e, expr, ctx)
	case unary:
		return evalScriptUnary(e, expr, ctx)
	case lambda:
		return nil, nil
	case call:
		return nil, nil
	case qualifiedCellAddr:
		return evalQualifiedCell(e, expr, ctx)
	case cellAddr:
		return evalCell(e, expr, ctx)
	case rangeAddr:
		return evalRange(e, expr, ctx)
	default:
		return nil, ErrEval
	}
}

func (e *Engine) execAndNormalize(expr Expr, ctx *env.Environment) (value.Value, error) {
	val, err := e.exec(expr, ctx)
	if err != nil {
		return types.ErrValue, err
	}
	return e.normalizeValue(ctx, val)
}

func (e *Engine) normalizeValue(ctx *env.Environment, val value.Value) (value.Value, error) {
	switch val := val.(type) {
	case *types.Range:
		view, err := getView(ctx, val.Target())
		if err != nil {
			return types.ErrValue, err
		}
		return val.Collect(view)
	default:
		return val, nil
	}
}

func evalRange(eg *Engine, expr rangeAddr, ctx *env.Environment) (value.Value, error) {
	rg := types.NewRangeValue(expr.startAddr.Position, expr.endAddr.Position)
	return rg, nil
}

func evalQualifiedCell(eg *Engine, expr qualifiedCellAddr, ctx *env.Environment) (value.Value, error) {
	return types.Empty(), nil
}

func evalCell(eg *Engine, expr cellAddr, ctx *env.Environment) (value.Value, error) {
	view, err := getView(ctx, expr.Sheet)
	if err != nil {
		return nil, err
	}
	cell, err := view.Cell(expr.Position)
	if err != nil {
		return nil, err
	}
	return cell.Value(), nil
}

func evalTemplate(eg *Engine, expr template, ctx *env.Environment) (value.Value, error) {
	var str strings.Builder
	for i := range expr.expr {
		v, err := eg.exec(expr.expr[i], ctx)
		if err != nil {
			return nil, err
		}
		str.WriteString(v.String())
	}
	return types.Text(str.String()), nil
}

func evalScriptBinary(eg *Engine, e binary, ctx *env.Environment) (value.Value, error) {
	left, err := eg.execAndNormalize(e.left, ctx)
	if err != nil {
		return nil, err
	}
	right, err := eg.execAndNormalize(e.right, ctx)
	if err != nil {
		return nil, err
	}
	switch {
	case types.IsScalar(left) && types.IsScalar(right):
		return evalScalarBinary(left, right, e.op)
	case types.IsArray(left) && types.IsScalar(right):
		return evalScalarArrayBinary(right, left, e.op)
	case types.IsArray(left) && types.IsArray(right):
		return evalArrayBinary(left, right, e.op)
	case types.IsObject(left) && types.IsObject(right):
		return evalObjectBinary(left, right, e.op)
	default:
		return types.ErrValue, nil
	}
}

func evalScalarBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	switch oper {
	case op.Add:
		return types.Add(left, right)
	case op.Sub:
		return types.Sub(left, right)
	case op.Mul:
		return types.Mul(left, right)
	case op.Div:
		return types.Div(left, right)
	case op.Pow:
		return types.Pow(left, right)
	case op.Concat:
		return types.Concat(left, right)
	case op.Eq:
		return types.Eq(left, right)
	case op.Ne:
		return types.Ne(left, right)
	case op.Lt:
		return types.Lt(left, right)
	case op.Le:
		return types.Le(left, right)
	case op.Gt:
		return types.Gt(left, right)
	case op.Ge:
		return types.Ge(left, right)
	default:
		return types.ErrValue, nil
	}
}

func evalScalarArrayBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	arr, err := types.CastToArray(right)
	if err != nil {
		return types.ErrValue, nil
	}
	err = arr.Apply(func(val value.ScalarValue) (value.ScalarValue, error) {
		ret, err := evalScalarBinary(left, val, oper)
		if err != nil {
			return types.ErrValue, err
		}
		scalar, ok := ret.(value.ScalarValue)
		if !ok {
			return types.ErrValue, nil
		}
		return scalar, nil
	})
	if err != nil {
		return types.ErrValue, nil
	}
	return arr, nil
}

func evalArrayBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	larr, err := types.CastToArray(left)
	if err != nil {
		return types.ErrValue, nil
	}
	rarr, err := types.CastToArray(right)
	if err != nil {
		return types.ErrValue, nil
	}
	res, err := larr.ApplyArray(rarr, func(left, right value.ScalarValue) (value.ScalarValue, error) {
		ret, err := evalScalarBinary(left, right, oper)
		if err != nil {
			return types.ErrValue, err
		}
		scalar, ok := ret.(value.ScalarValue)
		if !ok {
			return types.ErrValue, nil
		}
		return scalar, nil
	})
	if err != nil {
		return types.ErrValue, err
	}
	return res, nil
}

func evalObjectBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	switch oper {
	case op.Union:
	case op.Concat:
	default:
	}
	return nil, nil
}

func evalScriptUnary(eg *Engine, e unary, ctx *env.Environment) (value.Value, error) {
	val, err := eg.exec(e.right, ctx)
	if err != nil {
		return nil, err
	}
	n, err := types.CastToFloat(val)
	switch e.op {
	case op.Add:
		return n, nil
	case op.Sub:
		return types.Float(float64(-n)), nil
	default:
		return types.ErrValue, nil
	}
}

func evalAssignment(eg *Engine, e assignment, ctx *env.Environment) (value.Value, error) {
	var (
		lv  LValue
		err error
	)
	switch id := e.ident.(type) {
	case qualifiedCellAddr:
		return nil, nil
	case cellAddr:
		lv, err = resolveCell(ctx, id)
	case rangeAddr:
		lv, err = resolveRange(ctx, id)
	case identifier:
		lv, err = resolveIdent(ctx, id)
	default:
		err = fmt.Errorf("value can not be assigned to %s", e.expr)
	}
	if err != nil {
		return nil, err
	}
	value, err := eg.execAndNormalize(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	return nil, lv.Set(value)
}

func evalImport(eg *Engine, e importFile, ctx *env.Environment) (value.Value, error) {
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
	ctx.Define(e.alias, book)
	if e.defaultFile {
		ctx.SetDefault(book)
	}
	return types.Empty(), nil
}

func evalLock(eg *Engine, e lockRef, ctx *env.Environment) (value.Value, error) {
	return types.Empty(), nil
}

func evalUnlock(eg *Engine, e unlockRef, ctx *env.Environment) (value.Value, error) {
	return types.Empty(), nil
}

func evalUse(eg *Engine, e useRef, ctx *env.Environment) (value.Value, error) {
	v, err := ctx.Resolve(e.ident)
	if err != nil {
		return nil, err
	}
	wb, ok := v.(*types.File)
	if !ok {
		return nil, fmt.Errorf("default can only be used with workbook")
	}
	ctx.SetDefault(wb)
	return types.Empty(), nil
}

func evalPrint(eg *Engine, e printRef, ctx *env.Environment) (value.Value, error) {
	v, err := eg.exec(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	if v != nil {
		fmt.Fprintln(eg.Stdout, v.String())
	}
	return types.Empty(), nil
}

func evalAccess(eg *Engine, e access, ctx *env.Environment) (value.Value, error) {
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
		return types.Add(left, right)
	case op.Sub:
		return types.Sub(left, right)
	case op.Mul:
		return types.Mul(left, right)
	case op.Div:
		return types.Div(left, right)
	case op.Pow:
		return types.Pow(left, right)
	case op.Concat:
		return types.Concat(left, right)
	case op.Eq:
		return types.Eq(left, right)
	case op.Ne:
		return types.Ne(left, right)
	case op.Lt:
		return types.Lt(left, right)
	case op.Le:
		return types.Le(left, right)
	case op.Gt:
		return types.Gt(left, right)
	case op.Ge:
		return types.Ge(left, right)
	default:
		return types.ErrValue, nil
	}
}

func evalUnary(e unary, ctx value.Context) (value.Value, error) {
	val, err := Eval(e.right, ctx)
	if err != nil {
		return nil, err
	}
	n, ok := val.(types.Float)
	if !ok {
		return types.ErrValue, nil
	}
	switch e.op {
	case op.Add:
		return n, nil
	case op.Sub:
		return types.Float(float64(-n)), nil
	default:
		return types.ErrValue, nil
	}
}

func evalCall(e call, ctx value.Context) (value.Value, error) {
	id, ok := e.ident.(identifier)
	if !ok {
		return types.ErrName, nil
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

func (a arg) asFilter(ctx value.Context) (*value.Filter, bool, error) {
	var src value.Filter

	b, ok := a.expr.(binary)
	if !ok {
		return nil, false, fmt.Errorf("argument can not be used as a predicate")
	}
	v, err := Eval(b.right, ctx)
	if err != nil {
		return nil, false, err
	}
	src.Predicate, err = types.NewPredicate(b.op, v)
	if err != nil {
		return nil, false, err
	}
	src.Value, err = Eval(b.left, ctx)
	if err != nil {
		return nil, false, err
	}
	return &src, true, err
}
