package eval

import (
	"errors"

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
)

type LValue interface {
	Set(value.Value) error
}

type identValue struct {
	ident string
	ctx   *env.Environment
}

func (v identValue) Set(val value.Value) error {
	v.ctx.Define(v.ident, val)
	return nil
}

func resolveIdent(ctx *env.Environment, ident identifier) (LValue, error) {
	id := identValue{
		ident: ident.name,
		ctx:   ctx,
	}
	return id, nil
}

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

func (v rangeValue) setArray(val value.ArrayValue) error {
	return nil
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
