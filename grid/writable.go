package grid

import (
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Rangeable interface {
	Range(layout.Position, layout.Position) value.Value
	SetRange(layout.Position, layout.Position, value.Value) error
}

type WritableRange struct {
	view  Rangeable
	start layout.Position
	end   layout.Position
}

func NewWritableRange(view Rangeable, start, end layout.Position) *WritableRange {
	return &WritableRange{
		view:  view,
		start: start,
		end:   end,
	}
}

func (w WritableRange) SetRange(data value.Value) error {
	return w.view.SetRange(w.start, w.end, data)
}
