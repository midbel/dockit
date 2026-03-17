package grid

import (
	"github.com/midbel/dockit/layout"
)

type Dep struct {
	UsedBy    []layout.Position
	DependsOn []layout.Position

	dirty bool
}

type Graph struct {
	deps map[layout.Position]*Dep
}

func NewGraph() *Graph {
	g := &Graph{
		deps: make(map[layout.Position]*Dep),
	}
	return g
}
