package oxml

import (
	"fmt"
	"testing"
)

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
	}
	p := Parse()
	for _, c := range tests {
		expr, err := p.ParseString(c.Formula)
		if err != nil {
			t.Errorf("%s: error parsing formula: %s", c.Formula, err)
			continue
		}
		value, err := Eval(expr)
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
