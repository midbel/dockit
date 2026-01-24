package types

import (
	"fmt"
	"time"

	"github.com/mibel/dockit/value"
)

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

func (d Date) Equal(other value.Value) (bool, error) {
	x, ok := other.(Date)
	if !ok {
		return false, ErrCompatible
	}
	return time.Time(d).Equal(time.Time(x)), nil
}

func (d Date) Less(other value.Value) (bool, error) {
	x, ok := other.(Date)
	if !ok {
		return false, ErrCompatible
	}
	return time.Time(d).Before(time.Time(x)), nil
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

func (f Float) Equal(other value.Value) (bool, error) {
	x, ok := other.(Float)
	if !ok {
		return false, ErrCompatible
	}
	return float64(f) == float64(x), nil
}

func (f Float) Less(other value.Value) (bool, error) {
	x, ok := other.(Float)
	if !ok {
		return false, ErrCompatible
	}
	return float64(f) < float64(x), nil
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

func (t Text) Equal(other value.Value) (bool, error) {
	x, ok := other.(Text)
	if !ok {
		return false, ErrCompatible
	}
	return string(t) == string(x), nil
}

func (t Text) Less(other value.Value) (bool, error) {
	x, ok := other.(Text)
	if !ok {
		return false, ErrCompatible
	}
	return string(t) < string(x), nil
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

func (b Boolean) Equal(other value.Value) (bool, error) {
	x, ok := other.(Boolean)
	if !ok {
		return false, ErrCompatible
	}
	return bool(b) == bool(x), nil
}

func (b Boolean) Less(other value.Value) (bool, error) {
	x, ok := other.(Boolean)
	if !ok {
		return false, ErrCompatible
	}
	return bool(b) && !bool(x), nil
}
