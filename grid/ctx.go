package grid

import (
	"github.com/midbel/dockit/value"
)

type viewValue struct {
	view View
}

func (*viewValue) Kind() value.ValueKind {
	return value.KindObject
}

func (c *viewValue) String() string {
	return c.view.Name()
}

func (c *viewValue) Get(ident string) (value.ScalarValue, error) {
	return nil, nil
}


type fileValue struct {
	file File
}

func (*fileValue) Kind() value.ValueKind {
	return value.KindObject
}

func (c *fileValue) String() string {
	return "workbook"
}

func (c *fileValue) Get(ident string) (value.ScalarValue, error) {
	return nil, nil
}
