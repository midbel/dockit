package grid

import (
	"errors"
	"fmt"
	"iter"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var (
	ErrFile      = errors.New("invalid spreadsheet")
	ErrLock      = errors.New("spreadsheet locked")
	ErrSupported = errors.New("operation not supported")
	ErrFound     = errors.New("not found")
	ErrPosition  = errors.New("invalid position")
	ErrWritable  = errors.New("read only view")
	ErrEmpty     = errors.New("empty context")
	ErrMutate    = errors.New("context is not mutable")
	ErrType      = errors.New("invalid type")
)

func NoCell(pos layout.Position) error {
	return fmt.Errorf("%s no cell at given position", pos)
}

type CopyMode int

func CopyModeFromString(str string) (CopyMode, error) {
	var mode CopyMode
	switch str {
	case "value":
		mode |= CopyValue
	case "formula":
		mode |= CopyFormula
	case "style":
		mode |= CopyStyle
	case "", "all":
		mode |= CopyAll
	default:
		return mode, fmt.Errorf("%s invalid value for copy mode", str)
	}
	return mode, nil
}

const (
	CopyValue = iota << 1
	CopyFormula
	CopyStyle
	CopyAll = CopyValue | CopyFormula | CopyStyle
)

type Row interface {
	Values() []value.ScalarValue
	Sparse() bool
}

type Cell interface {
	At() layout.Position
	Value() value.ScalarValue
	Formula() value.Formula
}

func ResetAt(cell Cell, pos layout.Position) Cell {
	if a, ok := cell.(interface{ SetAt(layout.Position) }); ok {
		a.SetAt(pos)
	}
	return cell
}

type empty struct {
	pos layout.Position
}

func Empty(pos layout.Position) Cell {
	return empty{
		pos: pos,
	}
}

func (c empty) At() layout.Position {
	return c.pos
}

func (c empty) Value() value.ScalarValue {
	return value.Empty()
}

func (c empty) Formula() value.Formula {
	return nil
}

type Callable interface {
	Call(value.Context) (value.Value, error)
}

type UnwrapView interface {
	Unwrap() View
}

func Unwrap(view View) View {
	for {
		u, ok := view.(UnwrapView)
		if !ok {
			break
		}
		view = u.Unwrap()
	}
	return view
}

type View interface {
	Name() string
	Bounds() *layout.Range
	Rows() iter.Seq2[int64, []value.ScalarValue]
	Cell(layout.Position) (Cell, error)

	Sync(value.Context) error
}

type MutableView interface {
	View

	SetValue(layout.Position, value.ScalarValue) error
	SetFormula(layout.Position, value.Formula) error

	ClearCell(layout.Position) error
	ClearValue(layout.Position) error
	ClearFormula(layout.Position) error
	ClearRange(*layout.Range) error
}

type ViewInfo struct {
	Name      string
	Active    bool
	Hidden    bool
	Size      layout.Dimension
	Protected bool
}

type File interface {
	Infos() []ViewInfo
	ActiveSheet() (View, error)
	Sheet(string) (View, error)
	Sheets() []View

	Sync() error

	Rename(string, string) error
	Copy(string, string) error
	AppendSheet(View) error
	RemoveSheet(string) error
}

type transposedView struct {
	view View
}

func NewTransposedView(view View) View {
	if v, ok := view.(*transposedView); ok {
		return v
	}
	v := &transposedView{
		view: view,
	}
	return v
}

func (v *transposedView) Name() string {
	return v.view.Name()
}

func (v *transposedView) Type() string {
	return "transpose"
}

func (v *transposedView) Bounds() *layout.Range {
	bs := v.view.Bounds()
	return bs.Transpose()
}

func (v *transposedView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		bs := v.Bounds()
		for row := int64(1); row <= bs.Height(); row++ {
			rs := make([]value.ScalarValue, bs.Width())
			for col := int64(1); col <= bs.Width(); col++ {
				p := layout.Position{
					Line:   col,
					Column: row,
				}
				c, err := v.view.Cell(p)
				if err != nil {
					rs[col-1] = value.ErrNA
				} else {
					rs[col-1] = c.Value()
				}
			}
			if !yield(row, rs) {
				return
			}
		}
	}
	return it
}

