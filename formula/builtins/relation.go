package builtins

import (
	"github.com/midbel/dockit/formula/types"
	gbs "github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/gridx"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var joinBuiltin = gbs.Builtin{
	Name:     "join",
	Desc:     "",
	Category: "relation",
	Params: []gbs.Param{
		gbs.Object("sheet1", "", value.TypeAny),
		gbs.Object("sheet2", "", value.TypeAny),
		gbs.Object("cols1", "", value.TypeAny),
		gbs.Object("cols2", "", value.TypeAny),
	},
	Func: Join,
}

func Join(args []value.Value) value.Value {
	keys1, err := layout.SelectionFromString(args[2].String())
	if err != nil {
		return value.ErrValue
	}
	keys2, err := layout.SelectionFromString(args[3].String())
	if err != nil {
		return value.ErrValue
	}
	v1, ok := args[0].(*types.View)
	if !ok {
		return value.ErrValue
	}
	v2, ok := args[1].(*types.View)
	if !ok {
		return value.ErrValue
	}
	v := gridx.Join(v1.View(), v2.View(), keys1, keys2)
	return types.NewViewValue(v)
}

var groupBuiltin = gbs.Builtin{
	Name:     "group",
	Desc:     "",
	Category: "relation",
	Params:   []gbs.Param{},
	Func:     Group,
}

func Group(args []value.Value) value.Value {
	return value.ErrValue
}

var relationBuiltins = []gbs.Builtin{
	joinBuiltin,
	groupBuiltin,
}
