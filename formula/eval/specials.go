package eval

import (
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type SpecialForm interface {
	Eval(*Engine, []Expr, *EngineContext) (value.Value, error)
}

type SpecialFunction func(*Engine, []Expr, *EngineContext) (value.Value, error)

var specials = map[string]SpecialForm{
	"inspect": inspectForm{},
	"kindof":  kindofForm{},
}

type inspectForm struct{}

func (i inspectForm) Eval(eg *Engine, args []Expr, ctx *EngineContext) (value.Value, error) {
	if len(args) == 0 {
		return value.ErrValue, nil
	}
	switch a := args[0].(type) {
	case cellAddr:
		return i.inspectCell(eg, a, ctx)
	case rangeAddr:
		return i.inspectRange(eg, a, ctx)
	case qualifiedCellAddr:
		return i.inspectQualified(eg, a, ctx)
	case slice:
		return i.inspectSlice(eg, a, ctx)
	case identifier:
		return i.inspectIdent(eg, a, ctx)
	case number:
		return i.inspectNumber(eg, a, ctx)
	case literal:
		return i.inspectLiteral(eg, a, ctx)
	default:
		return value.ErrNA, nil
	}
}

func (i inspectForm) inspectCell(eg *Engine, expr cellAddr, ctx *EngineContext) (value.Value, error) {
	iv := types.InspectCell()
	iv.Set("position", value.Text(expr.Position.String()))
	iv.Set("kind", value.Text(expr.KindOf()))

	if view := ctx.CurrentActiveView(); view != nil {
		iv.Set("view", value.Text(view.Type()))
	}

	val, _ := eg.exec(expr, ctx)
	if val != nil {
		iv = types.ReinspectValue(iv, val)
	}
	return iv, nil
}

func (i inspectForm) inspectQualified(eg *Engine, expr qualifiedCellAddr, ctx *EngineContext) (value.Value, error) {
	return nil, nil
}

func (i inspectForm) inspectRange(eg *Engine, expr rangeAddr, ctx *EngineContext) (value.Value, error) {
	var (
		iv = types.InspectRange()
		rg = layout.NewRange(expr.startAddr.Position, expr.endAddr.Position)
	)
	iv.Set("start", value.Text(expr.startAddr.Position.String()))
	iv.Set("end", value.Text(expr.endAddr.Position.String()))
	iv.Set("kind", value.Text(expr.KindOf()))

	if view := ctx.CurrentActiveView(); view != nil {
		iv.Set("owner", value.Text(view.Type()))
	}

	rg = rg.Normalize()
	iv.Set("rows", value.Float(rg.Height()))
	iv.Set("cols", value.Float(rg.Width()))

	return iv, nil
}

func (i inspectForm) inspectSlice(eg *Engine, expr slice, ctx *EngineContext) (value.Value, error) {
	iv := types.InspectSlice()
	if expr.view == nil {
		iv.Set("owner", value.Text("view"))
	} else {
		val, err := eg.exec(expr.view, ctx)
		if err != nil {
			return value.ErrValue, err
		}
		v, ok := val.(*types.View)
		if !ok {
			return value.ErrValue, nil
		}
		iv.Set("owner", value.Text(v.Type()))
	}
	switch e := expr.expr.(type) {
	case rangeSlice:
		iv.Set("type", value.Text("range"))

		rg := layout.NewRange(e.startAddr.Position, e.endAddr.Position)
		iv.Set("start", value.Text(e.startAddr.Position.String()))
		iv.Set("end", value.Text(e.endAddr.Position.String()))

		rg = rg.Normalize()
		iv.Set("rows", value.Float(rg.Height()))
		iv.Set("cols", value.Float(rg.Width()))
	case columnsSlice:
		iv.Set("type", value.Text("column"))
		iv.Set("count", value.Float(len(e.columns)))
	case binary, and, or:
		iv.Set("type", value.Text("binary"))
	default:
		iv.Set("type", value.Text("unknown"))
	}
	return iv, nil
}

func (i inspectForm) inspectIdent(eg *Engine, expr identifier, ctx *EngineContext) (value.Value, error) {
	val, err := ctx.Resolve(expr.name)
	if err != nil {
		return value.ErrValue, err
	}
	if i, ok := val.(interface{ Inspect() *types.InspectValue }); ok {
		return i.Inspect(), nil
	}
	return types.InspectPrimitive(), nil
}

func (i inspectForm) inspectNumber(eg *Engine, expr number, ctx *EngineContext) (value.Value, error) {
	val, err := eg.exec(expr, ctx)
	if err != nil {
		return value.ErrValue, err
	}
	iv := types.InspectPrimitive()
	return types.ReinspectValue(iv, val), nil
}

func (i inspectForm) inspectLiteral(eg *Engine, expr literal, ctx *EngineContext) (value.Value, error) {
	val, err := eg.exec(expr, ctx)
	if err != nil {
		return value.ErrValue, err
	}
	iv := types.InspectPrimitive()
	return types.ReinspectValue(iv, val), nil
}

type kindofForm struct{}

func (kindofForm) Eval(eg *Engine, args []Expr, ctx *EngineContext) (value.Value, error) {
	if len(args) == 0 {
		return value.ErrValue, nil
	}
	var name string
	if k, ok := args[0].(interface{ KindOf() string }); ok {
		name = k.KindOf()
	} else {
		name = "unknown"
	}
	return value.Text(name), nil
}
