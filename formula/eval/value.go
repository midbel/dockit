package eval

import (
	"errors"
	"fmt"

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

func resolveQualified(ctx *EngineContext, addr Expr) (LValue, error) {
	var lv LValue
	switch a := addr.(type) {
	case cellAddr:
		lv = cellValue{
			ctx: ctx.Context(),
			pos: a.Position,
		}
	case rangeAddr:
		rg := layout.NewRange(a.startAddr.Position, a.endAddr.Position)
		lv = rangeValue{
			ctx: ctx.Context(),
			rg:  rg.Normalize(),
		}
	default:
		return nil, fmt.Errorf("unknown address type")
	}
	return lv, nil
}

func resolveRange(ctx *EngineContext, rg rangeAddr) (LValue, error) {
	r := layout.NewRange(rg.startAddr.Position, rg.endAddr.Position)
	val := rangeValue{
		rg:  r.Normalize(),
		ctx: ctx.Context(),
	}
	return val, nil
}

func resolveCell(ctx *EngineContext, addr cellAddr) (LValue, error) {
	val := cellValue{
		pos: addr.Position,
		ctx: ctx.Context(),
	}
	return val, nil
}

func resolveIdent(ctx *EngineContext, ident identifier) (LValue, error) {
	id := identValue{
		ident: ident.name,
		ctx:   ctx.Context(),
	}
	return id, nil
}

type identValue struct {
	ident string
	ctx   value.Context
}

func (v identValue) Set(val value.Value) error {
	d, ok := v.ctx.(interface{ Define(string, value.Value) })
	if !ok {
		return ErrValue
	}
	d.Define(v.ident, val)
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
	ctx value.Context
	rg  *layout.Range
}

func (v rangeValue) Set(val value.Value) error {
	var err error
	switch val := val.(type) {
	case deferred:
		f := deferredFormula{
			expr: val.expr,
		}
		err = v.setFormula(&f)
	case value.ScalarValue:
		err = v.setScalar(val)
	case value.ArrayValue:
		err = v.setArray(val)
	default:
		return ErrType
	}
	return err
}

func (v rangeValue) setFormula(val value.Formula) error {
	f, ok := v.ctx.(interface {
		SetFormula(layout.Position, value.Formula) error
	})
	if !ok {
		return ErrValue
	}
	for pos := range v.rg.Positions() {
		if err := f.SetFormula(pos, val); err != nil {
			return err
		}
	}
	return nil
}

func (v rangeValue) setScalar(val value.ScalarValue) error {
	f, ok := v.ctx.(interface {
		SetValue(layout.Position, value.Value) error
	})
	if !ok {
		return ErrValue
	}
	for pos := range v.rg.Positions() {
		if err := f.SetValue(pos, val); err != nil {
			return err
		}
	}
	return nil
}

func (v rangeValue) setArray(arr value.ArrayValue) error {
	f, ok := v.ctx.(interface {
		SetValue(layout.Position, value.Value) error
	})
	if !ok {
		return ErrValue
	}
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
		if err := f.SetValue(pos, val); err != nil {
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
	ctx value.Context
	pos layout.Position
}

func (v cellValue) Set(val value.Value) error {
	switch val := val.(type) {
	case deferred:
		return v.setFormula(val)
	case value.ScalarValue:
		return v.setValue(val)
	default:
		return ErrValue
	}
}

func (v cellValue) setFormula(val deferred) error {
	f, ok := v.ctx.(interface {
		SetFormula(layout.Position, value.Formula) error
	})
	if !ok {
		return ErrValue
	}
	df := deferredFormula{
		expr: val.expr,
	}
	return f.SetFormula(v.pos, &df)
}

func (v cellValue) setValue(val value.ScalarValue) error {
	f, ok := v.ctx.(interface {
		SetValue(layout.Position, value.Value) error
	})
	if !ok {
		return ErrValue
	}
	return f.SetValue(v.pos, val)
}
