package eval

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var (
	ErrNoDefault = errors.New("no default defined")
	ErrValue     = errors.New("invalid value")
	ErrReadOnly  = errors.New("read only view")
	ErrType      = errors.New("invalid type")
	ErrDimension = errors.New("dimension mismatched")
)

type LValue interface {
	Set(value.Value) error
}

func resolveQualifiedLValue(eg *Engine, ctx *env.Environment, expr qualifiedCellAddr) (LValue, error) {
	view, err := resolveMutableViewFromQualifiedPath(eg, ctx, expr.path)
	if err != nil {
		return nil, err
	}
	return resolveQualified(view, expr.addr)
}

func resolveQualified(view grid.MutableView, addr Expr) (LValue, error) {
	var lv LValue
	switch a := addr.(type) {
	case cellAddr:
		lv = cellValue{
			view: view,
			pos:  a.Position,
		}
	case rangeAddr:
		rg := layout.NewRange(a.startAddr.Position, a.endAddr.Position)
		lv = rangeValue{
			view: view,
			rg:   rg.Normalize(),
		}
	default:
		return nil, fmt.Errorf("unknown address type")
	}
	return lv, nil
}

func resolveRange(ctx *env.Environment, rg rangeAddr) (LValue, error) {
	view, err := getMutableView(ctx, rg.startAddr.Sheet)
	if err != nil {
		return nil, err
	}
	r := layout.NewRange(rg.startAddr.Position, rg.endAddr.Position)
	val := rangeValue{
		rg:   r.Normalize(),
		view: view,
	}
	return val, nil
}

func resolveCell(ctx *env.Environment, addr cellAddr) (LValue, error) {
	view, err := getMutableView(ctx, addr.Sheet)
	if err != nil {
		return nil, err
	}
	val := cellValue{
		pos:  addr.Position,
		view: view,
	}
	return val, nil
}

func resolveIdent(ctx *env.Environment, ident identifier) (LValue, error) {
	id := identValue{
		ident: ident.name,
		ctx:   ctx,
	}
	return id, nil
}

type identValue struct {
	ident string
	ctx   *env.Environment
}

func (v identValue) Set(val value.Value) error {
	v.ctx.Define(v.ident, val)
	return nil
}

type broadcastMode int8

const (
	broadcastExact broadcastMode = 1 << iota
	broadcastRow
	broadcastCol
	broadcastScalar
	broadcastFlat
)

type rangeValue struct {
	view grid.MutableView
	rg   *layout.Range
}

func (v rangeValue) Set(val value.Value) error {
	var err error
	switch val := val.(type) {
	case value.ScalarValue:
		err = v.setScalar(val)
	case value.ArrayValue:
		err = v.setArray(val)
	default:
		return ErrType
	}
	return err
}

func (v rangeValue) setScalar(val value.ScalarValue) error {
	for pos := range v.rg.Positions() {
		if err := v.view.SetValue(pos, val); err != nil {
			return err
		}
	}
	return nil
}

func (v rangeValue) setArray(arr value.ArrayValue) error {
	mode, err := v.mode(arr)
	if err != nil {
		return err
	}
	var (
		index int
		row   int
		col   int
		dim   = arr.Dimension()
	)
	for pos := range v.rg.Positions() {
		var val value.ScalarValue
		switch mode {
		case broadcastExact:
			val = arr.At(row, col)
		case broadcastRow:
			val = arr.At(0, col)
		case broadcastCol:
			val = arr.At(row, 0)
		case broadcastScalar:
			val = arr.At(0, 0)
		case broadcastFlat:
			r := index / int(dim.Lines)
			c := index % int(dim.Columns)
			val = arr.At(r, c)
			index++
		default:
			continue
		}
		if err := v.view.SetValue(pos, val); err != nil {
			return err
		}
		col++
		if col == int(v.rg.Width()) {
			col = 0
			row++
		}
	}
	return nil
}

func (v rangeValue) mode(val value.ArrayValue) (broadcastMode, error) {
	var (
		width  = v.rg.Width()
		height = v.rg.Height()
		dim    = val.Dimension()
		mode   broadcastMode
	)
	switch {
	case width == dim.Columns && height == dim.Lines:
		mode = broadcastExact
	case width == dim.Columns && dim.Lines == 1:
		mode = broadcastRow
	case dim.Columns == 1 && height == dim.Lines:
		mode = broadcastCol
	case dim.Lines == 1 && dim.Columns == 1:
		mode = broadcastScalar
	case dim.Lines*dim.Columns == width*height:
		mode = broadcastFlat
	default:
		return mode, ErrDimension
	}
	return mode, nil
}

type cellValue struct {
	view grid.MutableView
	pos  layout.Position
}

func (v cellValue) Set(val value.Value) error {
	scalar, ok := val.(value.ScalarValue)
	if !ok {
		return ErrValue
	}
	return v.view.SetValue(v.pos, scalar)
}

func getView(ctx *env.Environment, name string) (grid.View, error) {
	sh, err := getSheet(ctx, name)
	if err != nil {
		return nil, err
	}
	return sh.View(), nil
}

func getMutableView(ctx *env.Environment, name string) (grid.MutableView, error) {
	sh, err := getSheet(ctx, name)
	if err != nil {
		return nil, err
	}
	view, err := sh.Mutable()
	if err != nil {
		return nil, ErrReadOnly
	}
	return view, nil
}

func getSheet(ctx *env.Environment, name string) (*types.View, error) {
	obj := ctx.Default()
	if obj == nil {
		return nil, ErrNoDefault
	}
	x, ok := obj.(*types.File)
	if !ok {
		return nil, ErrValue
	}
	var (
		sheet value.Value
		err   error
	)
	if name == "" {
		sheet, err = x.Active()
	} else {
		sheet, err = x.Sheet(name)
	}
	if err != nil {
		return nil, err
	}
	tv, ok := sheet.(*types.View)
	if !ok {
		return nil, ErrValue
	}
	return tv, nil
}
