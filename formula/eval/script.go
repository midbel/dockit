package eval

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/builtins"
	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	gbs "github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/internal/ds"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type evaluator struct {
	ctx    *EngineContext
	stack  *ds.Stack[value.Value]
	phases *ds.Stack[scriptPhase]
}

func evalScript(ctx *EngineContext) *evaluator {
	return &evaluator{
		ctx:    ctx,
		stack:  ds.NewStack[value.Value](),
		phases: ds.NewStack[scriptPhase](),
	}
}

func (v *evaluator) Run(expr parse.Expr) (value.Value, error) {
	if err := v.visitExpr(expr); err != nil {
		return value.ErrValue, err
	}
	val := v.popValue()
	return v.normalize(val)
}

func (v *evaluator) VisitScript(expr parse.Script) error {
	for i := range expr.Body {
		if err := v.visitExpr(expr.Body[i]); err != nil {
			return err
		}
	}
	return nil
}

func (v *evaluator) VisitUseRef(expr parse.UseRef) error {
	val, err := v.resolve(expr.Identifier())
	if err != nil {
		return err
	}
	switch val.(type) {
	case *types.File:
		v.ctx.SetDefault(val)
	default:
		return fmt.Errorf("only file can be used as default")
	}
	return nil
}

func (v *evaluator) VisitIncludeFile(expr parse.IncludeFile) error {
	return nil
}

func (v *evaluator) VisitLock(expr parse.Lock) error {
	val, err := v.visitNormalize(expr.Ident())
	if err != nil {
		return nil
	}
	lock, ok := val.(interface{ Lock() })
	if !ok {
		return fmt.Errorf("given value can not be locked")
	}
	lock.Lock()
	return nil
}

func (v *evaluator) VisitUnlock(expr parse.Unlock) error {
	val, err := v.visitNormalize(expr.Ident())
	if err != nil {
		return nil
	}
	lock, ok := val.(interface{ Unlock() })
	if !ok {
		return fmt.Errorf("given value can not be unlocked")
	}
	lock.Unlock()
	return nil
}

func (v *evaluator) VisitRename(expr parse.Rename) error {
	val, err := v.visitNormalize(expr.Ident())
	if err != nil {
		return nil
	}
	view, ok := val.(interface{ Rename(string) })
	if !ok {
		return fmt.Errorf("given value can not be renamed")
	}
	view.Rename(expr.Name().String())
	return nil
}

func (v *evaluator) VisitSheet(expr parse.Sheet) error {
	if expr.Name() == nil && expr.Ident() == nil {
		return fmt.Errorf("sheet should have a name")
	}
	var (
		name  value.Value
		ident value.Value
		data  value.Value
		file  value.Value
		err   error
	)
	if n := expr.Name(); n != nil {
		name, err = v.visitNormalize(n)
		if err != nil {
			return err
		}
	} else if n := expr.Ident(); n != nil {
		name, err = v.visitNormalize(n)
		if err != nil {
			return err
		}
	}
	if n := expr.Ident(); n != nil {
		ident, err = v.visitNormalize(n)
		if err != nil {
			return err
		}
		if id, ok := n.(parse.Identifier); ok && ident == value.ErrRef {
			ident = value.Text(id.Ident())
		}
	}
	if d := expr.Data(); d != nil {
		data, err = v.visitNormalize(d)
		if err != nil {
			return err
		}
	} else {
		var (
			opts = expr.Options()
			cols = getInt(opts["columns"])
			rows = getInt(opts["rows"])
		)
		if rows == 0 {
			rows = 1
		}
		if cols == 0 {
			cols = 1
		}
		data = value.ScalarToArray(value.Empty(), rows, cols)

	}
	sheet, err := v.ctx.NewSheet(name, data, file)
	if err != nil {
		return err
	}
	if ident == nil {
		ident = name
	}
	v.ctx.Define(ident.String(), sheet)
	return nil
}

