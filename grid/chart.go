package grid

import (
	"github.com/midbel/dockit/layout"
)

type ChartType string

const (
	ChartBar     ChartType = "bar"
	ChartLine    ChartType = "line"
	ChartPie     ChartType = "pie"
	ChartScatter ChartType = "scatter"
	ChartArea    ChartType = "area"
)

type Chart struct {
	ID     string
	Title  string
	Type   ChartType
	Legend Legend
	Anchor Anchor

	Series []Series
	XAxis  *Axis
	YAxis  *Axis

	Options ChartOptions
}

type MarkerType string

const (
	MarkerNone     MarkerType = "none"
	MarkerCircle   MarkerType = "circle"
	MarkerSquare   MarkerType = "square"
	MarkerTriangle MarkerType = "triangle"
)

type SeriesStyle struct {
	Color string // "#RRGGBB"
	Width float64
	MarkerType
}

type Series struct {
	Name       string
	Categories layout.Range
	Values     layout.Range
	Style      *SeriesStyle
}

type Axis struct {
	Title string

	Min float64
	Max float64

	MajorUnit float64
	MinorUnit float64

	LogScale bool
	Reverse  bool
}

type LegendPosition string

const (
	LegendRight  LegendPosition = "right"
	LegendLeft   LegendPosition = "left"
	LegendTop    LegendPosition = "top"
	LegendBottom LegendPosition = "bottom"
)

type Legend struct {
	Visible  bool
	Position LegendPosition
}

type Anchor struct {
	From layout.Position
	To   layout.Position
}

type DataLabel struct {
	ShowValue bool
	ShowName  bool
	ShowPct   bool
}

type ChartOptions struct {
	Stacked     bool
	SmoothLines bool
	DataLabels  DataLabel
}
