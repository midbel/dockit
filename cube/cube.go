package cube

import (
	"fmt"
	"slices"

	"github.com/midbel/dockit/value"
)

type AggrFunc func([]value.ScalarValue) (value.ScalarValue, error)

type FilterFunc func(value.ScalarValue) bool

type Cube struct {
	dimensions []*Dimension
	measures   []*Measure
	measIndex  map[string]int
	rows       int
}

func New() *Cube {
	var c Cube
	c.measIndex = make(map[string]int)
	return &c
}

func (c *Cube) AggregateAll(measure string, aggr AggrFunc, keep FilterFunc) (value.ScalarValue, error) {
	ix, ok := c.measIndex[measure]
	if !ok {
		return nil, fmt.Errorf("%s: measurement not found", measure)
	}
	values := c.measures[i].Data
	if keep == nil {
		return agg(values)
	}
	var list []value.ScalarValue
	for _, v := range values {
		if !keep(v) {
			continue
		}
		list = append(list, v)
	}
	return aggr(list)
}

func (c *Cube) AddRow(dims []string, measures []value.ScalarValue) error {
	if len(c.dimensions) == 0 {
		return fmt.Errorf("cube has no dimension")
	}
	if len(dims) != len(c.dimensions) {
		return fmt.Errorf("invalid number of dimensions given")
	}
	if len(measures) != len(c.measures) {
		return fmt.Errorf("invalid number of measurements given")
	}
	for i := range c.dimensions {
		c.dimensions[i].RegisterValue(dims[i])
	}
	for i := range c.measures {
		c.measures[i].Append(measures[i])
	}
	c.rows++
	return nil
}

func (c *Cube) RegisterMeasure(name string) error {
	_, ok := c.measIndex[name]
	if ok {
		return fmt.Errorf("%s: measure already registered", name)
	}
	c.measIndex[name] = len(c.measures)
	c.measures = append(c.measures, NewMesure(name))
	return nil
}

func (c *Cube) RegisterDimension(name string, values []string) error {
	ix := slices.IndexFunc(c.dimensions, func(d Dimension) bool {
		return d.Name == name
	})
	if ix >= 0 {
		return fmt.Errorf("%s: dimension already registered", name)
	}
	d := NewDimension(name, values)
	c.dimensions = append(c.dimensions, d)
	return nil
}

type Dimension struct {
	Name string

	Dict  []string       // id → value
	Index map[string]int // value → id

	Column []int // dimension ids for each row
}

func NewDimension(name string, values []string) *Dimension {
	d := Dimension{
		Name:  name,
		Index: make(map[string]int),
	}
	for _, v := range values {
		d.RegisterValue(v)
	}
	return &d
}

func (d *Dimension) RegisterValue(value string) {
	_, ok := d.Index[v]
	if ok {
		return
	}
	d.Index[v] = len(d.Dict)
	d.Dict = append(d.Dict, v)
}

type Measure struct {
	Name string
	Data []value.ScalarValue
}

func NewMeasure(name string) *Measure {
	return &Measure{
		Name: name,
	}
}

func (m *Measure) Append(value value.ScalarValue) {
	m.Data = append(m.Data, value)
}
