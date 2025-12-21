package oxml

import (
	"fmt"
	"testing"
)

type MockSheet struct{}

func (m MockSheet) At(row int, column int) (Value, error) {
	value := fmt.Sprintf("%d%d", column, row)
	return value, nil
}

type TestCase struct {
	Formula string
	Want    string
}

func TestBasic(t *testing.T) {
	tests := []TestCase{
		{
			Formula: "='test'",
			Want:    "test",
		},
		{
			Formula: "=100",
			Want:    "100",
		},
		{
			Formula: "=+100",
			Want:    "100",
		},
		{
			Formula: "=-100",
			Want:    "-100",
		},
		{
			Formula: "=100+1",
			Want:    "101",
		},
		{
			Formula: "=100-1",
			Want:    "99",
		},
		{
			Formula: "=$A1",
			Want:    "11",
		},
		{
			Formula: "=$A1 & B1",
			Want:    "1121",
		},
		{
			Formula: "=SUM(1, 2, 3)",
			Want:    "6",
		},
		{
			Formula: "=AVERAGE(1, 2, 3)",
			Want:    "2",
		},
	}
	var (
		p = Parse()
		k MockSheet
	)
	for _, c := range tests {
		expr, err := p.ParseString(c.Formula)
		if err != nil {
			t.Errorf("%s: error parsing formula: %s", c.Formula, err)
			continue
		}
		value, err := Eval(expr, k)
		if err != nil {
			t.Errorf("%s: error evaluating formula: %s", c.Formula, err)
			continue
		}
		got := fmt.Sprint(value)
		if got != c.Want {
			t.Errorf("unexpected result! want %s, got %s", c.Want, got)
		}
	}
}
