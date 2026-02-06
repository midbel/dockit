package eval

import (
	"fmt"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
)

func resolveMutableViewFromQualifiedPath(eg *Engine, ctx *env.Environment, path Expr) (grid.MutableView, error) {
	switch p := path.(type) {
	case identifier:
		return getMutableView(ctx, p.name)
	case access:
		val, err := eg.exec(p.expr, ctx)
		if err != nil {
			return nil, err
		}
		file, ok := val.(*types.File)
		if !ok {
			return nil, fmt.Errorf("view can not be resolved from expr")
		}
		val, err = file.Sheet(p.prop)
		if err != nil {
			return nil, err
		}
		v, ok := val.(*types.View)
		if !ok {
			return nil, fmt.Errorf("view can not be resolved from expr")
		}
		return v.Mutable()
	default:
		return nil, fmt.Errorf("view can not be resolved from expr")
	}
}

func resolveViewFromValue(val value.Value) (grid.View, error) {
	var view grid.View
	switch x := val.(type) {
	case *types.File:
		v, err := x.Active()
		if err != nil {
			return nil, err
		}
		view = v.(grid.View)
	case *types.View:
		view = x.View().(grid.View)
	case grid.View:
		view = x
	case grid.MutableView:
		view = x
	default:
		return nil, fmt.Errorf("view can not be resolved from value")
	}
	return view, nil
}

func resolveMutableViewFromValue(val value.Value) (grid.MutableView, error) {
	var view grid.MutableView
	switch x := val.(type) {
	case *types.File:
		v, err := x.Active()
		if err != nil {
			return nil, err
		}
		mv, ok := v.(grid.MutableView)
		if !ok {
			return nil, fmt.Errorf("active sheet is not mutable")
		}
		view = mv
	case *types.View:
		mv, err := x.Mutable()
		if err != nil {
			return nil, err
		}
		view = mv
	case grid.MutableView:
		view = x
	default:
		return nil, fmt.Errorf("mutable view can not be resolved from value")
	}
	return view, nil
}
