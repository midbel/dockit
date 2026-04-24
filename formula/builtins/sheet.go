package builtins

import (
	gbs "github.com/midbel/dockit/grid/builtins"
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
