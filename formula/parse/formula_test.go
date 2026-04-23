package parse

import (
	"testing"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/layout"
)

func TestParseFormula(t *testing.T) {
	t.Run("oxml", testParseOxmlFormula)
	t.Run("ods", testParseOdsFormula)
}

func testParseOdsFormula(t *testing.T) {
	tests := []struct {
		Expr string
		Want Expr
	}{
		{
			Expr: "of:=1+2",
			Want: NewBinary(
				NewNumber(1),
				NewNumber(2),
				op.Add,
			),
		},
		// {
		// 	Expr: "of:=[Sheet1.A1]",
		// 	Want: nil,
		// },
		// {
		// 	Expr: "of:=SUM([.A1:.A10])",
		// 	Want: nil,
		// },
		// {
		// 	Expr: "of:=[.A1]+[.B1]",
		// 	Want: nil,
		// },
		// {
		// 	Expr: "of:=CONCATENATE([.A1];" ";[.B1])",
		// 	Want: nil,
		// },
	}
	for _, c := range tests {
		f, err := ParseOdsFormula(c.Expr)
		if err != nil {
			t.Errorf("%s: error parsing ODS formumla: %s", c.Expr, err)
			continue
		}
		assertEqualExpr(t, c.Want, f)
	}
}

func testParseOxmlFormula(t *testing.T) {
	tests := []struct {
		Expr string
		Want Expr
	}{
		{
			Expr: "=A+1",
			Want: NewBinary(
				NewIdentifier("A"),
				NewNumber(1),
				op.Add,
			),
		},
		{
			Expr: "=sheet!A + 1",
			Want: NewBinary(
				NewCellAccess(
					NewIdentifier("sheet"),
					NewCellAddr(layout.NewPosition(0, 1), false, false),
				),
				NewNumber(1),
				op.Add,
			),
		},
		{
			Expr: "=A:C + 1",
			Want: NewBinary(
				NewRangeAddr(
					NewCellAddr(layout.NewPosition(0, 1), false, false),
					NewCellAddr(layout.NewPosition(0, 3), false, false),
				),
				NewNumber(1),
				op.Add,
			),
		},
		{
			Expr: "=10:25 + 1",
			Want: NewBinary(
				NewRangeAddr(
					NewCellAddr(layout.NewPosition(10, 0), false, false),
					NewCellAddr(layout.NewPosition(25, 0), false, false),
				),
				NewNumber(1),
				op.Add,
			),
		},
		{
			Expr: "=1+1",
			Want: NewBinary(
				NewNumber(1),
				NewNumber(1),
				op.Add,
			),
		},
		{
			Expr: "=sum(A1, A2, A3)",
			Want: NewCall(
				NewIdentifier("sum"),
				[]Expr{
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewCellAddr(layout.NewPosition(2, 1), false, false),
					NewCellAddr(layout.NewPosition(3, 1), false, false),
				},
			),
		},
		{
			Expr: "=A1 + sheet2!A2",
			Want: NewBinary(
				NewCellAddr(layout.NewPosition(1, 1), false, false),
				NewCellAccess(
					NewIdentifier("sheet2"),
					NewCellAddr(layout.NewPosition(2, 1), false, false),
				),
				op.Add,
			),
		},
	}
	for _, c := range tests {
		f, err := ParseOxmlFormula(c.Expr)
		if err != nil {
			t.Errorf("%s: error parsing OXML formumla: %s", c.Expr, err)
			continue
		}
		assertEqualExpr(t, c.Want, f)
	}
}