func (v *evaluator) VisitInsert(expr parse.Insert) error {
	var (
		sheet value.Value
		count value.Value
		data  value.Value
		err   error
	)
	if i := expr.Ident(); i != nil {
		sheet, err = v.visitNormalize(i)
		if err != nil {
			return err
		}
	}
	if c := expr.Count(); c != nil {
		count, err = v.visitNormalize(c)
		if err != nil {
			return err
		}
	}
	ix, err := v.resolveTarget(sheet, expr.Target(), expr.Type())
	if err != nil {
		return err
	}
	switch expr.Where() {
	case parse.AnchorBefore:
		ix--
	case parse.AnchorAfter:
	default:
		return fmt.Errorf("invalid anchor for insert statement")
	}
	var wrg *grid.WritableRange
	switch expr.Type() {
	case parse.Column:
		wrg, err = v.ctx.InsertColumns(sheet, count, value.Float(ix))
	case parse.Row:
		wrg, err = v.ctx.InsertRows(sheet, count, value.Float(ix))
	default:
	}
	if err != nil || wrg == nil {
		return err
	}
	if d := expr.Value(); d != nil {
		data, err = v.visitNormalize(d)
		if err != nil {
			return err
		}
		err = wrg.SetRange(data)
	}
	return err
}

func (v *evaluator) VisitRemove(expr parse.Remove) error {
	if k := expr.Target().Kind; k == parse.TargetFirst && expr.Anchor == parse.AnchorBefore {
		return fmt.Errorf("row/column can not be removed before first row/column")
	} else if k == parse.TargetLast && expr.Anchor == parse.AnchorAfter {
		return fmt.Errorf("row/column can not be removed after last row/column")
	}
	var (
		sheet value.Value
		count value.Value
		err   error
	)
	if i := expr.Ident(); i != nil {
		sheet, err = v.visitNormalize(i)
		if err != nil {
			return err
		}
	}
	if c := expr.Count(); c != nil {
		count, err = v.visitNormalize(c)
		if err != nil {
			return err
		}
	}
	ix, err := v.resolveTarget(sheet, expr.Target(), expr.Type())
	if err != nil {
		return err
	}
	switch expr.Where() {
	case parse.AnchorAfter:
		ix++
	case parse.AnchorBefore:
		ix--
	case parse.AnchorAt:
	default:
		return fmt.Errorf("invalid anchor for remove statement")
	}
	var ret value.Value
	switch expr.Type() {
	case parse.Column:
		ret, err = v.ctx.RemoveColumns(sheet, count, value.Float(ix))
	case parse.Row:
		ret, err = v.ctx.RemoveRows(sheet, count, value.Float(ix))
	default:
	}
	v.pushValue(ret)
	return err
}

func (v *evaluator) resolveTarget(source value.Value, target parse.Target, kind parse.Colrow) (int64, error) {
	view, ok := source.(*types.View)
	if !ok {
		return 0, fmt.Errorf("expected view")
	}

	var max int64
	switch kind {
	case parse.Row:
		max = view.Bounds().Height()
	case parse.Column:
		max = view.Bounds().Width()
	}

	var index int64
	switch target.Kind {
	case parse.TargetFirst:
		index = 1
	case parse.TargetLast:
		index = max
	case parse.TargetIndex:
		if target.Expr == nil {
			index = max
			break
		}
		val, err := v.visitNormalize(target.Expr)
		if err != nil {
			return 0, err
		}
		n, ok := val.(value.Float)
		if !ok {
			return 0, fmt.Errorf("target: number expected")
		}
		index = int64(n)
	default:
		return 0, fmt.Errorf("invalid target")
	}
	if index < 0 {
		index = 0
	}
	return index, nil
}

