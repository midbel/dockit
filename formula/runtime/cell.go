package runtime

import (
	"github.com/midbel/dockit/grid"
)

type Cell struct {
	grid.Cell
	view *View
}
