package builtins

import (
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
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
	Params: []gbs.Param{
		gbs.Object("sheet", "", value.TypeAny),
	},
	Func: Group,
}

func Group(args []value.Value) value.Value {
	return value.ErrValue
}

var unionBuiltin = gbs.Builtin{
	Name:     "union",
	Desc:     "",
	Category: "relation",
	Params: []gbs.Param{
		gbs.Object("sheet1", "", value.TypeAny),
		gbs.Object("sheet2", "", value.TypeAny),
	},
	Func: Union,
}

func Union(args []value.Value) value.Value {
	return combineViews(args[0], args[1], gridx.Union)
}

var intersectBuiltin = gbs.Builtin{
	Name:     "intersect",
	Desc:     "",
	Category: "relation",
	Params: []gbs.Param{
		gbs.Object("sheet1", "", value.TypeAny),
		gbs.Object("sheet2", "", value.TypeAny),
	},
	Func: Group,
}

func Intersect(args []value.Value) value.Value {
	return combineViews(args[0], args[1], gridx.Intersect)
}

var exceptBuiltin = gbs.Builtin{
	Name:     "except",
	Desc:     "",
	Category: "relation",
	Params: []gbs.Param{
		gbs.Object("sheet1", "", value.TypeAny),
		gbs.Object("sheet2", "", value.TypeAny),
	},
	Func: Group,
}

func Except(args []value.Value) value.Value {
	return combineViews(args[0], args[1], gridx.Except)
}

type combineFunc func(grid.View, grid.View) (grid.View, error)

func combineViews(v1 value.Value, v2 value.Value, fn combineFunc) value.Value {
	left, ok := v1.(*types.View)
	if !ok {
		return value.ErrValue
	}
	right, ok := v2.(*types.View)
	if !ok {
		return value.ErrValue
	}
	v, err := fn(left.View(), right.View())
	if err != nil {
		return value.ErrValue
	}
	return types.NewViewValue(v)
}

var relationBuiltins = []gbs.Builtin{
	joinBuiltin,
	groupBuiltin,
	unionBuiltin,
	intersectBuiltin,
	exceptBuiltin,
}
