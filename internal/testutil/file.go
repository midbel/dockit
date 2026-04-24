package testutil

import (
	"io"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

func FileContext() value.Context {
	file := CreateFile()
	return grid.FileContext(file)
}

func CreateCsvFile(r io.Reader) (grid.File, error) {
	return flat.OpenReader(csv.NewReader(r))
}

func CreateFile() grid.File {
	sheet1 := flat.NewSheet("sheet1", value.Rows(
		[]value.ScalarValue{value.Text("foo"), value.Float(2)},
		[]value.ScalarValue{value.Text("bar"), value.Float(5)},
	))
	f1, _ := grid.ParseOxmlFormula("=(B1 + sheet2!B1)*2")
	f2, _ := grid.ParseOxmlFormula("=(B2 + sheet2!B2)*2")
	sheet1.SetFormula(layout.NewPosition(1, 3), f1)
	sheet1.SetFormula(layout.NewPosition(2, 3), f2)

	sheet2 := flat.NewSheet("sheet2", value.Rows(
		[]value.ScalarValue{value.Text("quz"), value.Float(10)},
		[]value.ScalarValue{value.Text("bee"), value.Float(5)},
	))
	f3, _ := grid.ParseOxmlFormula("=UPPER(sheet2!A1)")
	f4, _ := grid.ParseOxmlFormula("=UPPER(sheet2!A2)")
	f5, _ := grid.ParseOxmlFormula("=UPPER(sheet1!A1)")
	f6, _ := grid.ParseOxmlFormula("=UPPER(sheet1!A2)")
	sheet2.SetFormula(layout.NewPosition(1, 3), f3)
	sheet2.SetFormula(layout.NewPosition(1, 4), f5)
	sheet2.SetFormula(layout.NewPosition(2, 3), f4)
	sheet2.SetFormula(layout.NewPosition(2, 4), f6)

	sheet3 := flat.NewSheet("sheet3", value.Rows(
		[]value.ScalarValue{value.Text("sum")},
		[]value.ScalarValue{value.Text("avg")},
		[]value.ScalarValue{value.Text("min")},
		[]value.ScalarValue{value.Text("max")},
	))
	f7, _ := grid.ParseOxmlFormula("=SUM(sheet1!B1:B10)")
	f8, _ := grid.ParseOxmlFormula("=AVERAGE(sheet1!B1:B10)")
	f9, _ := grid.ParseOxmlFormula("=MIN(sheet2!B1:B2)")
	f10, _ := grid.ParseOxmlFormula("=MAX(sheet2!B1:B2)")

	sheet3.SetFormula(layout.NewPosition(1, 2), f7)
	sheet3.SetFormula(layout.NewPosition(2, 2), f8)
	sheet3.SetFormula(layout.NewPosition(3, 2), f9)
	sheet3.SetFormula(layout.NewPosition(4, 2), f10)

	return flat.NewFileFromSheets(sheet1, sheet2, sheet3)
}
