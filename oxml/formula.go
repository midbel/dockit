package oxml

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

var ErrZero = errors.New("division by zero")

type Expr interface{}

type Value interface{}

func execSum(args []Value) (Value, error) {
	var total float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		total += args[i].(float64)
	}
	return total, nil
}

func execAvg(args []Value) (Value, error) {
	if len(args) == 0 {
		return 0, nil
	}
	var total float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		total += args[i].(float64)
	}
	return total / float64(len(args)), nil
}

func execCount(args []Value) (Value, error) {
	return nil, nil
}

var builtins = map[string]func([]Value) (Value, error){
	"sum":     execSum,
	"average": execAvg,
	"count":   execCount,
}

type Context interface {
	At(sheet string, row, col int) (Value, error)
}

type fileContext struct {
	*File
}

func (c fileContext) At(sheet string, row, col int) (Value, error) {
	if sheet == "" {

	}
	return nil, nil
}

func Eval(expr Expr, ctx Context) (Value, error) {
	switch e := expr.(type) {
	case binary:
		return evalBinary(e, ctx)
	case unary:
		return evalUnary(e, ctx)
	case literal:
		return e.value, nil
	case number:
		return e.value, nil
	case call:
		return evalCall(e, ctx)
	case cellAddr:
		return evalCellAddr(e, ctx)
	case rangeAddr:
		return nil, nil
	default:
		return nil, fmt.Errorf("unuspported expression type: %T", expr)
	}
}

func evalBinary(e binary, ctx Context) (Value, error) {
	left, err := Eval(e.left, ctx)
	if err != nil {
		return nil, err
	}
	right, err := Eval(e.right, ctx)
	if err != nil {
		return nil, err
	}
	switch e.op {
	case Add:
		return applyValues(left, right, func(left, right float64) (float64, error) {
			return left + right, nil
		})
	case Sub:
		return applyValues(left, right, func(left, right float64) (float64, error) {
			return left - right, nil
		})
	case Mul:
		return applyValues(left, right, func(left, right float64) (float64, error) {
			return left * right, nil
		})
	case Div:
		return applyValues(left, right, func(left, right float64) (float64, error) {
			if right == 0 {
				return 0, ErrZero
			}
			return left / right, nil
		})
	case Pow:
		return applyValues(left, right, func(left, right float64) (float64, error) {
			return math.Pow(left, right), nil
		})
	case Concat:
		if !isScalar(left) || !isScalar(right) {
			return nil, fmt.Errorf("expected scalar operand(s) for concat operator")
		}
		return fmt.Sprintf("%v%v", left, right), nil
	default:
		return nil, fmt.Errorf("invalid infix operator")
	}
}

func applyValues(left, right Value, do func(left, right float64) (float64, error)) (Value, error) {
	if !isNumber(left) {
		return nil, fmt.Errorf("expected numeric operands")
	}
	if !isNumber(right) {
		return nil, fmt.Errorf("expected numeric operands")
	}
	return do(left.(float64), right.(float64))
}

func isNumber(v Value) bool {
	_, ok := v.(float64)
	return ok
}

func isScalar(v Value) bool {
	switch v.(type) {
	case float64:
	case string:
	case bool:
	default:
		return false
	}
	return true
}

func evalUnary(e unary, ctx Context) (Value, error) {
	val, err := Eval(e.right, ctx)
	if err != nil {
		return nil, err
	}
	n, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("expected number")
	}
	switch e.op {
	case Add:
		return n, nil
	case Sub:
		return -n, nil
	default:
		return nil, fmt.Errorf("invalid prefix operator")
	}
}

func evalCall(e call, ctx Context) (Value, error) {
	id, ok := e.ident.(identifier)
	if !ok {
		return nil, fmt.Errorf("identifier expected")
	}
	var args []Value
	for _, a := range e.args {
		v, err := Eval(a, ctx)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}
	fn, ok := builtins[strings.ToLower(id.name)]
	if !ok {
		return nil, fmt.Errorf("%s: unsupported function", id.name)
	}
	return fn(args)

}

