package workbook

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/midbel/dockit/driver"
	"github.com/midbel/dockit/grid"
)

var registry []driver.Loader

func Register(loader driver.Loader) {
	registry = append(registry, loader)
}

func OpenFormat(file, name string) (grid.File, error) {
	if name == "" {
		return Open(file)
	}
	ix := slices.IndexFunc(registry, func(loader driver.Loader) bool {
		return loader.Name() == name
	})
	if ix < 0 {
		return nil, fmt.Errorf("%s: format not registered", name)
	}
	return registry[ix].Open(file)
}

func Open(file string) (grid.File, error) {
	ix := slices.IndexFunc(registry, func(loader driver.Loader) bool {
		ok, err := loader.Detect(file)
		if err != nil {
			ok = false
		}
		return ok
	})
	if ix < 0 {
		return nil, fmt.Errorf("unable to open given %s - unsupported format", file)
	}
	return registry[ix].Open(file)
}

func Formats() []string {
	var names []string
	for _, x := range registry {
		names = append(names, x.Name())
	}
	return names
}

func Merge(file string, sources []string) error {
	ix := slices.IndexFunc(registry, func(x driver.Loader) bool {
		return x.IsSupportedExt(filepath.Ext(file))
	})
	if ix < 0 {
		return fmt.Errorf("new file can not be created for %s", filepath.Ext(file))
	}
	wb, err := registry[ix].New()
	if err != nil {
		return err
	}
	mg, ok := wb.(driver.Merger)
	if !ok {
		return fmt.Errorf("%s does not support merging files", registry[ix].Name())
	}
	for _, s := range sources {
		wb, err := Open(s)
		if err != nil {
			return err
		}
		if err := mg.Merge(wb); err != nil {
			return err
		}
	}
	if w, ok := wb.(interface{ WriteFile(string) error }); ok {
		return w.WriteFile(file)
	}
	return nil
}
