package builtins

import (
	"github.com/midbel/dockit/formula/types"
	gbs "github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/gridx"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var copyBuiltin = gbs.Builtin{
	Name:     "copy",
	Desc:     "",
	Category: "sheet",
	Func:     CopySheet,
}

func CopySheet(args []value.Value) value.Value {
	return value.ErrValue
}

var joinBuiltin = gbs.Builtin{
	Name:     "join",
	Desc:     "",
	Category: "sheet",
	Params: []gbs.Param{
		gbs.Object("sheet1", "", value.TypeAny),
		gbs.Object("sheet2", "", value.TypeAny),
		gbs.Object("cols1", "", value.TypeAny),
		gbs.Object("cols2", "", value.TypeAny),
	},
	Func: Lock,
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

var lockBuiltin = gbs.Builtin{
	Name:     "lock",
	Desc:     "",
	Category: "sheet",
	Params: []gbs.Param{
		gbs.Object("value", "", value.TypeAny),
	},
	Func: Lock,
}

func Lock(args []value.Value) value.Value {
	k, ok := args[0].(interface{ Lock() error })
	if ok {
		err := k.Lock()
		if err != nil {
			return value.ErrValue
		}
	}
	return value.Boolean(true)
}

var unlockBuiltin = gbs.Builtin{
	Name:     "unlock",
	Desc:     "",
	Category: "sheet",
	Params: []gbs.Param{
		gbs.Object("value", "", value.TypeAny),
	},
	Func: Unlock,
}

func Unlock(args []value.Value) value.Value {
	k, ok := args[0].(interface{ Unlock() error })
	if ok {
		err := k.Unlock()
		if err != nil {
			return value.ErrValue
		}
	}
	return value.Boolean(true)
}

var fileBuiltin = gbs.Builtin{
	Name:     "file",
	Desc:     "",
	Category: "sheet",
	Params:   []gbs.Param{},
	Func:     EmptyFile,
}

func EmptyFile(args []value.Value) value.Value {
	return value.ErrValue
}

var newSheetBuiltin = gbs.Builtin{
	Name:     "newsheet",
	Desc:     "",
	Category: "sheet",
	Params:   []gbs.Param{},
	Func:     EmptySheet,
}

func EmptySheet(args []value.Value) value.Value {
	return value.ErrValue
}

var mkRangeBuiltin = gbs.Builtin{
	Name:     "mkrange",
	Desc:     "",
	Category: "sheet",
	Params:   []gbs.Param{},
	Func:     MakeRange,
}

func MakeRange(args []value.Value) value.Value {
	return value.ErrValue
}

var mkAddrBuiltin = gbs.Builtin{
	Name:     "mkaddr",
	Desc:     "",
	Category: "sheet",
	Params:   []gbs.Param{},
	Func:     MakeAddr,
}

func MakeAddr(args []value.Value) value.Value {
	return value.ErrValue
}

var sheetBuiltins = []gbs.Builtin{
	mkAddrBuiltin,
	mkRangeBuiltin,
	newSheetBuiltin,
	fileBuiltin,
	unlockBuiltin,
	lockBuiltin,
	joinBuiltin,
	copyBuiltin,
}
