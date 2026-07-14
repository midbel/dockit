package runtime

import (
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type MutationAxis int8

const (
	Row MutationAxis = 1 << iota
	Column
)

type MutationKind int8

const (
	Insertion MutationKind = 1 << iota
	Removal
)

type Mutation struct {
	Kind   MutationKind
	Axis   MutationAxis
	Count  int64
	Offset int64
}

func RemoveRows(offset, count int64) Mutation {
	return Mutation{
		Kind:   Removal,
		Axis:   Row,
		Count:  count,
		Offset: offset,
	}
}

func RemoveColumns(offset, count int64) Mutation {
	return Mutation{
		Kind:   Removal,
		Axis:   Column,
		Count:  count,
		Offset: offset,
	}
}

func InsertRows(offset, count int64) Mutation {
	return Mutation{
		Kind:   Insertion,
		Axis:   Row,
		Count:  count,
		Offset: offset,
	}
}

func InsertColumns(offset, count int64) Mutation {
	return Mutation{
		Kind:   Insertion,
		Axis:   Column,
		Count:  count,
		Offset: offset,
	}
}

type WritableRange struct {
	view  *View
	start layout.Position
	end   layout.Position
}

func NewWritableRange(view *View, start, end layout.Position) *WritableRange {
	return &WritableRange{
		view:  view,
		start: start,
		end:   end,
	}
}

func (w WritableRange) SetRange(data value.Value) error {
	return w.view.SetRange(w.start, w.end, data)
}
