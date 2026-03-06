package builtins

import (
	"strings"

	"github.com/midbel/dockit/value"
)

func IsText(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	ok := value.IsText(args[0])
	return value.Boolean(ok), nil
}

func Concat(args []value.Value) (value.Value, error) {
	if len(args) < 2 {
		return value.ErrValue, ErrArity
	}
	parts := make([]string, 0, len(args))
	for i := range args {
		t, err := value.CastToText(args[i])
		if err != nil {
			return value.ErrValue, err
		}
		parts = append(parts, string(t))
	}
	ret := strings.Join(parts, "")
	return value.Text(ret), nil
}

func Left(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Right(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Mid(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Len(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	t, err := value.CastToText(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	ret := len(t)
	return value.Float(ret), nil
}

func Upper(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	t, err := value.CastToText(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	ret := strings.ToUpper(string(t))
	return value.Text(ret), nil
}

func Lower(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	t, err := value.CastToText(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	ret := strings.ToLower(string(t))
	return value.Text(ret), nil
}

func Substr(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Replace(args []value.Value) (value.Value, error) {
	return nil, nil
}
