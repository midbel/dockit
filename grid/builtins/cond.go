package builtins

import (
	"math"

	"github.com/midbel/dockit/value"
)

var ifsBuiltin = Builtin{
	Name:     "ifs",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		Var(Scalar("value", "", value.TypeAny)),
	},
	Func: Ifs,
}

func Ifs(args []value.Value) value.Value {
	if err := value.HasErrors(args[:3]...); err != nil {
		return err
	}
	if len(args)%2 == 1 {
		return value.ErrValue
	}
	for i := 0; i < len(args); i += 2 {
		ok := value.True(args[i])
		if ok {
			return args[i+1]
		}
	}
	return value.ErrNA
}

var ifBuiltin = Builtin{
	Name:     "if",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		Scalar("value", "", value.TypeAny),
		Deferrable(Scalar("csq", "", value.TypeAny)),
		Deferrable(Scalar("alt", "", value.TypeAny)),
	},
	Func: If,
}

func If(args []value.Value) value.Value {
	if err := value.HasErrors(args[:3]...); err != nil {
		return err
	}
	if value.True(args[0]) {
		return args[1]
	}
	return args[2]
}

var ifErrorBuiltin = Builtin{
	Name:     "iferror",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		Scalar("value", "", value.TypeAny),
		Deferrable(ScalarArray("replace", "", value.TypeAny)),
	},
	Func: IfError,
}

func IfError(args []value.Value) value.Value {
	if value.IsError(args[0]) {
		return args[1]
	}
	return args[0]
}

var ifNaBuiltin = Builtin{
	Name:     "ifna",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		Scalar("value", "", value.TypeAny),
		Deferrable(ScalarArray("replace", "", value.TypeAny)),
	},
	Func: IfNA,
}

func IfNA(args []value.Value) value.Value {
	if value.IsError(args[0]) && args[0] == value.ErrNA {
		return args[1]
	}
	return args[0]
}

var andBuiltin = Builtin{
	Name:     "and",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		ScalarArray("value1", "", value.TypeAny),
		ScalarArray("value2", "", value.TypeAny),
	},
	Func: And,
}

func And(args []value.Value) value.Value {
	if err := value.HasErrors(args[:2]...); err != nil {
		return err
	}
	var (
		ok1 = value.True(args[0])
		ok2 = value.True(args[1])
	)
	return value.Boolean(ok1 && ok2)
}

var orBuiltin = Builtin{
	Name:     "or",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		ScalarArray("value1", "", value.TypeAny),
		ScalarArray("value2", "", value.TypeAny),
	},
	Func: Or,
}

func Or(args []value.Value) value.Value {
	if err := value.HasErrors(args[:2]...); err != nil {
		return err
	}
	var (
		ok1 = value.True(args[0])
		ok2 = value.True(args[1])
	)
	return value.Boolean(ok1 || ok2)
}

var xorBuiltin = Builtin{
	Name:     "xor",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		Var(ScalarArray("value", "", value.TypeAny)),
	},
	Func: Xor,
}

func Xor(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	for i := range args {
		if !value.True(args[i]) {
			return value.Boolean(true)
		}
	}
	return value.Boolean(false)
}

var notBuiltin = Builtin{
	Name:     "not",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		ScalarArray("value", "", value.TypeAny),
	},
	Func: Not,
}

func Not(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	ok := value.True(args[0])
	return value.Boolean(!ok)
}

var chooseBuiltin = Builtin{
	Name:     "choose",
	Desc:     "Returns the value at the given 1-based index. If the index is out of range, returns ErrNA",
	Category: "conditional",
	Params: []Param{
		Scalar("index", "", value.TypeNumber),
		Deferrable(Var(Scalar("value", "", value.TypeAny))),
	},
	Func: Choose,
}

func Choose(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f := math.Floor(asFloat(args[0])) - 1
	if int(f) < 0 || int(f) >= len(args)-1 {
		return value.ErrNA
	}
	return args[int(f)]
}

var switchBuiltin = Builtin{
	Name:     "switch",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		Scalar("var", "", value.TypeNumber),
		Var(Scalar("value", "", value.TypeAny)),
		Opt(Scalar("default", "", value.TypeAny)),
	},
	Func: Switch,
}

func Switch(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		base     = args[0]
		rest     = args[1:]
		fallback value.Value
	)
	if z := len(rest); z%2 == 1 {
		fallback = rest[z-1]
		rest = rest[:z]
	}
	for i := 0; i < len(rest); i += 2 {
		ok := value.Eq(base, rest[i])
		if value.True(ok) {
			return rest[i+1]
		}
	}
	if fallback != nil {
		return fallback
	}
	return value.ErrNA
}

var condBuiltins = []Builtin{
	ifsBuiltin,
	ifBuiltin,
	ifErrorBuiltin,
	ifNaBuiltin,
	andBuiltin,
	orBuiltin,
	xorBuiltin,
	notBuiltin,
	switchBuiltin,
	chooseBuiltin,
}
