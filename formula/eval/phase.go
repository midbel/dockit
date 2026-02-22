package eval

import (
	"fmt"

	"github.com/midbel/dockit/formula/parse"
)

type scriptPhase int8

const (
	phaseStmt scriptPhase = 1 << iota
	phaseImport
	phaseBinary
	phaseCall
	phaseAssign
)

func (p scriptPhase) Allows(k parse.Kind) bool {
	switch p {
	case phaseStmt:
		return k == parse.KindStmt
	case phaseImport:
		return k == parse.KindStmt || k == parse.KindImport
	default:
		return false
	}
}

func (p scriptPhase) Next(k parse.Kind) scriptPhase {
	switch {
	case p == phaseImport && k == parse.KindImport:
		return p
	case p == phaseImport && k == parse.KindStmt:
		return phaseStmt
	case p == phaseStmt && k == parse.KindStmt:
		return p
	default:
		return p
	}
}

func execPhase(expr parse.Expr, phase scriptPhase) (scriptPhase, error) {
	currKind := parse.KindStmt
	if ek, ok := expr.(parse.ExprKind); ok {
		currKind = ek.Kind()
	}
	if !phase.Allows(currKind) {
		return phase, fmt.Errorf("unknown script phase!")
	}
	return phase.Next(currKind), nil
}
