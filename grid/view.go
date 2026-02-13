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

type CellType int8

const (
	TypeSharedString CellType = 1 << iota
	TypeString
	TypeNumber
	TypeDate
	TypeBool
	TypeFormula
)

type Cell interface {
	At() layout.Position
	Value() value.ScalarValue
	Reload(value.Context) error
	// Type() CellType
}

type Encoder interface {
	EncodeSheet(View) error
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
	Rows() iter.Seq[[]value.ScalarValue]
	Encode(Encoder) error
	Cell(layout.Position) (Cell, error)

	Reload(value.Context) error
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

	Reload() error

	// Merge(File) error
	Rename(string, string) error
	Copy(string, string) error
	Remove(string) error
}

type filteredView struct {
	view View
}

func FilterView(view View) View {
	return &filteredView{
		view: view,
	}
}

func (v *filteredView) Name() string {
	return v.view.Name()
}

func (v *filteredView) Bounds() *layout.Range {
	return nil
}

func (v *filteredView) Rows() iter.Seq[[]value.ScalarValue] {
	return nil
}

func (v *filteredView) Unwrap() View {
	return v.view
}

func (v *filteredView) Encode(encoder Encoder) error {
	return encoder.EncodeSheet(v.view)
}

func (v *filteredView) Cell(layout.Position) (Cell, error) {
	return nil, nil
}

func (v *filteredView) Reload(ctx value.Context) error {
	return v.view.Reload(ctx)
}

type readonlyView struct {
	view View
}

func ReadOnly(view View) View {
	if _, ok := view.(*readonlyView); ok {
		return view
	}
	return &readonlyView{
		view: view,
	}
}

func (v *readonlyView) Name() string {
	return v.view.Name()
}

func (v *readonlyView) Bounds() *layout.Range {
	return v.view.Bounds()
}

func (v *readonlyView) Rows() iter.Seq[[]value.ScalarValue] {
	return v.view.Rows()
}

func (v *readonlyView) Unwrap() View {
	return v.view
}

func (v *readonlyView) Encode(encoder Encoder) error {
	return encoder.EncodeSheet(v.view)
}

func (v *readonlyView) Cell(pos layout.Position) (Cell, error) {
	return v.view.Cell(pos)
}

func (v *readonlyView) Reload(ctx value.Context) error {
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

func (v *horizontalStackedView) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		var (
			list []func() ([]value.ScalarValue, bool)
			stop []func()
			dim  = v.Bounds()
		)
		for _, vs := range v.views {
			next, s := iter.Pull[[]value.ScalarValue](vs.Rows())
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
				rs, ok := n()
				if !ok {
					return
				}
				row = append(row, rs...)
			}
			if !yield(row) {
				return
			}
		}
	}
	return it
}

func (v *horizontalStackedView) Encode(e Encoder) error {
	return e.EncodeSheet(v)
}

func (v *horizontalStackedView) Cell(pos layout.Position) (Cell, error) {
	return nil, nil
}

func (v *horizontalStackedView) Reload(ctx value.Context) error {
	for i := range v.views {
		if err := v.views[i].Reload(ctx); err != nil {
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

func (v *verticalStackedView) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		for i := range v.views {
			for r := range v.views[i].Rows() {
				if !yield(r) {
					return
				}
			}
		}
	}
	return it
}

func (v *verticalStackedView) Encode(e Encoder) error {
	return e.EncodeSheet(v)
}

func (v *verticalStackedView) Cell(pos layout.Position) (Cell, error) {
	return nil, nil
}

func (v *verticalStackedView) Reload(ctx value.Context) error {
	for i := range v.views {
		if err := v.views[i].Reload(ctx); err != nil {
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
		v.mapping[c] = int64(i)
	}
	return &v
}

func (v *projectedView) Name() string {
	return v.view.Name()
}

func (v *projectedView) Reload(ctx value.Context) error {
	return v.view.Reload(ctx)
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
	pos = v.getOriginalPosition(pos)
	return v.view.Cell(pos)
}

func (v *projectedView) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		out := make([]value.ScalarValue, len(v.columns))
		for row := range v.view.Rows() {
			for i, col := range v.columns {
				if int(col) < len(row) {
					out[i] = row[col]
				}
			}
			if !yield(out) {
				return
			}
		}
	}
	return it
}

