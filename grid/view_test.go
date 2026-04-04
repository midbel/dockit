package grid_test

import (
	"iter"
	"strings"
	"testing"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/testutil"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

const sample1 = `project,star,commit,language
foo,10,2023,Go
bar,13,452,C
flim,156,892,Rust
glam,42,1105,TypeScript
zorp,804,342,Go
munt,424,11127,C`

const sample2 = `project,star,commit,language
rift,2092,8125,C
skarn,631,10320,Rust
yoke,2447,1080,C
zenith,2172,4159,Go
quark,2204,12342,Java
hewn,461,4848,C++
`
const sample3 = `license,repo
MIT,https://github.com/midbel/foo
Apache-2.0,https://github.com/midbel/bar
MIT,https://github.com/midbel/flim
BSD-3,https://github.com/midbel/glam
MIT,https://github.com/midbel/zorp
GPL-3.0,https://github.com/octocat/munt`

func TestViews(t *testing.T) {
	t.Run("bounded-view", testBoundedView)
	t.Run("project-view", testProjectView)
	t.Run("transpose-view", testTransposeView)
	t.Run("horizontal-stack-view", testHorizontalStackView)
	t.Run("vertical-stack-view", testVerticalStackView)
	t.Run("combined-view", testCombinedViews)
}

func testCombinedViews(t *testing.T) {

}

func testBoundedView(t *testing.T) {
	var (
		sheet = getSheetFromSample(t, sample1)
		sbd   = sheet.Bounds()
		rg    = layout.NewRange(
			layout.NewPosition(2, 1),
			layout.NewPosition(4, 2),
		)
		view = grid.NewBoundedView(sheet, rg)
		vbd  = view.Bounds()
	)
	if vbd.Width() != rg.Width() || vbd.Height() != rg.Height() {
		t.Fatalf("view bounds does not match building range")
	}
	if vbd.Width() == sbd.Width() && vbd.Height() == sbd.Height() {
		t.Fatalf("view bounds should not match sheet bounds")
	}

	for pos := range vbd.Positions() {
		other := pos.Offset(1, 0)

		var (
			cell1, _ = view.Cell(pos)
			cell2, _ = sheet.Cell(other)
			ok       = value.Eq(cell1.Value(), cell2.Value())
		)
		if !value.True(ok) {
			t.Errorf("value mismatched at %s vs %s! want %s, got %s", pos, other, cell1.Value(), cell2.Value())
		}
	}
}

func testProjectView(t *testing.T) {
	var (
		sheet   = getSheetFromSample(t, sample1)
		sbd     = sheet.Bounds()
		cols, _ = layout.SelectionFromString("A;D")
		view    = grid.NewProjectView(sheet, cols)
		vbd     = view.Bounds()
	)
	if vbd.Width() == sbd.Width() {
		t.Fatalf("view width should not match sheet width")
	}
	if vbd.Height() != sbd.Height() {
		t.Fatalf("view height should match sheet height")
	}
	var (
		other   layout.Position
		columns = []int64{1, 4}
	)
	for pos := range vbd.Positions() {
		other.Line = pos.Line
		other.Column = columns[pos.Column-1]
		var (
			cell1, _ = view.Cell(pos)
			cell2, _ = sheet.Cell(other)
			ok       = value.Eq(cell1.Value(), cell2.Value())
		)
		if !value.True(ok) {
			t.Errorf("value mismatched at %s vs %s! want %s, got %s", pos, other, cell1.Value(), cell2.Value())
		}
	}
}

func testTransposeView(t *testing.T) {
	var (
		sheet = getSheetFromSample(t, sample1)
		sbd   = sheet.Bounds()
		view  = grid.NewTransposedView(sheet)
		vbd   = view.Bounds()
	)
	if vbd.Width() != sbd.Height() {
		t.Fatalf("view width should be equal to sheet height")
	}
	if vbd.Height() != sbd.Width() {
		t.Fatalf("view height should be equal to sheet width")
	}
	var other layout.Position
	for pos := range vbd.Positions() {
		other.Line = pos.Column
		other.Column = pos.Line

		var (
			cell1, _ = view.Cell(pos)
			cell2, _ = sheet.Cell(other)
			ok       = value.Eq(cell1.Value(), cell2.Value())
		)
		if !value.True(ok) {
			t.Errorf("value mismatched at %s vs %s! want %s, got %s", pos, other, cell1.Value(), cell2.Value())
		}
	}
}

func testHorizontalStackView(t *testing.T) {
	var (
		sh1  = getSheetFromSample(t, sample1)
		sbd1 = sh1.Bounds()
		sh2  = getSheetFromSample(t, sample2)
		sbd2 = sh2.Bounds()
		view = grid.HorizontalView(sh1, sh2)
		vbd  = view.Bounds()
	)
	if vbd.Width() != sbd1.Width() || vbd.Width() != sbd2.Width() {
		t.Fatalf("view width should be equal to sheets width")
	}
	if vbd.Height() != sbd1.Height()+sbd2.Height() {
		t.Fatalf("view height should be equal to sheets height")
	}

	it, stop := iter.Pull(vbd.Positions())
	defer stop()
	for pos := range sbd1.Positions() {
		other, has := it()
		if !has {
			t.Fatalf("all positions consumed")
		}
		var (
			cell1, _ = view.Cell(other)
			cell2, _ = sh1.Cell(pos)
			ok       = value.Eq(cell1.Value(), cell2.Value())
		)
		if !value.True(ok) {
			t.Errorf("value mismatched at %s vs %s! want %s, got %s", other, pos, cell1.Value(), cell2.Value())
		}
	}
	for pos := range sbd2.Positions() {
		other, has := it()
		if !has {
			t.Fatalf("all positions consumed")
		}
		var (
			cell1, _ = view.Cell(other)
			cell2, _ = sh2.Cell(pos)
			ok       = value.Eq(cell1.Value(), cell2.Value())
		)
		if !value.True(ok) {
			t.Errorf("value mismatched at %s vs %s! want %s, got %s", other, pos, cell1.Value(), cell2.Value())
		}
	}
	_, ok := it()
	if ok {
		t.Fatalf("no positions should remains")
	}
}

func testVerticalStackView(t *testing.T) {
	var (
		sh1  = getSheetFromSample(t, sample1)
		sbd1 = sh1.Bounds()
		sh2  = getSheetFromSample(t, sample3)
		sbd2 = sh2.Bounds()
		view = grid.VerticalView(sh1, sh2)
		vbd  = view.Bounds()
	)
	if vbd.Height() != sbd1.Height() || vbd.Height() != sbd2.Height() {
		t.Fatalf("view height should be equal to sheets height")
	}
	if vbd.Width() != sbd1.Width()+sbd2.Width() {
		t.Fatalf("view width should be equal to sheets width")
	}
}

func getSheetFromSample(t *testing.T, content string) grid.View {
	t.Helper()

	file, err := testutil.CreateCsvFile(strings.NewReader(content))
	if err != nil {
		t.Fatalf("fail to create csv file: %s", err)
	}

	sheet, err := file.ActiveSheet()
	if err != nil {
		t.Fatalf("fail to retrieve active sheet: %s", err)
	}
	return sheet
}

func TestSync(t *testing.T) {
	t.Run("force-sync", testForceSync)
	t.Run("empty-no-sync", testEmptyWithoutSync)
}

func testEmptyWithoutSync(t *testing.T) {
	var (
		file  = testutil.CreateFile()
		blank = value.Empty()
		pos   = layout.NewPosition(1, 3)
		want  = value.Float(24)
	)

	sh, err := file.Sheet("sheet1")
	if err != nil {
		t.Fatalf("fail to get sheet1: %s", err)
	}
	cell, err := sh.Cell(pos)
	if err != nil {
		t.Fatalf("fail to get cell at %s: %s", pos, err)
	}
	if got := cell.Value(); value.True(value.Ne(blank, got)) {
		t.Errorf("expected empty value in sheet1!C1! got %s", got)
	}
	if f := cell.Formula(); f == nil {
		t.Fatalf("expected formula to not be nil")
	} else {
		val, err := grid.Eval(f, grid.FileContext(file))
		if err != nil {
			t.Fatalf("error evaluating formula: %s", err)
		}
		if value.True(value.Ne(val, want)) {
			t.Errorf("evaluation failed! want %s, got %s", want, val)
		}
	}
	if got := cell.Value(); value.True(value.Ne(blank, got)) {
		t.Errorf("expected empty value in sheet1!C1! got %s", got)
	}
}

func testForceSync(t *testing.T) {
	var (
		file  = testutil.CreateFile()
		blank = value.Empty()
		pos   = layout.NewPosition(1, 3)
		want  = value.Float(24)
	)

	sh, err := file.Sheet("sheet1")
	if err != nil {
		t.Fatalf("fail to get sheet1: %s", err)
	}
	cell, err := sh.Cell(pos)
	if err != nil {
		t.Fatalf("fail to get cell at %s: %s", pos, err)
	}
	if got := cell.Value(); value.True(value.Ne(blank, got)) {
		t.Errorf("expected empty value in sheet1!C1! got %s", got)
	}
	if err := file.Sync(); err != nil {
		t.Errorf("fail to sync file: %s", err)
	}
	if got := cell.Value(); value.True(value.Ne(want, got)) {
		t.Errorf("expected value in sheet1!C1 to be %s! got %s", want, got)
	}
}
