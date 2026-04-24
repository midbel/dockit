package builtins

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	gbs "github.com/midbel/dockit/grid/builtins"
)

var registry = map[string]gbs.Builtin{}

func Lookup(ident string) (gbs.BuiltinFunc, error) {
	fn, err := gbs.Lookup(ident)
	if err == nil {
		return fn, nil
	}
	b, err := Get(ident)
	if err != nil {
		return nil, err
	}
	return b.Make(), nil
}

func Get(ident string) (gbs.Builtin, error) {
	fn, ok := registry[strings.ToLower(ident)]
	if ok {
		return fn, nil
	}
	vs := List()
	ix := slices.IndexFunc(vs, func(b gbs.Builtin) bool {
		return slices.Contains(b.Alias, ident)
	})
	if ix < 0 {
		return gbs.Builtin{}, fmt.Errorf("%s undefined builtin", ident)
	}
	return vs[ix], nil
}

func List() []gbs.Builtin {
	vs := maps.Values(registry)
	return slices.Collect(vs)
}

func init() {
	registerBuiltins(sheetBuiltins)
	registerBuiltins(relationBuiltins)
	registerBuiltins(numberBuiltins)
}

func registerBuiltins(list []gbs.Builtin) {
	for _, b := range list {
		registry[b.Name] = b
	}
}