func (v *evaluator) VisitImportFile(expr parse.ImportFile) error {
	options := expr.Options()
	switch spec := expr.Specifier(); expr.Format() {
	case "csv":
		if spec == "" {
			spec = v.ctx.GetOptionString(ConfigImportCsvDelim)
		}
		options["delimiter"] = csvDelimiter(spec)
	case "log":
		if spec == "" {
			spec = v.ctx.GetOptionString(ConfigImportLogPattern)
		}
		options["pattern"] = spec
	case "json":
		options["query"] = spec
	case "xml":
		options["query"] = spec
	default:
	}
	source, err := v.visitNormalize(expr.File())
	if err != nil {
		return err
	}
	name := source.String()
	file, err := v.ctx.Open(name, options)
	if err != nil {
		return err
	}
	alias := expr.Alias()
	if alias == "" {
		name = filepath.Base(name)
		for {
			ext := filepath.Ext(name)
			if ext == "" {
				break
			}
			name = strings.TrimSuffix(name, ext)
		}
		alias = name
	}
	wb := types.NewFileValue(file, expr.ReadOnly())
	v.ctx.Define(alias, wb)
	if expr.Default() {
		v.ctx.SetDefault(wb)
	}
	return nil
}

func (v *evaluator) VisitPrintRef(expr parse.PrintRef) error {
	val, err := v.visitNormalize(expr.Expr())
	if err != nil {
		return err
	}
	return v.ctx.Print(val)
}

func (v *evaluator) VisitExportFile(expr parse.ExportFile) error {
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	val := v.popValue()

	target, err := v.visitNormalize(expr.File())
	if err != nil {
		return err
	}
	return v.ctx.Export(val, target.String(), expr.Format())
}

func (v *evaluator) VisitCellAccess(expr parse.CellAccess) error {
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	val := v.popValue()
	switch x := val.(type) {
	case value.ScalarValue:
		val = types.NewViewValue(NewScalarView(x))
	case value.ArrayValue:
		val = types.NewViewValue(NewArrayView(x))
	default:
	}
	var (
		sub = evalScript(v.ctx.Sub(val))
		err error
	)
	sub.stack = v.stack.Clone()
	sub.phases = v.phases.Clone()

	val, err = sub.Run(expr.Addr())
	if err != nil {
		val = value.ErrValue
	}
	v.pushValue(val)
	return err
}

func (v *evaluator) VisitSpecial(expr parse.SpecialAccess) error {
	var target value.Value
	if src := expr.Object(); src != nil {
		err := v.visitExpr(src)
		if err != nil {
			return err
		}
		target = v.popValue()
	} else {
		target = v.ctx.Default()
	}
	obj, ok := target.(*types.File)
	if !ok {
		return fmt.Errorf("expected file")
	}
	id, ok := expr.Property().(parse.Identifier)
	if !ok {
		return fmt.Errorf("expected identifier")
	}
	val := obj.Get(id.Ident())
	v.pushValue(val)
	return nil
}

func (v *evaluator) VisitAccess(expr parse.Access) error {
	var obj value.Value
	if expr := expr.Object(); expr != nil {
		if err := v.visitExpr(expr); err != nil {
			return err
		}
		obj = v.popValue()
	} else {
		obj = v.ctx.Default()
	}
	obj, ok := obj.(value.ObjectValue)
	if !ok {
		return fmt.Errorf("expected file/view")
	}
	var val value.Value
	switch prop := expr.Property().(type) {
	case parse.Identifier:
		val = evalAccess(obj, prop)
	default:
		return fmt.Errorf("unexpected property type")
	}
	v.pushValue(val)
	return nil
}

func evalAccess(obj value.Value, prop parse.Identifier) value.Value {
	switch obj := obj.(type) {
	case *types.File:
		prop, err := obj.Sheet(prop.Ident())
		if err != nil {
			prop = value.ErrName
		}
		return prop
	case value.ObjectValue:
		return obj.Get(prop.Ident())
	default:
		return value.ErrValue
	}
}

