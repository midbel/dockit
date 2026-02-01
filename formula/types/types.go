package types

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

func IsComparable(v value.Value) bool {
	_, ok := v.(value.Comparable)
	return ok
}

func IsNumber(v value.Value) bool {
	_, ok := v.(Float)
	return ok
}

func IsScalar(v value.Value) bool {
	return v.Kind() == value.KindScalar
}

func IsObject(v value.Value) bool {
	return v.Kind() == value.KindObject
}

func IsArray(v value.Value) bool {
	return v.Kind() == value.KindArray
}

var ErrCompatible = errors.New("incompatible type")

type ErrorCode string

var (
	ErrNull  = createError("#NULL!")
	ErrDiv0  = createError("#DIV/0!")
	ErrValue = createError("#VALUE!")
	ErrRef   = createError("#REF!")
	ErrName  = createError("#NAME?")
	ErrNum   = createError("#NUM!")
	ErrNA    = createError("#N/A")
)

type Error struct {
	code string
}

func createError(code string) Error {
	return Error{
		code: code,
	}
}

func (Error) Type() string {
	return "error"
}

func (Error) Kind() value.ValueKind {
	return value.KindError
}

func (e Error) Error() string {
	return e.code
}

func (e Error) String() string {
	return e.code
}

func (e Error) Scalar() any {
	return e.code
}

type Array struct {
	Data [][]value.ScalarValue
}

func NewArray(data [][]value.ScalarValue) value.ArrayValue {
	return Array{
		Data: data,
	}
}

func (a Array) Type() string {
	dim := a.Dimension()
	return fmt.Sprintf("array(%d, %d)", dim.Lines, dim.Columns)
}

func (Array) Kind() value.ValueKind {
	return value.KindArray
}

func (Array) String() string {
	return ""
}

func (a Array) Dimension() layout.Dimension {
	var (
		d layout.Dimension
		n = len(a.Data)
	)
	if n > 0 {
		d.Lines = int64(n)
		d.Columns = int64(len(a.Data[0]))
	}
	return d
}

func (a Array) At(row, col int) value.ScalarValue {
	if len(a.Data) == 0 || row >= len(a.Data) {
		return nil
	}
	v := a.Data[row]
	if len(v) == 0 || col >= len(v) {
		return nil
	}
	return a.Data[row][col]
}

func (a Array) SetAt(row, col int, val value.ScalarValue) {
	if len(a.Data) == 0 || row >= len(a.Data) {
		return
	}
	v := a.Data[row]
	if len(v) == 0 || col >= len(v) {
		return
	}
	a.Data[row][col] = val
}

func (a Array) Apply(do func(value.ScalarValue) (value.ScalarValue, error)) error {
	if len(a.Data) == 0 {
		return nil
	}
	dim := a.Dimension()
	for i := range dim.Lines {
		for j := range dim.Columns {
			v, err := do(a.At(int(i), int(j)))
			if err != nil {
				return err
			}
			a.SetAt(int(i), int(j), v)
		}
	}
	return nil
}

func (a Array) ApplyArray(other Array, do func(value.ScalarValue, value.ScalarValue) (value.ScalarValue, error)) (value.Value, error) {
	var (
		dleft  = a.Dimension()
		dright = other.Dimension()
		dim    = dleft.Max(dright)
		data   = make([][]value.ScalarValue, dim.Lines)
	)
	for i := range data {
		data[i] = make([]value.ScalarValue, dim.Columns)
	}
	for i := range dim.Lines {
		for j := range dim.Columns {
			var (
				left  = a.At(int(i%dleft.Lines), int(j%dleft.Columns))
				right = other.At(int(i%dright.Lines), int(j%dright.Columns))
			)

			v, err := do(left, right)
			if err != nil {
				return ErrValue, err
			}

			data[i][j] = v
		}
	}
	return NewArray(data), nil
}
