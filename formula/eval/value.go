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
	rg   layout.Range
}

func (v rangeValue) Set(val value.Value) error {
	return nil
}

func resolveRange(ctx *env.Environment, rg rangeAddr) (LValue, error) {
	return nil, nil
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
	if addr.Sheet == "" {
		sheet, err = x.Active()
	} else {
		sheet, err = x.Sheet(addr.Sheet)
	}
	if err != nil {
		return nil, err
	}
	val := cellValue{
		pos: addr.Position,
	}
	if mv, ok := sheet.(*types.View); ok {
		val.view, err = mv.Mutable()
		if err != nil {
			return nil, ErrReadOnly
		}
	} else {
		return nil, ErrValue
	}
	return val, nil
}
