package parse

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/midbel/dockit/formula/op"
)

func FuzzScannerRandomFormula(f *testing.F) {
	for i := 0; i < 50; i++ {
		f.Add(createRandomFormula())
	}
	f.Fuzz(func(t *testing.T, input string) {
		scan, err := Scan(strings.NewReader(input), ScanFormula)
		if err != nil {
			panic(initScanner(err))
		}
		validInput := isExpectedValid(input)
		for {
			tok := scan.Scan()
			if tok.Type == op.EOF {
				break
			}
			if tok.Type == op.Invalid && validInput {
				t.Errorf("invalid token generated for valid input")
			}
		}
	})
}

func isExpectedValid(input string) bool {
	if len(input) == 0 {
		return false
	}
	if strings.ContainsAny(input, "#\\") {
		return false
	}
	var (
		squotes int
		dquotes int
	)
	for i := range input {
		if input[i] == '"' {
			dquotes++
		}
		if input[i] == '\'' {
			squotes++
		}
	}
	if dquotes%2 != 0 || squotes%2 != 0 {
		return false
	}
	return true
}

func createRandomFormula() string {
	var (
		cells = []string{"A1", "B10", "C3"}
		ops   = []string{"+", "-", "*", "/"}
		nums  = []string{"1", "42", "100"}
	)

	formula := "=" + cells[rand.Intn(len(cells))]
	for i := 0; i < rand.Intn(5)+1; i++ { // add 1-3 operations
		formula += ops[rand.Intn(len(ops))] + cells[rand.Intn(len(cells))]
	}
	formula += nums[rand.Intn(len(nums))]
	return formula
}

func FuzzScannerFormula2(f *testing.F) {
	f.Add("=A1 + 1")
	f.Add("=42/0")
	f.Add("=B100 - D873")
	f.Add("=B12 - ($C$1 * 9)")
	f.Add("=avg(F89:F167)")
	f.Add("=A1 + sum(A3, 5, B10:C100)")
	f.Add("=A1 + min(A3, B100) ^ (10 - 2)")
	f.Add("=$AA$143")
	f.Add("=100 / ($B1 ^ 5) * (42 - P9)")
	f.Add("=upper('hello' & ' ' & \"world\")")
	f.Add("=\"this is a \"\"test\"\"\"")

	f.Fuzz(func(t *testing.T, input string) {
		if strings.ContainsAny(input, "\\") {
			t.Skip()
		}
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
				t.Errorf("invalid token generated for valid input")
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
