package builtins

import (
	"strings"
	"unicode"

	"github.com/midbel/dockit/grid/format"
	"github.com/midbel/dockit/value"
)

func IsText(args []value.Value) value.Value {
	ok := value.IsText(args[0])
	return value.Boolean(ok)
}

func Concat(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	parts := make([]string, 0, len(args))
	for i := range args {
		t := asString(args[i])
		parts = append(parts, t)
	}
	ret := strings.Join(parts, "")
	return value.Text(ret)
}

func Left(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var chars int
	if len(args) == 2 {
		c := asFloat(args[1])
		if c <= 0 {
			return value.ErrValue
		}
		chars = int(c)
	} else {
		chars++
	}
	str := asString(args[0])
	if chars > len(str) {
		return args[0]
	}
	return value.Text(str[:chars])
}

func Right(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var chars int
	if len(args) == 2 {
		c := asFloat(args[1])
		if c <= 0 {
			return value.ErrValue
		}
		chars = int(c)
	} else {
		chars++
	}
	str := asString(args[0])
	if chars > len(str) {
		return args[0]
	}
	return value.Text(str[len(str)-chars:])
}

func Mid(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ix := asFloat(args[1])
	if ix <= 0 {
		return value.ErrValue
	}
	ix -= 1
	ch := asFloat(args[2])
	if ch <= 0 {
		return value.ErrValue
	}
	str := asString(args[0])
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
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ret := len(asString(args[0]))
	return value.Float(ret)
}

func Upper(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ret := strings.ToUpper(asString(args[0]))
	return value.Text(ret)
}

func Lower(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ret := strings.ToLower(asString(args[0]))
	return value.Text(ret)
}

func Proper(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		str  = asString(args[0])
		last rune
	)

	ret := strings.Map(func(c rune) rune {
		prev := last
		last = c
		if (prev == 0 || unicode.IsPunct(prev) || unicode.IsSpace(prev)) && unicode.IsLetter(c) {
			return unicode.ToUpper(c)
		}
		if unicode.IsLetter(c) {
			return unicode.ToLower(c)
		}
		return c
	}, str)

	return value.Text(ret)
}

func Trim(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ret := strings.TrimSpace(asString(args[0]))
	return value.Text(ret)
}

func Search(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		str    = asString(args[0])
		find   = asString(args[1])
		offset = 1
	)
	if len(args) >= 2 {
		ix, _ := value.CastToFloat(args[2])
		offset = int(ix)
		offset++
	}
	if offset >= len(str) {
		return value.ErrValue
	}

	var (
		in     = strings.ToLower(str)
		needle = strings.ToLower(find)
	)

	ix := strings.Index(in[offset:], needle)
	if ix < 0 {
		return value.ErrValue
	}
	return value.Float(float64(ix + 1))
}

func Find(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		str    = asString(args[0])
		find   = asString(args[1])
		offset = 1
	)
	if len(args) >= 2 {
		ix, _ := value.CastToFloat(args[2])
		offset = int(ix)
		offset++
	}
	if offset >= len(str) {
		return value.ErrValue
	}

	ix := strings.Index(str[offset:], find)
	if ix < 0 {
		return value.ErrValue
	}
	return value.Float(float64(ix + 1))
}

func Replace(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		str  = asString(args[0])
		pos  = asFloat(args[1])
		num  = asFloat(args[2])
		repl = asString(args[3])
	)
	if int(pos) >= len(str) {
		return value.Text(str)
	}
	if int(pos+num) >= len(str) {
		return value.Text(str[:int(pos)+1] + repl)
	}
	str = str[:int(pos)+1] + repl + str[int(pos+num):]
	return value.Text(str)
}

func Substitute(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		str  = asString(args[0])
		old  = asString(args[1])
		repl = asString(args[2])
		num  = asFloat(args[3])
	)
	if num <= 0 {
		return value.ErrValue
	}
	str = strings.Replace(str, old, repl, int(num))
	return value.Text(str)
}

func Text(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		str = asString(args[1])
		ft  format.Formatter
		err error
	)
	switch args[0].Type() {
	case value.TypeNumber:
		ft, err = format.ParseNumberFormatter(str)
	case value.TypeDate:
		ft, err = format.ParseDateFormatter(str)
	default:
		return value.ErrValue
	}
	if err != nil {
		return value.ErrValue
	}
	str, err = ft.Format(args[0])
	if err != nil {
		return value.ErrValue
	}
	return value.Text(str)
}

func Value(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	v, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	return v
}

func Textjoin(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		delim  = asString(args[0])
		ignore = asBool(args[1])
		parts  = make([]string, 0, len(args))
	)
	for i := 2; i < len(args); i++ {
		str := asString(args[i])
		if str == "" && ignore {
			continue
		}
		parts = append(parts, str)
	}
	str := strings.Join(parts, delim)
	return value.Text(str)
}

func Exact(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		s1 = asString(args[0])
		s2 = asString(args[1])
		x  = strings.Compare(s1, s2)
	)
	return value.Boolean(x == 0)
}

func Rept(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		str = asString(args[0])
		num = asFloat(args[1])
	)
	if float64(num) <= 0 {
		return value.Text("")
	}
	str = strings.Repeat(str, int(num))
	return value.Text(str)
}

func Clean(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		str = asString(args[0])
		all = make([]byte, 0, len(str))
	)
	for i := 0; i < len(str); i++ {
		if str[i] <= 31 {
			continue
		}
		all = append(all, str[i])
	}
	return value.Text(string(all))
}
