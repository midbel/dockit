package eval

import (
	"fmt"

	"github.com/midbel/dockit/formula/op"
)

const (
	powLowest = iota
	powAssign
	powEq
	powCmp
	powConcat
	powAdd
	powMul
	powPow
	powUnary
	powPercent
	powProp
	powCall
)

var defaultBindings = map[op.Op]int{
	op.AddAssign:    powAssign,
	op.SubAssign:    powAssign,
	op.MulAssign:    powAssign,
	op.DivAssign:    powAssign,
	op.PowAssign:    powAssign,
	op.ConcatAssign: powAssign,
	op.Assign:       powAssign,
	op.Add:          powAdd,
	op.Sub:          powAdd,
	op.Mul:          powMul,
	op.Div:          powMul,
	op.Percent:      powPercent,
	op.Pow:          powPow,
	op.Concat:       powConcat,
	op.Eq:           powEq,
	op.Ne:           powEq,
	op.Lt:           powCmp,
	op.Le:           powCmp,
	op.Gt:           powCmp,
	op.Ge:           powCmp,
	op.BegGrp:       powCall,
	op.BegProp:      powProp,
	op.Dot:          powProp,
}

type (
	PrefixFunc func(*Parser) (Expr, error)
	InfixFunc  func(*Parser, Expr) (Expr, error)
)

var errForbidden = fmt.Errorf("not allowed")

func forbiddenInfix(_ *Parser, _ Expr) (Expr, error) {
	return nil, fmt.Errorf("%w: infix operator/keyword", errForbidden)
}

func forbiddenPrefix(_ *Parser) (Expr, error) {
	return nil, fmt.Errorf("%w: prefix operator/keyword", errForbidden)
}

type Grammar struct {
	name string
	mode ScanMode

	prefix   map[op.Op]PrefixFunc
	infix    map[op.Op]InfixFunc
	postfix  map[op.Op]InfixFunc
	bindings map[op.Op]int

	kwPrefix map[string]PrefixFunc
	kwInfix  map[string]InfixFunc
}

func (g *Grammar) Context() string {
	return g.name
}

func (g *Grammar) Pow(kind op.Op) int {
	pow, ok := g.bindings[kind]
	if !ok {
		pow = powLowest
	}
	return pow
}

func (g *Grammar) Prefix(tok Token) (PrefixFunc, error) {
	if tok.Type == op.Keyword {
		fn, ok := g.kwPrefix[tok.Literal]
		if !ok {
			return nil, fmt.Errorf("(%s) %s: unknown keyword %s", tok.Position, g.name, tok.Literal)
		}
		return fn, nil
	}
	fn, ok := g.prefix[tok.Type]
	if !ok {
		return nil, fmt.Errorf("(%s) %s: unsupported prefix operator (%s)", tok.Position, g.name, tok)
	}
	return fn, nil
}

func (g *Grammar) Infix(tok Token) (InfixFunc, error) {
	if tok.Type == op.Keyword {
		fn, ok := g.kwInfix[tok.Literal]
		if !ok {
			return nil, fmt.Errorf("(%s) %s: unknown keyword %s", tok.Position, g.name, tok.Literal)
		}
		return fn, nil
	}
	fn, ok := g.infix[tok.Type]
	if !ok {
		return nil, fmt.Errorf("(%s) %s: unsupported infix operator (%s)", tok.Position, g.name, tok)
	}
	return fn, nil
}

func (g *Grammar) Postfix(tok Token) (InfixFunc, error) {
	fn, ok := g.postfix[tok.Type]
	if !ok {
		return nil, fmt.Errorf("(%s) %s: unsupported postfix operator (%s)", tok.Position, g.name, tok)
	}
	return fn, nil
}

func (g *Grammar) RegisterInfix(kd op.Op, fn InfixFunc) {
	g.infix[kd] = fn
}

func (g *Grammar) UnregisterInfix(kd op.Op) {
	g.infix[kd] = forbiddenInfix
}

func (g *Grammar) RegisterPostfix(kd op.Op, fn InfixFunc) {
	g.postfix[kd] = fn
}

func (g *Grammar) UnregisterPostfix(kd op.Op) {
	g.postfix[kd] = forbiddenInfix
}

func (g *Grammar) RegisterInfixKeyword(kw string, fn InfixFunc) {
	g.kwInfix[kw] = fn
}

func (g *Grammar) UnregisterInfixKeyword(kw string) {
	g.kwInfix[kw] = forbiddenInfix
}

func (g *Grammar) RegisterPrefix(kd op.Op, fn PrefixFunc) {
	g.prefix[kd] = fn
}

func (g *Grammar) UnregisterPrefix(kd op.Op) {
	g.prefix[kd] = forbiddenPrefix
}

func (g *Grammar) RegisterPrefixKeyword(kw string, fn PrefixFunc) {
	g.kwPrefix[kw] = fn
}

func (g *Grammar) UnregisterPrefixKeyword(kw string) {
	g.kwPrefix[kw] = forbiddenPrefix
}

func (g *Grammar) RegisterBinding(kd op.Op, pow int) {
	g.bindings[kd] = pow
}

type GrammarStack []*Grammar

func (gs *GrammarStack) Top() *Grammar {
	n := len(*gs)
	return (*gs)[n-1]
}

func (gs *GrammarStack) Mode() ScanMode {
	return gs.Top().mode
}

func (gs *GrammarStack) Context() string {
	return gs.Top().Context()
}

func (gs *GrammarStack) Pow(kind op.Op) int {
	for i := len(*gs) - 1; i >= 0; i-- {
		pow := (*gs)[i].Pow(kind)
		if pow > powLowest {
			return pow
		}
	}
	return powLowest
}

func (gs *GrammarStack) Prefix(tok Token) (PrefixFunc, error) {
	var lastErr error
	for i := len(*gs) - 1; i >= 0; i-- {
		fn, err := (*gs)[i].Prefix(tok)
		if err == nil {
			return fn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (gs *GrammarStack) Infix(tok Token) (InfixFunc, error) {
	var lastErr error
	for i := len(*gs) - 1; i >= 0; i-- {
		fn, err := (*gs)[i].Infix(tok)
		if err == nil {
			return fn, err
		}
		lastErr = err
	}
	return nil, lastErr
}

func (gs *GrammarStack) Postfix(tok Token) (InfixFunc, error) {
	var lastErr error
	for i := len(*gs) - 1; i >= 0; i-- {
		fn, err := (*gs)[i].Postfix(tok)
		if err == nil {
			return fn, err
		}
		lastErr = err
	}
	return nil, lastErr
}

func (gs *GrammarStack) Pop() {
	n := len(*gs)
	if n > 1 {
		*gs = (*gs)[:n-1]
	}
}

func (gs *GrammarStack) Push(g *Grammar) {
	*gs = append(*gs, g)
}
