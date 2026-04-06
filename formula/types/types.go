package types

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

func Rebase(fm value.Formula, anchor, target layout.Position) value.Formula {
	var (
		dy     = target.Line - anchor.Line
		dx     = target.Column - anchor.Column
		offset = layout.NewPosition(dy, dx)
		other = fm
	)
	if c, ok := fm.(interface {
		Clone(layout.Position) value.Formula
	}); ok {
		other = c.Clone(offset)
	}
	return other
}

var (
	ErrValue     = errors.New("invalid value")
	ErrReadOnly  = errors.New("read only view")
	ErrType      = errors.New("invalid type")
	ErrDimension = errors.New("dimension mismatched")
)

type broadcastMode int8

const (
	broadcastExact broadcastMode = 1 << iota
	broadcastRow
	broadcastCol
	broadcastScalar
	broadcastFlat
)

func getBroadcastMode(target *layout.Range, val value.ArrayValue) (broadcastMode, error) {
	var (
		width  = target.Width()
		height = target.Height()
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
		fmt.Println(width, height, dim.Columns, dim.Lines)
		return mode, ErrDimension
	}
	return mode, nil
}
