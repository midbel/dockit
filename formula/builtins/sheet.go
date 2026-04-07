package builtins

import (
	gbs "github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/value"
)

var lockBuiltin = gbs.Builtin{
	Name:     "lock",
	Desc:     "",
	Category: "sheet",
	Params:   []gbs.Param{},
	Func:     Lock,
}

func Lock(args []value.Value) value.Value {
	return value.ErrValue
}

var unlockBuiltin = gbs.Builtin{
	Name:     "unlock",
	Desc:     "",
	Category: "sheet",
	Params:   []gbs.Param{},
	Func:     Unlock,
}

func Unlock(args []value.Value) value.Value {
	return value.ErrValue
}

var newFileBuiltin = gbs.Builtin{
	Name:     "newfile",
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
	newFileBuiltin,
	unlockBuiltin,
	lockBuiltin,
}
