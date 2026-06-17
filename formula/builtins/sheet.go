package builtins

import (
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/formula/types"
	gbs "github.com/midbel/dockit/grid/builtins"
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
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f := flat.NewFile()
	return types.NewFileValue(f, false)
}

var newSheetBuiltin = gbs.Builtin{
	Name:     "sheet",
	Desc:     "",
	Category: "sheet",
	Params: []gbs.Param{
		gbs.Scalar("name", "", value.TypeText),
	},
	Func: EmptySheet,
}

func EmptySheet(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		name = asString(args[0])
		sh   = flat.NewSheet(name, nil)
	)
	return types.NewViewValue(sh)
}

var mkRangeBuiltin = gbs.Builtin{
	Name:     "mkrange",
	Desc:     "",
	Category: "sheet",
	Params: []gbs.Param{
		gbs.Scalar("fromCol", "", value.TypeNumber),
		gbs.Scalar("fromRow", "", value.TypeNumber),
		gbs.Scalar("toCol", "", value.TypeNumber),
		gbs.Scalar("toRow", "", value.TypeNumber),
	},
	Func: MakeRange,
}

func MakeRange(args []value.Value) value.Value {
	var (
		fromCol = asFloat(args[0])
		fromRow = asFloat(args[1])
		toCol   = asFloat(args[2])
		toRow   = asFloat(args[2])
		start   = layout.NewPosition(int64(fromRow), int64(fromCol))
		end     = layout.NewPosition(int64(toRow), int64(toCol))
	)
	return types.NewRangeValue(start, end)
}

var mkRefBuiltin = gbs.Builtin{
	Name:     "mkref",
	Desc:     "",
	Category: "sheet",
	Params:   []gbs.Param{},
	Func:     MakeAddr,
}

func MakeAddr(args []value.Value) value.Value {
	return value.ErrValue
}

var mergeBuiltin = gbs.Builtin{
	Name:     "merge",
	Desc:     "",
	Category: "sheet",
	Params: []gbs.Param{
		gbs.Var(gbs.Object("value", "", value.TypeAny)),
	},
	Func:     Merge,
}

func Merge(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f := types.NewFileValue(flat.NewFile(), false).(*types.File)
	for _, a := range args {
		v, ok := a.(*types.View)
		if !ok {
			return value.ErrValue
		}
		f.Append(v)
	}
	return nil
}

var sheetBuiltins = []gbs.Builtin{
	mkRefBuiltin,
	mkRangeBuiltin,
	newSheetBuiltin,
	fileBuiltin,
	unlockBuiltin,
	lockBuiltin,
	joinBuiltin,
	copyBuiltin,
	mergeBuiltin,
}