func evalCellAddr(e cellAddr, ctx Context) (Value, error) {
	return ctx.At(e.sheet, e.line, e.column)
}

type binary struct {
	left  Expr
	right Expr
	op    rune
}

type unary struct {
	right Expr
	op    rune
}

type literal struct {
	value string
}

type number struct {
	value float64
}

type call struct {
	ident Expr
	args  []Expr
}

type identifier struct {
	name string
}

type cellAddr struct {
	sheet   string
	column  int
	absCols bool
	line    int
	absLine bool
}

type rangeAddr struct {
	startAddr cellAddr
	endAddr   cellAddr
}

func parseCellAddr(addr string) (cellAddr, error) {
	var (
		pos    cellAddr
		offset int
	)
	if addr[offset] == dollar {
		pos.absCols = true
		offset++
	}
	for offset < len(addr) && isLetter(rune(addr[offset])) {
		delta := byte('A')
		if isLower(rune(addr[offset])) {
			delta = 'a'
		}
		pos.column = pos.column*26 + int(addr[offset]-delta+1)
		offset++
	}
	if offset < len(addr) && addr[offset] == dollar {
		pos.absLine = true
		offset++
	}
	if offset < len(addr) {
		pos.line, _ = strconv.Atoi(addr[offset:])
	}
	return pos, nil
}

const (
	powLowest = iota
	powEq
	powCmp
	powConcat
	powAdd
	powMul
	powPow
	powUnary
	powPercent
	powCall
)

var bindings = map[rune]int{
	Add:     powAdd,
	Sub:     powAdd,
	Mul:     powMul,
	Div:     powMul,
	Percent: powPercent,
	Pow:     powPow,
	Concat:  powConcat,
	Eq:      powEq,
	Ne:      powEq,
	Lt:      powCmp,
	Le:      powCmp,
	Gt:      powCmp,
	Ge:      powCmp,
	BegGrp:  powCall,
}

type Parser struct {
	scan *Scanner
	curr Token
	peek Token

	prefix map[rune]func() (Expr, error)
	infix  map[rune]func(Expr) (Expr, error)
}

var defaultParser = Parse()

func parseFormula(str string) (Expr, error) {
	return defaultParser.ParseString(str)
}

func Parse() *Parser {
	var p Parser
	p.prefix = map[rune]func() (Expr, error){
		Ident:   p.parseAdressOrIdentifier,
		Number:  p.parseNumber,
		Literal: p.parseLiteral,
		Sub:     p.parseUnary,
		Add:     p.parseUnary,
		BegGrp:  p.parseGroup,
	}
	p.infix = map[rune]func(Expr) (Expr, error){
		BegGrp: p.parseCall,
		Add:    p.parseBinary,
		Sub:    p.parseBinary,
		Mul:    p.parseBinary,
		Div:    p.parseBinary,
		Concat: p.parseBinary,
		Pow:    p.parseBinary,
		Eq:     p.parseBinary,
		Ne:     p.parseBinary,
		Lt:     p.parseBinary,
		Le:     p.parseBinary,
		Gt:     p.parseBinary,
		Ge:     p.parseBinary,
	}
	return &p
}

func (p *Parser) ParseString(str string) (Expr, error) {
	return p.Parse(strings.NewReader(str))
}

func (p *Parser) Parse(r io.Reader) (Expr, error) {
	scan, err := Scan(r)
	if err != nil {
		return nil, err
	}
	p.scan = scan
	p.next()
	p.next()
	return p.parse(powLowest)
}