func (v *projectedView) SetValue(pos layout.Position, val value.ScalarValue) error {
	bd := v.Bounds()
	if !bd.Contains(pos) {
		return ErrPosition
	}
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	pos = v.getOriginalPosition(pos)
	return mv.SetValue(pos, val)
}

func (v *projectedView) SetFormula(pos layout.Position, val value.Formula) error {
	bd := v.Bounds()
	if !bd.Contains(pos) {
		return ErrPosition
	}
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	pos = v.getOriginalPosition(pos)
	return mv.SetFormula(pos, val)
}

func (v *projectedView) ClearCell(pos layout.Position) error {
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	return mv.ClearCell(pos)
}

func (v *projectedView) ClearValue(pos layout.Position) error {
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	pos = v.getOriginalPosition(pos)
	return mv.ClearValue(pos)
}

func (v *projectedView) ClearFormula(pos layout.Position) error {
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	pos = v.getOriginalPosition(pos)
	return mv.ClearFormula(pos)
}

func (v *projectedView) ClearRange(rg *layout.Range) error {
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	return mv.ClearRange(rg)
}

func (v *projectedView) Encode(encoder Encoder) error {
	return encoder.EncodeSheet(v)
}

func (v *projectedView) getOriginalPosition(pos layout.Position) layout.Position {
	if pos.Column < 0 || pos.Column > int64(len(v.columns)) {
		return pos
	}
	mod := layout.Position{
		Column: v.columns[pos.Column],
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

func (v *boundedView) Reload(ctx value.Context) error {
	return v.view.Reload(ctx)
}

func (v *boundedView) Cell(pos layout.Position) (Cell, error) {
	if !v.part.Contains(pos) {
		return nil, fmt.Errorf("position outside view range")
	}
	return v.view.Cell(pos)
}

func (v *boundedView) Bounds() *layout.Range {
	return v.part.Range()
}

func (v *boundedView) Unwrap() View {
	return v.view
}

func (v *boundedView) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		var (
			width = v.part.Ends.Column - v.part.Starts.Column + 1
			data  = make([]value.ScalarValue, width)
		)
		for row := v.part.Starts.Line; row <= v.part.Ends.Line; row++ {
			for col, ix := v.part.Starts.Column, 0; col <= v.part.Ends.Column; col++ {
				p := layout.Position{
					Line:   row,
					Column: col,
				}
				c, err := v.view.Cell(p)
				if err == nil {
					data[ix] = c.Value()
				}
				ix++
			}
			if !yield(data) {
				break
			}
		}
	}
	return it
}

func (v *boundedView) SetValue(pos layout.Position, val value.ScalarValue) error {
	if !v.part.Contains(pos) {
		return ErrPosition
	}
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}

	b := v.view.Bounds()

	pos.Line += b.Width()
	pos.Column += b.Height()
	return mv.SetValue(pos, val)
}

func (v *boundedView) SetFormula(pos layout.Position, val value.Formula) error {
	if !v.part.Contains(pos) {
		return ErrPosition
	}
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}

	b := v.view.Bounds()

	pos.Line += b.Width()
	pos.Column += b.Height()
	return mv.SetFormula(pos, val)
}

func (v *boundedView) ClearCell(pos layout.Position) error {
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	b := v.view.Bounds()
	pos.Line += b.Width()
	pos.Column += b.Height()
	return mv.ClearCell(pos)
}

func (v *boundedView) ClearValue(pos layout.Position) error {
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	b := v.view.Bounds()
	pos.Line += b.Width()
	pos.Column += b.Height()
	return mv.ClearValue(pos)
}

func (v *boundedView) ClearFormula(pos layout.Position) error {
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	b := v.view.Bounds()
	pos.Line += b.Width()
	pos.Column += b.Height()
	return mv.ClearFormula(pos)
}

func (v *boundedView) ClearRange(rg *layout.Range) error {
	mv, err := mutableView(v.view)
	if err != nil {
		return err
	}
	// b := v.view.Bounds()
	// pos.Line += b.Width()
	// pos.Column += b.Height()
	return mv.ClearRange(rg)
}

func (v *boundedView) Encode(e Encoder) error {
	return e.EncodeSheet(v)
}

func mutableView(v View) (MutableView, error) {
	mv, ok := v.(MutableView)
	if !ok {
		return nil, ErrWritable
	}
	return mv, nil
}
