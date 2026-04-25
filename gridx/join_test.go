package gridx

import (
	"strings"
	"testing"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/testutil"
	"github.com/midbel/dockit/layout"
)

const leftJoinSample = `1,go,100
2,js,50
3,c,20`
const rightJoinSample = `1,1,dockit,github.com/midbel/dockit
2,1,sweet,github.com/midbel/sweet-ql
3,2,dockit-ui,github.com/midbel/dockit-ui`

func TestJoin(t *testing.T) {
	t.Run("single-key", testJoinSingleKey)
}

func testJoinSingleKey(t *testing.T) {
	var (
		v1      = getJoinView(t, leftJoinSample)
		v2      = getJoinView(t, rightJoinSample)
		key1, _ = layout.SelectionFromString("A")
		key2, _ = layout.SelectionFromString("B")
	)

	view := Join(v1, v2, key1, key2)

	want := [][]string{
		{"1", "go", "100", "1", "1", "dockit", "github.com/midbel/dockit"},
		{"1", "go", "100", "2", "1", "sweet", "github.com/midbel/sweet-ql"},
		{"2", "js", "50", "3", "2", "dockit-ui", "github.com/midbel/dockit-ui"},
	}
	got := testutil.Collect(view)
	testutil.AssertSize(t, view, got)
	testutil.AssertViewEqual(t, want, got, nil)
}

func getJoinView(t *testing.T, content string) grid.View {
	t.Helper()

	f, err := testutil.CreateCsvFile(strings.NewReader(content))
	if err != nil {
		t.Fatalf("fail to create file from sample: %s", err)
	}
	sh, err := f.ActiveSheet()
	if err != nil {
		t.Fatalf("fail to retrieve active sheet: %s", err)
	}
	return sh
}
