package eval

import (
	"fmt"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
	"github.com/midbel/dockit/formula/types"
)

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

func resolveValueFromAddr(view grid.View, addr Expr) (value.Value, error) {
	switch e := addr.(type) {
	case cellAddr:
		cell, err := view.Cell(e.Position)
		if err != nil {
			return types.ErrValue, err
		}
		return cell.Value(), nil
	case rangeAddr:
		rg := types.NewRangeValue(e.startAddr.Position, e.endAddr.Position)
		return rg.(*types.Range).Collect(view)
	default:
		return types.ErrValue, nil
	}
}