func (p *Parser) parse(pow int) (Expr, error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("unexpected prefix")
	}
	left, err := fn()
	if err != nil {
		return nil, err
	}
	for !p.done() && pow < bindings[p.curr.Type] {
		fn, ok := p.infix[p.curr.Type]
		if !ok {
			return nil, fmt.Errorf("unexpected infix operator")
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) parseCall(expr Expr) (Expr, error) {
	p.next()
	c := call{
		ident: expr,
	}
	for !p.done() && p.curr.Type != EndGrp {
		arg, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		switch p.curr.Type {
		case Comma:
			p.next()
		case EndGrp:
		default:
			return nil, fmt.Errorf("unexpected character in function call")
		}
		c.args = append(c.args, arg)
	}
	if p.curr.Type != EndGrp {
		return nil, fmt.Errorf("unexpected character in function call")
	}
	p.next()
	return c, nil
}

func (p *Parser) parseBinary(left Expr) (Expr, error) {
	bin := binary{
		left: left,
		op:   p.curr.Type,
	}
	p.next()
	right, err := p.parse(bindings[bin.op])
	if err != nil {
		return nil, err
	}
	bin.right = right
	return bin, nil
}

func (p *Parser) parseUnary() (Expr, error) {
	una := unary{
		op: p.curr.Type,
	}
	p.next()
	right, err := p.parse(powUnary)
	if err != nil {
		return nil, err
	}
	una.right = right
	return una, nil
}

func (p *Parser) parseGroup() (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if p.curr.Type != EndGrp {
		return nil, fmt.Errorf("missing ')' at end of expression")
	}
	p.next()
	return expr, nil
}

func (p *Parser) parseNumber() (Expr, error) {
	defer p.next()

	x, err := strconv.ParseFloat(p.curr.Literal, 64)
	if err != nil {
		return nil, err
	}
	n := number{
		value: x,
	}
	return n, nil
}

func (p *Parser) parseLiteral() (Expr, error) {
	defer p.next()
	i := literal{
		value: p.curr.Literal,
	}
	return i, nil
}

func (p *Parser) parseAdressOrIdentifier() (Expr, error) {
	defer p.next()
	if p.peek.Type == BegGrp {
		i := identifier{
			name: p.curr.Literal,
		}
		return i, nil
	} 
	var sheet string
	if p.peek.Type == SheetRef {
		sheet = p.curr.Literal
		p.next()
		p.next()
	}
	a, err := parseCellAddr(p.curr.Literal)
	if err != nil {
		return nil, err
	}
	if p.curr.Type == RangeRef {
		p.next()
	}
	a.sheet = sheet
	return a, nil
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *Parser) done() bool {
	return p.curr.Type == Eol
}

const (
	Invalid rune = -(1 << iota)
	Eol
	Ident
	Number
	Literal
	Add
	Sub
	Mul
	Div
	Percent
	Pow
	Concat
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
	Comma
	BegGrp
	EndGrp
	BegArr
	EndArr
	RangeRef
	SheetRef
)

type Token struct {
	Literal  string
	Type     rune
	Position int
}

type Scanner struct {
	input []byte
	pos   int
	next  int
	char  rune

	buf bytes.Buffer
}

func Scan(r io.Reader) (*Scanner, error) {
	var (
		scan Scanner
		err  error
	)
	scan.input, err = io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	scan.read()
	if scan.char == equal {
		scan.read()
	}
	// scan.read()
	return &scan, nil
}

func (s *Scanner) Scan() Token {
	s.skipBlanks()

	var tok Token
	if s.done() {
		tok.Type = Eol
		return tok
	}
	defer s.reset()
	switch {
	case isOperator(s.char):
		s.scanOperator(&tok)
	case isDelimiter(s.char):
		s.scanDelimiter(&tok)
	case isQuote(s.char):
		s.scanLiteral(&tok)
	case isDigit(s.char):
		s.scanNumber(&tok)
	default:
		s.scanIdent(&tok)
	}
	return tok
}

func (s *Scanner) scanIdent(tok *Token) {
	for !s.done() && isAlpha(s.char) {
		s.write()
		s.read()
	}
	tok.Type = Ident
	tok.Literal = s.literal()
}

func (s *Scanner) scanNumber(tok *Token) {
	tok.Type = Number
	for !s.done() && isDigit(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.literal()
	if s.char != dot {
		return
	}
	s.write()
	s.read()
	for !s.done() && isDigit(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.literal()
}

func (s *Scanner) scanLiteral(tok *Token) {
	quote := s.char
	s.read()
	for !s.done() && !isQuote(s.char) {
		s.write()
		s.read()
	}
	tok.Type = Literal
	tok.Literal = s.literal()
	if isQuote(s.char) && quote == s.char {
		s.read()
	} else {
		tok.Type = Invalid
	}
}

func (s *Scanner) scanOperator(tok *Token) {
	tok.Type = Invalid
	switch s.char {
	case amper:
		tok.Type = Concat
	case percent:
		tok.Type = Percent
	case plus:
		tok.Type = Add
	case minus:
		tok.Type = Sub
	case star:
		tok.Type = Mul
	case slash:
		tok.Type = Div
	case caret:
		tok.Type = Pow
	case langle:
		tok.Type = Lt
		s.read()
		if s.char == equal {
			tok.Type = Le
		} else if s.char == rangle {
			tok.Type = Ne
		}
	case rangle:
		tok.Type = Gt
		s.read()
		if s.char == equal {
			tok.Type = Ge
		}
	case equal:
		tok.Type = Eq
	case colon:
		tok.Type = RangeRef
	case bang:
		tok.Type = SheetRef
	default:
	}
	s.read()
}

func (s *Scanner) scanDelimiter(tok *Token) {
	tok.Type = Invalid
	switch s.char {
	case semi, comma:
		tok.Type = Comma
	case lparen:
		tok.Type = BegGrp
	case rparen:
		tok.Type = EndGrp
	case lcurly:
		tok.Type = BegArr
	case rcurly:
		tok.Type = EndArr
	default:
	}
	s.read()
}

func (s *Scanner) literal() string {
	return s.buf.String()
}

func (s *Scanner) write() {
	s.buf.WriteRune(s.char)
}

func (s *Scanner) reset() {
	s.buf.Reset()
}

func (s *Scanner) read() {
	if s.pos >= len(s.input) {
		s.char = 0
		return
	}
	r, n := utf8.DecodeRune(s.input[s.next:])
	if r == utf8.RuneError {
		s.char = 0
		s.next = len(s.input)
	}
	s.char, s.pos, s.next = r, s.next, s.next+n
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.input[s.next:])
	return r
}

func (s *Scanner) done() bool {
	return s.pos >= len(s.input) && s.char == 0
}

func (s *Scanner) skipBlanks() {
	for isBlank(s.char) {
		s.read()
	}
}

const (
	underscore = '_'
	bang       = '!'
	semi       = ';'
	comma      = ','
	rparen     = ')'
	lparen     = '('
	lcurly     = '{'
	rcurly     = '}'
	squote     = '\''
	dquote     = '"'
	space      = ' '
	tab        = '\t'
	plus       = '+'
	minus      = '-'
	star       = '*'
	slash      = '/'
	caret      = '^'
	equal      = '='
	langle     = '<'
	rangle     = '>'
	colon      = ':'
	dot        = '.'
	amper      = '&'
	percent    = '%'
	dollar     = '$'
)

func isQuote(c rune) bool {
	return c == squote || c == dquote
}

func isLower(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func isLetter(c rune) bool {
	return isLower(c) || isUpper(c) || c == underscore
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c rune) bool {
	return isLetter(c) || isDigit(c) || c == dollar
}

func isBlank(c rune) bool {
	return c == space || c == tab
}

func isDelimiter(c rune) bool {
	return c == semi || c == lparen || c == rparen ||
		c == lcurly || c == rcurly || c == comma
}

func isOperator(c rune) bool {
	return c == plus || c == minus || c == slash || c == star ||
		c == langle || c == rangle || c == colon || c == bang ||
		c == equal || c == caret || c == amper || c == percent
}