func (v *transposedView) Unwrap() View {
	return v.view
}

func (v *transposedView) Cell(pos layout.Position) (Cell, error) {
	p := layout.Position{
		Line:   pos.Column,
		Column: pos.Line,
	}
	cell, err := v.view.Cell(p)
	if err != nil {
		cell = Empty(pos)
	}
	return ResetAt(cell, pos), nil
}

func (v *transposedView) Sync(ctx value.Context) error {
	return ErrWritable
}

type horizontalStackedView struct {
	views []View
}

func HorizontalView(views ...View) View {
	if len(views) == 1 {
		return views[0]
	}
	v := horizontalStackedView{
		views: views,
	}
	return &v
}

func (v *horizontalStackedView) Name() string {
	return v.views[0].Name()
}

func (v *horizontalStackedView) Type() string {
	return "horizontal-stack"
}

func (v *horizontalStackedView) Bounds() *layout.Range {
	start := layout.Position{
		Line:   1,
		Column: 1,
	}
	var (
		width  int64
		height int64
	)
	for i := range v.views {
		rg := v.views[i].Bounds()
		if i == 0 {
			width = rg.Width()
		}
		height += rg.Height()
	}
	end := layout.Position{
		Line:   height,
		Column: width,
	}
	return layout.NewRange(start, end)
}

func (v *horizontalStackedView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		var (
			list []func() (int64, []value.ScalarValue, bool)
			stop []func()
			dim  = v.Bounds()
		)
		for _, vs := range v.views {
			next, s := iter.Pull2[int64, []value.ScalarValue](vs.Rows())
			stop = append(stop, s)
			list = append(list, next)
		}
		defer func() {
			for _, s := range stop {
				s()
			}
		}()
		for i := int64(0); i < dim.Height(); i++ {
			row := make([]value.ScalarValue, 0, dim.Width())
			for _, n := range list {
				_, rs, ok := n()
				if !ok {
					return
				}
				row = append(row, rs...)
			}
			if !yield(i, row) {
				return
			}
		}
	}
	return it
}

func (v *horizontalStackedView) Cell(pos layout.Position) (Cell, error) {
	var (
		lino int64
		ori  = pos
	)
	lino++
	for _, sh := range v.views {
		b := sh.Bounds()
		if pos.Line >= lino && pos.Line < lino+b.Height() {
			pos.Line = pos.Line - (lino - 1)
			cell, _ := sh.Cell(pos)
			return ResetAt(cell, ori), nil
		}
		lino += b.Height()
	}
	return Empty(pos), nil
}

func (v *horizontalStackedView) Sync(ctx value.Context) error {
	for i := range v.views {
		if err := v.views[i].Sync(ctx); err != nil {
			return err
		}
	}
	return nil
}

type verticalStackedView struct {
	views []View
}

func VerticalView(views ...View) View {
	if len(views) == 1 {
		return views[0]
	}
	v := verticalStackedView{
		views: views,
	}
	return &v
}

func (v *verticalStackedView) Name() string {
	return v.views[0].Name()
}

func (v *verticalStackedView) Type() string {
	return "vertical-stack"
}

func (v *verticalStackedView) Bounds() *layout.Range {
	start := layout.Position{
		Line:   1,
		Column: 1,
	}
	var (
		width  int64
		height int64
	)
	for i := range v.views {
		rg := v.views[i].Bounds()
		if i == 0 {
			height = rg.Height()
		}
		width += rg.Width()
	}
	end := layout.Position{
		Line:   height,
		Column: width,
	}
	return layout.NewRange(start, end)
}

func (v *verticalStackedView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		var lino int64
		for i := range v.views {
			for _, r := range v.views[i].Rows() {
				lino++
				if !yield(lino, r) {
					return
				}
			}
		}
	}
	return it
}

func (v *verticalStackedView) Cell(pos layout.Position) (Cell, error) {
	var (
		col int64
		ori = pos
	)
	col++
	for _, sh := range v.views {
		b := sh.Bounds()
		if pos.Column >= col && pos.Column < col+b.Width() {
			pos.Column = pos.Column - (col - 1)
			cell, _ := sh.Cell(pos)
			return ResetAt(cell, ori), nil
		}
		col += b.Width()
	}
	return Empty(pos), nil
}

