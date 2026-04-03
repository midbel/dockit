package eval

import (
	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type SpecialForm interface {
	Run(Runnable, []parse.Expr, *EngineContext) (value.Value, error)
}

type SpecialFunction func(Runnable, []parse.Expr, *EngineContext) (value.Value, error)

var specials = map[string]SpecialForm{
	"inspect": inspectForm{},
	"kindof":  kindofForm{},
}

type inspectForm struct{}

func (i inspectForm) Run(eg Runnable, args []parse.Expr, ctx *EngineContext) (value.Value, error) {
	if len(args) == 0 {
		return value.ErrValue, nil
	}
	switch a := args[0].(type) {
	case parse.CellAddr:
		return i.inspectCell(eg, a, ctx)
	case parse.RangeAddr:
		return i.inspectRange(eg, a, ctx)
	case parse.Slice:
		return i.inspectSlice(eg, a, ctx)
	case parse.Identifier:
		return i.inspectIdent(eg, a, ctx)
	case parse.Number:
		return i.inspectNumber(eg, a, ctx)
	case parse.Literal:
		return i.inspectLiteral(eg, a, ctx)
	default:
		return value.ErrNA, nil
	}
}

func (i inspectForm) inspectCell(eg Runnable, expr parse.CellAddr, ctx *EngineContext) (value.Value, error) {
	iv := types.InspectCell()
	iv.Set("position", value.Text(expr.Position.String()))
	iv.Set("kind", value.Text(expr.KindOf()))

	if view := ctx.CurrentActiveView(); view != nil {
		iv.Set("view", value.Text(view.Type()))
	}

	val, _ := eg.Run(expr)
	if val != nil {
		iv = types.ReinspectValue(iv, val)
	}
	return iv, nil
}

func (i inspectForm) inspectRange(eg Runnable, expr parse.RangeAddr, ctx *EngineContext) (value.Value, error) {
	var (
		iv = types.InspectRange()
		rg = layout.NewRange(expr.StartAt().Position, expr.EndAt().Position)
	)
	iv.Set("start", value.Text(expr.StartAt().Position.String()))
	iv.Set("end", value.Text(expr.EndAt().Position.String()))
	iv.Set("kind", value.Text(expr.KindOf()))

	if view := ctx.CurrentActiveView(); view != nil {
		iv.Set("owner", value.Text(view.Type()))
	}

	rg = rg.Normalize()
	iv.Set("rows", value.Float(rg.Height()))
	iv.Set("cols", value.Float(rg.Width()))

	return iv, nil
}

func (i inspectForm) inspectSlice(eg Runnable, expr parse.Slice, ctx *EngineContext) (value.Value, error) {
	iv := types.InspectSlice()
	if v := expr.View(); v == nil {
		iv.Set("owner", value.Text("view"))
	} else {
		val, err := eg.Run(v)
		if err != nil {
			return value.ErrValue, err
		}
		v, ok := val.(*types.View)
		if !ok {
			return value.ErrValue, nil
		}
		iv.Set("owner", value.Text(v.Type()))
	}
	switch e := expr.Expr().(type) {
	case parse.RangeAddr:
		iv.Set("type", value.Text("range"))

		rg := layout.NewRange(e.StartAt().Position, e.EndAt().Position)
		iv.Set("start", value.Text(e.StartAt().Position.String()))
		iv.Set("end", value.Text(e.EndAt().Position.String()))

		rg = rg.Normalize()
		iv.Set("rows", value.Float(rg.Height()))
		iv.Set("cols", value.Float(rg.Width()))
	case parse.IntervalList:
		iv.Set("type", value.Text("column"))
		iv.Set("count", value.Float(e.Count()))
	case parse.Binary, parse.And, parse.Or, parse.Not:
		iv.Set("type", value.Text("logical"))
	default:
		iv.Set("type", value.Text("unknown"))
	}
	return iv, nil
}

func (i inspectForm) inspectIdent(eg Runnable, expr parse.Identifier, ctx *EngineContext) (value.Value, error) {
	val := ctx.Resolve(expr.Ident())
	if value.IsError(val) {
		return val, nil
	}
	if i, ok := val.(interface{ Inspect() *types.InspectValue }); ok {
		return i.Inspect(), nil
	}
	return types.InspectPrimitive(), nil
}

func (i inspectForm) inspectNumber(eg Runnable, expr parse.Number, ctx *EngineContext) (value.Value, error) {
	val, err := eg.Run(expr)
	if err != nil {
		return value.ErrValue, err
	}
	iv := types.InspectPrimitive()
	return types.ReinspectValue(iv, val), nil
}

func (i inspectForm) inspectLiteral(eg Runnable, expr parse.Literal, ctx *EngineContext) (value.Value, error) {
	val, err := eg.Run(expr)
	if err != nil {
		return value.ErrValue, err
	}
	iv := types.InspectPrimitive()
	return types.ReinspectValue(iv, val), nil
}

type kindofForm struct{}

func (kindofForm) Run(eg Runnable, args []parse.Expr, ctx *EngineContext) (value.Value, error) {
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
