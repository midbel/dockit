package gridx

import (
	"slices"
	"strings"
	"testing"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/internal/testutil"
	"github.com/midbel/dockit/layout"
)

const groupSample = `go,foo,100,42
go,foo,50,0
go,bar,50,0
ts,bar,1,0
js,bar,5,1
js,foo,15,58
java,quz,10,12
go,bar,1000,3`

func TestGroup(t *testing.T) {
	r := csv.NewReader(strings.NewReader(groupSample))
	f, err := flat.OpenReader(r)
	if err != nil {
		t.Fatalf("fail to create file from sample: %s", err)
	}
	sh, err := f.ActiveSheet()
	if err != nil {
		t.Fatalf("fail to retrieve active sheet: %s", err)
	}
	var (
		keys, _ = layout.SelectionFromString("A")
		aggr    = []Aggr{
			*NewAggr(3, Sum()),
			*NewAggr(3, Min()),
			*NewAggr(3, Max()),
			*NewAggr(3, Avg()),
			*NewAggr(3, Count()),
		}
	)
	view, err := Group(sh, keys, aggr)
	if err != nil {
		t.Fatalf("error creating group view")
	}
	want := [][]string{
		{"java", "10", "10", "10", "10", "1"},
		{"go", "1200", "50", "1000", "300", "4"},
		{"ts", "1", "1", "1", "1", "1"},
		{"js", "20", "5", "15", "10", "2"},
	}
	got := testutil.Collect(view)

	slices.SortFunc(got, func(rs1, rs2 []string) int {
		return strings.Compare(rs1[0], rs2[0])
	})
	slices.SortFunc(want, func(rs1, rs2 []string) int {
		return strings.Compare(rs1[0], rs2[0])
	})
	if len(got) != len(want) {
		t.Fatalf("number of rows mismatched! want %d, got %d", len(want), len(got))
	}
	for i := 0; i < len(got); i++ {
		if len(got[i]) != len(want[i]) {
			t.Errorf("number of values mismatched! want %d, got %d", len(want[i]), len(got[i]))
		}
		for j := 0; j < len(got[i]); j++ {
			if got[i][j] != want[i][j] {
				t.Errorf("values mismatched! want %s, got %s", want[i][j], got[i][j])
			}
		}
	}
}
