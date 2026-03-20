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
	var chars int
	if len(args) == 2 {
		c, _ := value.CastToFloat(args[1])
		if c <= 0 {
			return value.ErrValue
		}
		chars = int(c)
	} else {
		chars++
	}
	str, _ := value.CastToText(args[0])
	if chars > len(str) {
		return args[0]
	}
	return value.Text(str[:chars])
}

func Right(args []value.Value) value.Value {
	var chars int
	if len(args) == 2 {
		c, _ := value.CastToFloat(args[1])
		if c <= 0 {
			return value.ErrValue
		}
		chars = int(c)
	} else {
		chars++
	}
	str, _ := value.CastToText(args[0])
	if chars > len(str) {
		return args[0]
	}
	return value.Text(str[len(str)-chars:])
}

func Mid(args []value.Value) value.Value {
	ix, _ := value.CastToFloat(args[1])
	if ix <= 0 {
		return value.ErrValue
	}
	ix -= 1
	ch, _ := value.CastToFloat(args[2])
	if ch <= 0 {
		return value.ErrValue
	}
	str, _ := value.CastToText(args[0])
	if int(ix) >= len(str) {
		return args[0]
	}
	str = str[int(ix):]
	if int(ch) >= len(str) {
		return value.Text(str)
	}
	return value.Text(str[:int(ch)])
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

func Trim(args []value.Value) value.Value {
	return nil
}

func Split(args []value.Value) value.Value {
	return nil
}

func Join(args []value.Value) value.Value {
	return nil
}

func Proper(args []value.Value) value.Value {
	return nil
}

func Search(args []value.Value) value.Value {
	return nil
}

func Find(args []value.Value) value.Value {
	return nil
}

func Substitute(args []value.Value) value.Value {
	return nil
}

func Text(args []value.Value) value.Value {
	return nil
}

func Value(args []value.Value) value.Value {
	return nil
}

func Textjoin(args []value.Value) value.Value {
	return nil
}
