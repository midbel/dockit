package parse

import (
	"fmt"
	"strings"
	"testing"

	"github.com/midbel/dockit/formula/op"
)

func FuzzScannerFormula(f *testing.F) {
	f.Add("=A1 + 1")
	f.Add("=A1 + sum(A3, 5, B10:C100)")
	f.Add("=A1 + min(A3, B100) ^ (10 - 2)")
	f.Add("=#AA143")
	f.Add("upper('hello' & ' ' & 'word')")
	f.Add("=A1 ++ 1")
	f.Add("=sum(,)")
	f.Add("=min(A1, )")

	f.Fuzz(func(t *testing.T, input string) {
		scan, err := Scan(strings.NewReader(input), ScanFormula)
		if err != nil {
			panic(initScanner(err))
		}
		for {
			tok := scan.Scan()
			if tok.Type == op.EOF {
				break
			}
			if tok.Type == op.Invalid {
				panic(invalidToken(tok))
			}
		}
	})
}

func FuzzScannerScript(f *testing.F) {
	f.Add("foo := 100; A1 += foo;;")
	f.Add("(view1[A;B;C] & view2[C:10]) | view3[D1 >= 100 and A1 = 'test']")

	f.Fuzz(func(t *testing.T, input string) {
		scan, err := Scan(strings.NewReader(input), ScanFormula)
		if err != nil {
			panic(initScanner(err))
		}
		for {
			tok := scan.Scan()
			if tok.Type == op.EOF {
				break
			}
			if tok.Type == op.Invalid {
				panic(invalidToken(tok))
			}
		}
	})
}

func invalidToken(tok Token) error {
	return fmt.Errorf("invalid token generated %s", tok)
}

func initScanner(err error) error {
	return fmt.Errorf("error init scanning: %s", err)
}

func initParser(err error) error {
	return fmt.Errorf("error init parser: %s", err)
}

// func testScannerInvalid(t *testing.T, mode ScanMode, input string) {
// 	t.Logf("try scanning: %s", input)
// 	scan, err := Scan(strings.NewReader(input), mode)
// 	if err != nil {
// 		t.Errorf("error init scanner: %s", err)
// 		return
// 	}
// 	for {
// 		tok := scan.Scan()
// 		if tok.Type == op.EOF {
// 			break
// 		}
// 		if tok.Type == op.Invalid {
// 			t.Errorf("invalid token generated: %s", tok)
// 		}
// 	}
// }
