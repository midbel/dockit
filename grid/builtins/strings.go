package builtins

import (
	"strings"
	"unicode"

	"github.com/midbel/dockit/grid/format"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/value"
)

var concatBuiltin = Builtin{
	Name:     "concatenate",
	Desc:     "",
	Category: "text",
	Alias:    slx.Make("concat"),
	Params: []Param{
		Var(Scalar("str", "", value.TypeText)),
	},
	Func: Concat,
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

var leftBuiltin = Builtin{
	Name:     "left",
	Desc:     "Returns the first characters from text",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Opt(Scalar("chars", "", value.TypeNumber)),
	},
	Func: Left,
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

var rightBuiltin = Builtin{
	Name:     "right",
	Desc:     "Returns the first characters from text",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Opt(Scalar("chars", "", value.TypeNumber)),
	},
	Func: Right,
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

var midBuiltin = Builtin{
	Name:     "mid",
	Desc:     "Returns part of text starting at a given position",
	Alias:    []string{"substr"},
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Scalar("index", "", value.TypeNumber),
		Scalar("chars", "", value.TypeNumber),
	},
	Func: Mid,
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

var lenBuiltin = Builtin{
	Name:     "len",
	Desc:     "Returns the number of characters in text",
	Category: "text",
	Alias:    []string{"length"},
	Params: []Param{
		Scalar("str", "", value.TypeText),
	},
	Func: Len,
}

func Len(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ret := len(asString(args[0]))
	return value.Float(ret)
}

var upperBuiltin = Builtin{
	Name:     "upper",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
	},
	Func: Upper,
}

func Upper(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ret := strings.ToUpper(asString(args[0]))
	return value.Text(ret)
}

var lowerBuiltin = Builtin{
	Name:     "lower",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
	},
	Func: Lower,
}

func Lower(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ret := strings.ToLower(asString(args[0]))
	return value.Text(ret)
}

var properBuiltin = Builtin{
	Name:     "proper",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
	},
	Func: Proper,
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

var trimBuiltin = Builtin{
	Name:     "trim",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
	},
	Func: Trim,
}

func Trim(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	ret := strings.TrimSpace(asString(args[0]))
	return value.Text(ret)
}

var searchBuiltin = Builtin{
	Name:     "search",
	Desc:     "Finds the position of text (case-insensitive). Supports * and ?",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Scalar("find", "", value.TypeText),
		Opt(Scalar("offset", "", value.TypeNumber)),
	},
	Func: Search,
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

var findBuiltin = Builtin{
	Name:     "find",
	Desc:     "Finds the position of text (case-sensitive)",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Scalar("find", "", value.TypeText),
		Opt(Scalar("offset", "", value.TypeNumber)),
	},
	Func: Find,
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

var replaceBuiltin = Builtin{
	Name:     "replace",
	Desc:     "Replace part of text at given position with new text",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Scalar("pos", "", value.TypeNumber),
		Scalar("num", "", value.TypeNumber),
		Scalar("new", "", value.TypeText),
	},
	Func: Replace,
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

var substituteBuiltin = Builtin{
	Name:     "substitue",
	Desc:     "Replaces occurrences of text with new text",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Scalar("old", "", value.TypeText),
		Scalar("new", "", value.TypeText),
		Opt(Scalar("num", "", value.TypeNumber)),
	},
	Func: Substitute,
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

var textBuiltin = Builtin{
	Name:     "text",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Scalar("pattern", "", value.TypeText),
	},
	Func: Text,
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

var valueBuiltin = Builtin{
	Name:     "value",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
	},
	Func: Value,
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

var textjoinBuiltin = Builtin{
	Name:     "textjoin",
	Desc:     "Joins multiple text values using a delimiter",
	Category: "text",
	Params: []Param{
		Scalar("delimiter", "", value.TypeText),
		Scalar("ignore", "", value.TypeBool),
		Var(Scalar("str", "", value.TypeText)),
	},
	Func: Textjoin,
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

var exactBuiltin = Builtin{
	Name:     "exact",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str1", "", value.TypeText),
		Scalar("str2", "", value.TypeText),
	},
	Func: Exact,
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

var reptBuiltin = Builtin{
	Name:     "rept",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
		Scalar("num", "", value.TypeNumber),
	},
	Func: Rept,
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

var cleanBuiltin = Builtin{
	Name:     "clean",
	Desc:     "",
	Category: "text",
	Params: []Param{
		Scalar("str", "", value.TypeText),
	},
	Func: Clean,
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

var textBuiltins = []Builtin{
	concatBuiltin,
	leftBuiltin,
	rightBuiltin,
	midBuiltin,
	lenBuiltin,
	upperBuiltin,
	lowerBuiltin,
	properBuiltin,
	trimBuiltin,
	searchBuiltin,
	findBuiltin,
	replaceBuiltin,
	substituteBuiltin,
	textBuiltin,
	valueBuiltin,
	textjoinBuiltin,
	exactBuiltin,
	reptBuiltin,
	cleanBuiltin,
}