func (v *evaluator) VisitTemplate(expr parse.Template) error {
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

func (v *evaluator) VisitDeferred(expr parse.Deferred) error {
	v.pushValue(expr)
	return nil
}

func (v *evaluator) VisitAssignment(expr parse.Assignment) error {
	v.enterPhase(phaseAssign)
	defer v.leavePhase()

	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	val, err := v.normalize(v.popValue())
	if err != nil {
		return err
	}
	switch e := expr.Ident().(type) {
	case parse.CellAddr:
		err = v.ctx.SetAt(e.Position, val)
	case parse.ColumnAddr:
		var (
			view  = v.ctx.CurrentActiveView()
			bd    = view.Bounds()
			start = layout.NewPosition(bd.Starts.Line, e.Column)
			end   = layout.NewPosition(bd.Ends.Line, e.Column)
		)
		err = v.ctx.SetRange(start, end, val)
	case parse.RangeAddr:
		err = v.ctx.SetRange(e.StartAt().Position, e.EndAt().Position, val)
	case parse.Access:
	case parse.SpecialAccess:
	case parse.Identifier:
		v.ctx.Define(e.Ident(), val)
	default:
		err = fmt.Errorf("target value is not assignable")
	}
	return err
}

func (v *evaluator) VisitAssert(expr parse.Assert) error {
	if err := v.visitExpr(expr.Expr()); err != nil {
		return err
	}
	ok := value.True(v.popValue())
	if !ok {
		mode := expr.Type()
		if mode == parse.AssertUnknown {
			opt := v.ctx.GetOptionString(ConfigAssertMode)
			mode = assertMode(opt)
		}
		switch mode {
		default:
		case parse.AssertFail:
			msg := expr.Failure()
			if msg == "" {
				msg = fmt.Sprintf("assertion failed: %s", expr.Expr())
			}
			return Abort(msg)
		case parse.AssertWarn:
			v.pushValue(value.ErrValue)
		case parse.AssertIgnore:
		}
	}
	return nil
}

func (v *evaluator) VisitBinary(expr parse.Binary) error {
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
	case (value.IsArray(right) || value.IsObject(right)) && value.IsScalar(left):
		val, err = v.evalScalarInArrayBinary(left, right, expr.Op())
	case (value.IsArray(left) || value.IsObject(left)) && value.IsScalar(right):
		val, err = v.evalArrayWithScalarBinary(left, right, expr.Op())
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

func (v *evaluator) evalScalarBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	if err := value.HasErrors(left, right); err != nil {
		return err, nil
	}
	var ret value.Value
	switch oper {
	case op.Add:
		ret = value.Add(left, right)
	case op.Sub:
		ret = value.Sub(left, right)
	case op.Mul:
		ret = value.Mul(left, right)
	case op.Div:
		ret = value.Div(left, right)
	case op.Pow:
		ret = value.Pow(left, right)
	case op.Concat:
		ret = value.Concat(left, right)
	case op.Eq:
		ret = value.Eq(left, right)
	case op.Ne:
		ret = value.Ne(left, right)
	case op.Lt:
		ret = value.Lt(left, right)
	case op.Le:
		ret = value.Le(left, right)
	case op.Gt:
		ret = value.Gt(left, right)
	case op.Ge:
		ret = value.Ge(left, right)
	default:
		ret = value.ErrValue
	}
	return ret, nil
}

func (v *evaluator) evalScalarInArrayBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	if v, ok := right.(interface{ AsArray() value.ArrayValue }); ok {
		right = v.AsArray()
	}
	arr, err := value.CastToArray(right)
	if err != nil {
		return value.ErrValue, nil
	}
	scalar, ok := left.(value.ScalarValue)
	if !ok {
		return value.ErrValue, nil
	}
	return value.ApplyScalarInArray(scalar, arr, func(left, right value.Value) (value.Value, error) {
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
}

func (v *evaluator) evalArrayWithScalarBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	if v, ok := left.(interface{ AsArray() value.ArrayValue }); ok {
		left = v.AsArray()
	}
	arr, err := value.CastToArray(left)
	if err != nil {
		return value.ErrValue, nil
	}
	scalar, ok := right.(value.ScalarValue)
	if !ok {
		return value.ErrValue, nil
	}
	return value.ApplyArrayWithScalar(arr, scalar, func(left, right value.Value) (value.Value, error) {
		ret, err := v.evalScalarBinary(left, right, oper)
		if err != nil {
			return value.ErrValue, err
		}
		if !value.IsScalar(ret) {
			return value.ErrValue, nil
		}
		return ret, nil
	})
}

func (v *evaluator) evalArrayBinary(left, right value.Value, oper op.Op) (value.Value, error) {
	left, _ = v.normalize(left)
	right, _ = v.normalize(right)

	larr, err := value.CastToArray(left)
	if err != nil {
		return value.ErrValue, nil
	}
	rarr, err := value.CastToArray(right)
	if err != nil {
		return value.ErrValue, nil
	}
	res := larr.ApplyArray(rarr, func(left, right value.Value) value.Value {
		ret, err := v.evalScalarBinary(left, right, oper)
		if err != nil {
			return value.ErrValue
		}
		if !value.IsScalar(ret) {
			return value.ErrValue
		}
		return ret
	})
	return res, nil
}

func (v *evaluator) evalViewBinary(left, right value.Value, oper op.Op) (value.Value, error) {
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

func (v *evaluator) VisitNot(expr parse.Not) error {
	val, err := v.visitNormalize(expr.Expr())
	if err != nil {
		v.pushValue(value.ErrValue)
		return err
	}
	ok := value.True(val)
	v.pushValue(value.Boolean(!ok))
	return nil
}

func (v *evaluator) VisitAnd(expr parse.And) error {
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

func (v *evaluator) VisitOr(expr parse.Or) error {
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

func (v *evaluator) VisitPostfix(expr parse.Postfix) error {
	val, err := v.visitNormalize(expr.Expr())
	if err != nil {
		return err
	}
	switch expr.Op() {
	case op.Percent:
		val = value.Div(val, value.Float(100))
	default:
		val = value.ErrValue
	}
	if err != nil {
		val = value.ErrValue
	}
	v.pushValue(val)
	return nil
}

func (v *evaluator) VisitUnary(expr parse.Unary) error {
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

func (v *evaluator) VisitArray(expr parse.Array) error {
	return nil
}

func (v *evaluator) VisitLiteral(expr parse.Literal) error {
	val := value.Text(expr.Text())
	v.pushValue(val)
	return nil
}

func (v *evaluator) VisitNumber(expr parse.Number) error {
	val := value.Float(expr.Float())
	v.pushValue(val)
	return nil
}

func (v *evaluator) VisitCall(expr parse.Call) error {
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
	fn, err := builtins.Lookup(id.Ident())
	if err != nil {
		return fmt.Errorf("%s: builtin undefined", id.Ident())
	}
	if ok := expr.Vectorizable(); ok {
		return v.vectorizeCall(fn, expr.Args())
	}
	var args []value.Value
	for _, a := range expr.Args() {
		arg, err := v.visitNormalize(a)
		if err != nil {
			return err
		}
		args = append(args, arg)
	}
	val := fn(args)
	v.pushValue(val)
	return nil
}

func (v *evaluator) vectorizeCall(fn gbs.BuiltinFunc, args []parse.Expr) error {
	var (
		count  int
		values []value.Value
	)
	for i, e := range args {
		x, err := v.visitNormalize(e)
		if err != nil {
			return err
		}
		values = append(values, x)
		var vector bool
		if v, ok := args[i].(interface{ Vectorizable() bool }); ok {
			vector = v.Vectorizable()
		}
		if a, ok := x.(value.ArrayValue); ok && count == 0 && vector {
			d := a.Dimension()
			count = int(d.Lines)
		}
	}
	arr := make([][]value.Value, count)
	for i := 0; i < count; i++ {
		var params []value.Value
		for j := range args {
			if a, ok := values[j].(value.ArrayValue); ok {
				params = append(params, a.At(i, 0))
			} else {
				params = append(params, values[j])
			}
		}
		res := fn(params)
		if value.IsScalar(res) {
			arr[i] = slx.One(res)
		}
	}
	v.pushValue(value.NewArray(arr))
	return nil
}

func (v *evaluator) VisitSlice(expr parse.Slice) error {
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
	case parse.IntervalList:
		sel, err := e.Selection()
		if err != nil {
			return err
		}
		view = view.ProjectView(sel)
	case parse.Binary, parse.And, parse.Or, parse.Not:
		p := types.NewExprPredicate(grid.NewFormula(e))
		view = view.FilterView(p)
	case parse.Identifier:
	default:
		return fmt.Errorf("invalid slice expression")
	}
	v.pushValue(view)
	return nil
}

func (v *evaluator) VisitIdentifier(expr parse.Identifier) error {
	val, _ := v.resolve(expr.Ident())
	v.pushValue(val)
	return nil
}

func (v *evaluator) VisitAliasRef(expr parse.AliasRef) error {
	return v.visitExpr(expr.Target())
}

func (v *evaluator) VisitColumnAddr(expr parse.ColumnAddr) error {
	var (
		view  = v.ctx.CurrentActiveView()
		bd    = view.Bounds()
		start = layout.NewPosition(bd.Starts.Line, expr.Column)
		end   = layout.NewPosition(bd.Ends.Line, expr.Column)
		rg    = types.NewRangeValue(start, end)
	)
	v.pushValue(rg)
	return nil
}

func (v *evaluator) VisitCellAddr(expr parse.CellAddr) error {
	val := v.ctx.At(expr.Position)
	v.pushValue(val)
	return nil
}

func (v *evaluator) VisitRangeAddr(expr parse.RangeAddr) error {
	rg := types.NewRangeValue(expr.StartAt().Position, expr.EndAt().Position)
	v.pushValue(rg)
	return nil
}

func (v *evaluator) visitExpr(expr parse.Expr) error {
	a, ok := expr.(parse.VisitableExpr)
	if !ok {
		return ErrEval
	}
	return a.Accept(v)
}

func (v *evaluator) visitNormalize(expr parse.Expr) (value.Value, error) {
	if err := v.visitExpr(expr); err != nil {
		return value.ErrValue, err
	}
	return v.normalize(v.popValue())
}

func (v *evaluator) normalize(val value.Value) (value.Value, error) {
	switch val := val.(type) {
	case *types.Range:
		rg := val.Range()
		return v.ctx.Range(rg.Starts, rg.Ends), nil
	case *types.View, *types.File:
		return val, nil
	default:
		if a, ok := val.(interface{ AsArray() value.ArrayValue }); ok {
			return a.AsArray(), nil
		}
		return val, nil
	}
}

func (v *evaluator) resolve(ident string) (value.Value, error) {
	return v.ctx.Resolve(ident), nil
}

func (v *evaluator) pushValue(val value.Value) {
	v.stack.Push(val)
}

func (v *evaluator) popValue() value.Value {
	val, ok := v.stack.Pop()
	if !ok {
		return value.ErrValue
	}
	return val
}

func (v *evaluator) enterPhase(ph scriptPhase) {
	v.phases.Push(ph)
}

func (v *evaluator) leavePhase() {
	v.phases.Pop()
}

func getInt(value any) int {
	if i, ok := value.(int); ok {
		return i
	}
	if s, ok := value.(string); ok {
		i, err := strconv.Atoi(s)
		if err == nil {
			return i
		}
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return int(f)
		}
	}
	return 0
}
