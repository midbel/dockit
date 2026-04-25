package gridx

import (
	"strings"
	"testing"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/testutil"
	"github.com/midbel/dockit/layout"
)

const groupSample = `go,foo,100
go,foo,50
go,bar,50
ts,bar,1
js,bar,5
js,foo,15
java,quz,10
go,bar,1000`

const numCol = 3

func TestGroup(t *testing.T) {
	t.Run("single-key", testGroupSingleKey)
	t.Run("multi-key", testGroupMultiKey)
}

func testGroupMultiKey(t *testing.T) {
	var (
		want = [][]string{
			{"ts", "bar", "1", "1", "1", "1", "1"},
			{"js", "bar", "5", "5", "5", "5", "1"},
			{"js", "foo", "15", "15", "15", "15", "1"},
			{"java", "quz", "10", "10", "10", "10", "1"},
			{"go", "bar", "1050", "50", "1000", "525", "2"},
			{"go", "foo", "150", "50", "100", "75", "2"},
		}
		keys, _ = layout.SelectionFromString("A;B")
		view    = createGroupView(t, keys)
		got     = testutil.Collect(view)
	)

	testutil.AssertViewEqual(t, want, got, func(rs1, rs2 []string) int {
		res := strings.Compare(rs1[0], rs2[0])
		if res == 0 {
			res = strings.Compare(rs1[1], rs2[1])
		}
		return res
	})
}

func testGroupSingleKey(t *testing.T) {
	var (
		want = [][]string{
			{"java", "10", "10", "10", "10", "1"},
			{"go", "1200", "50", "1000", "300", "4"},
			{"ts", "1", "1", "1", "1", "1"},
			{"js", "20", "5", "15", "10", "2"},
		}
		keys, _ = layout.SelectionFromString("A")
		view    = createGroupView(t, keys)
		got     = testutil.Collect(view)
	)

	testutil.AssertViewEqual(t, want, got, func(rs1, rs2 []string) int {
		return strings.Compare(rs1[0], rs2[0])
	})
}

func createGroupView(t *testing.T, keys layout.Selection) grid.View {
	t.Helper()

	f, err := testutil.CreateCsvFile(strings.NewReader(groupSample))
	if err != nil {
		t.Fatalf("fail to create file from sample: %s", err)
	}
	sh, err := f.ActiveSheet()
	if err != nil {
		t.Fatalf("fail to retrieve active sheet: %s", err)
	}

	aggr := []Aggr{
		*NewAggr(numCol, Sum()),
		*NewAggr(numCol, Min()),
		*NewAggr(numCol, Max()),
		*NewAggr(numCol, Avg()),
		*NewAggr(numCol, Count()),
	}
	view, err := Group(sh, keys, aggr)
	if err != nil {
		t.Fatalf("error creating group view: %v", err)
	}
	return view
}
