package builtins

import (
	"strings"

	"github.com/midbel/dockit/value"
)

func IsText(args []value.Value) value.Value {
	ok := value.IsText(args[0])
	return value.Boolean(ok)
}

func Concat(args []value.Value) value.Value {
	parts := make([]string, 0, len(args))
	for i := range args {
		t, err := value.CastToText(args[i])
		if err != nil {
			return value.ErrValue
		}
		parts = append(parts, string(t))
	}
	ret := strings.Join(parts, "")
	return value.Text(ret)
}

func Left(args []value.Value) value.Value {
	return nil
}

func Right(args []value.Value) value.Value {
	return nil
}

func Mid(args []value.Value) value.Value {
	return nil
}

func Len(args []value.Value) value.Value {
	t, err := value.CastToText(args[0])
	if err != nil {
		return value.ErrValue
	}
	ret := len(t)
	return value.Float(ret)
}

func Upper(args []value.Value) value.Value {
	t, err := value.CastToText(args[0])
	if err != nil {
		return value.ErrValue
	}
	ret := strings.ToUpper(string(t))
	return value.Text(ret)
}

func Lower(args []value.Value) value.Value {
	t, err := value.CastToText(args[0])
	if err != nil {
		return value.ErrValue
	}
	ret := strings.ToLower(string(t))
	return value.Text(ret)
}

func Substr(args []value.Value) value.Value {
	return nil
}

func Replace(args []value.Value) value.Value {
	return nil
}