func (v *verticalStackedView) Sync(ctx value.Context) error {
	for i := range v.views {
		if err := v.views[i].Sync(ctx); err != nil {
			return err
		}
	}
	return nil
}

type projectedView struct {
	view    View
	columns []int64
	mapping map[int64]int64
}

func NewProjectView(view View, sel layout.Selection) View {
	v := projectedView{
		view:    view,
		columns: sel.Indices(view.Bounds()),
		mapping: make(map[int64]int64),
	}
	for i, c := range v.columns {
		v.mapping[int64(i)+1] = c + 1
	}
	return &v
}

func (v *projectedView) Name() string {
	return v.view.Name()
}

func (v *projectedView) Type() string {
	return "projected"
}

func (v *projectedView) Sync(ctx value.Context) error {
	return v.view.Sync(ctx)
}

func (v *projectedView) Bounds() *layout.Range {
	rg := v.view.Bounds()

	start := layout.Position{
		Line:   1,
		Column: 1,
	}
	if rg.Width() == 0 && rg.Height() == 0 {
		return layout.NewRange(start, start)
	}
	end := layout.Position{
		Line:   rg.Height(),
		Column: max(1, int64(len(v.columns))),
	}
	return layout.NewRange(start, end)
}

func (v *projectedView) Unwrap() View {
	return v.view
}

func (v *projectedView) Cell(pos layout.Position) (Cell, error) {
	mapped := v.getOriginalPosition(pos)
	cell, err := v.view.Cell(mapped)
	if err != nil {
		cell = Empty(pos)
	}
	return ResetAt(cell, pos), nil
}

func (v *projectedView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		var (
			out  = make([]value.ScalarValue, len(v.columns))
			lino int64
		)
		for _, row := range v.view.Rows() {
			lino++
			for i, col := range v.columns {
				if int(col) < len(row) {
					out[i] = row[col]
				}
			}
			if !yield(lino, out) {
				return
			}
		}
	}
	return it
}

func (v *projectedView) getOriginalPosition(pos layout.Position) layout.Position {
	if pos.Column < 0 || pos.Column > int64(len(v.columns)) {
		return pos
	}
	mod := layout.Position{
		Column: v.mapping[pos.Column],
		Line:   pos.Line,
	}
	return mod
}

type boundedView struct {
	view View
	part *layout.Range
}

func NewBoundedView(view View, rg *layout.Range) View {
	v := boundedView{
		view: view,
		part: rg.Normalize(),
	}
	return &v
}

func (v *boundedView) Name() string {
	return v.view.Name()
}

func (v *boundedView) Type() string {
	return "bounded"
}

func (v *boundedView) Sync(ctx value.Context) error {
	return v.view.Sync(ctx)
}

func (v *boundedView) Cell(pos layout.Position) (Cell, error) {
	rg := v.Bounds()
	if !rg.Contains(pos) {
		return Empty(pos), nil
	}
	cell, err := v.view.Cell(v.recomputePosition(pos))
	if err != nil {
		cell = Empty(pos)
	}
	return ResetAt(cell, pos), nil
}

func (v *boundedView) Bounds() *layout.Range {
	return v.part.Range()
}

func (v *boundedView) Unwrap() View {
	return v.view
}

func (v *boundedView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		var (
			width = v.part.Ends.Column - v.part.Starts.Column + 1
			out   = make([]value.ScalarValue, width)
			lino  int64
		)
		for row := v.part.Starts.Line; row <= v.part.Ends.Line; row++ {
			lino++
			for col, ix := v.part.Starts.Column, 0; col <= v.part.Ends.Column; col++ {
				p := layout.Position{
					Line:   row,
					Column: col,
				}
				c, err := v.view.Cell(p)
				if err == nil {
					out[ix] = c.Value()
				}
				ix++
			}
			if !yield(lino, out) {
				break
			}
		}
	}
	return it
}

func (v *boundedView) recomputePosition(pos layout.Position) layout.Position {
	pos.Line = v.part.Starts.Line + pos.Line - 1
	pos.Column = v.part.Starts.Column + pos.Column - 1
	return pos
}
