package workbook

import (
	"fmt"
	"io"
	"os"
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
		return nil, fmt.Errorf("unable to open %s - unsupported format", file)
	}
	return registry[ix].Open(file)
}

func WriteView(view grid.View, file string) error {
	wb, err := createEmpty(filepath.Ext(file))
	if err != nil {
		return err
	}
	if err := wb.AppendSheet(view); err != nil {
		return err
	}
	return WriteFile(wb, file)
}

func WriteFile(wb grid.File, file string) error {
	if len(wb.Sheets()) == 0 {
		return fmt.Errorf("no sheet in workbook")
	}
	wt, ok := wb.(interface{ WriteTo(io.Writer) error })
	if ok {
		dir := filepath.Dir(file)
		tmp, err := os.CreateTemp(dir, ".dockit_tmp*")
		if err != nil {
			return err
		}
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()
		if err := wt.WriteTo(tmp); err != nil {
			return err
		}
		if err := tmp.Close(); err != nil {
			return err
		}
		return os.Rename(tmp.Name(), file)
	}
	wf, ok := wb.(interface{ WriteFile(string) error })
	if ok {
		return wf.WriteFile(file)
	}
	return fmt.Errorf("fail to write workbook to file %s", file)
}

func Formats() []string {
	var names []string
	for _, x := range registry {
		names = append(names, x.Name())
	}
	return names
}

func Merge(format string, sources []string) (grid.File, error) {
	wb, err := createEmpty(format)
	if err != nil {
		return nil, err
	}
	mg, ok := wb.(driver.Merger)
	if !ok {
		return nil, fmt.Errorf("%s does not support merging files", format)
	}
	for _, s := range sources {
		wb, err := Open(s)
		if err != nil {
			return nil, err
		}
		if err := mg.Merge(wb); err != nil {
			return nil, err
		}
	}
	return wb, nil
}

func createEmpty(format string) (grid.File, error) {
	ix := slices.IndexFunc(registry, func(x driver.Loader) bool {
		return x.IsSupportedExt(format)
	})
	if ix < 0 {
		return nil, fmt.Errorf("new file can not be created for %s", format)
	}
	return registry[ix].New()
}
