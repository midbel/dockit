package formula

import (
	"strconv"
	"time"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

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

type Blank struct{}

func (Blank) Kind() value.ValueKind {
	return value.KindScalar
}

func (Blank) String() string {
	return ""
}

func (Blank) Scalar() any {
	return nil
}

type Date time.Time

func (Date) Kind() value.ValueKind {
	return value.KindScalar
}

func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

func (d Date) Scalar() any {
	return time.Time(d)
}

func (d Date) ToString() (value.ScalarValue, error) {
	return Text(d.String()), nil
}

func (d Date) ToBool() (value.ScalarValue, error) {
	return Boolean(!time.Time(d).IsZero()), nil
}

func (d Date) ToFloat() (value.ScalarValue, error) {
	unix := time.Time(d).Unix()
	return Float(float64(unix)), nil
}

type Float float64

func (Float) Kind() value.ValueKind {
	return value.KindScalar
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f Float) Scalar() any {
	return float64(f)
}

func (f Float) ToString() (value.ScalarValue, error) {
	return Text(f.String()), nil
}

func (f Float) ToBool() (value.ScalarValue, error) {
	return Boolean(float64(f) != 0), nil
}

func (f Float) ToFloat() (value.ScalarValue, error) {
	return f, nil
}

type Text string

func (Text) Kind() value.ValueKind {
	return value.KindScalar
}

func (t Text) String() string {
	return string(t)
}

func (t Text) Scalar() any {
	return string(t)
}

func (t Text) ToString() (value.ScalarValue, error) {
	return t, nil
}

func (t Text) ToBool() (value.ScalarValue, error) {
	return Boolean(string(t) != ""), nil
}

func (t Text) ToFloat() (value.ScalarValue, error) {
	n, err := strconv.ParseFloat(string(t), 64)
	if err != nil {
		return ErrNA, nil
	}
	return Float(n), nil
}

type Boolean bool

func (Boolean) Kind() value.ValueKind {
	return value.KindScalar
}

func (b Boolean) String() string {
	return strconv.FormatBool(bool(b))
}

func (b Boolean) Scalar() any {
	return bool(b)
}

func (b Boolean) ToString() (value.ScalarValue, error) {
	s := strconv.FormatBool(bool(b))
	return Text(s), nil
}

func (b Boolean) ToBool() (value.ScalarValue, error) {
	return b, nil
}

func (b Boolean) ToFloat() (value.ScalarValue, error) {
	if !bool(b) {
		return Float(0), nil
	}
	return Float(1), nil
}

type Array struct {
	Data [][]value.ScalarValue
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
