package parse

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/midbel/dockit/formula/op"
)

type Scanner interface {
	Scan() Token
	Peek() Token
	Script() bool
	Type() DialectType
}

func ScanScript(r io.Reader) (Scanner, error) {
	rs, err := prepare(r)
	if err != nil {
		return nil, err
	}
	scan := ScriptLexer{
		reader: rs,
	}
	return &scan, nil
}

func ScanDialect(r io.Reader, dialect Dialect) (Scanner, error) {
	rs, err := prepare(r)
	if err != nil {
		return nil, err
	}
	scan := FormulaLexer{
		reader:  rs,
		dialect: dialect,
	}
	if err := scan.dialect.Init(&scan); err != nil {
		return nil, err
	}
	if scan.is(equal) {
		scan.read()
	}
	return &scan, nil
}

func ScanFormula(r io.Reader) (Scanner, error) {
	return ScanDialect(r, Oxml)
}

func ScanOxml(r io.Reader) (Scanner, error) {
	return ScanDialect(r, Oxml)
}

func ScanOds(r io.Reader) (Scanner, error) {
	return ScanDialect(r, Ods)
}

type ScriptLexer struct {
	*reader
}

func (x *ScriptLexer) Script() bool {
	return true
}

func (x *ScriptLexer) Type() DialectType {
	return TypeOxml
}

func (x *ScriptLexer) Scan() Token {
	if x.is(backslash) && isNL(x.peek()) {
		x.read()
		x.read()
	}
	x.skipBlanks()

	var tok Token
	tok.Position = x.Position
	if x.done() {
		tok.Type = op.EOF
		return tok
	}
	defer x.reset()

	switch {
	case isNL(x.char):
		x.scanNL(&tok)
	case isComment(x.char) && x.peek() != bang:
		x.scanComment(&tok)
	case isComment(x.char) && x.peek() == bang:
		x.scanDirective(&tok)
	case isOperator(x.char):
		x.scanOperator(&tok)
	case isDelimiter(x.char):
		x.scanDelimiter(&tok)
	case isQuote(x.char):
		x.scanLiteral(&tok)
	case isDigit(x.char):
		x.scanNumber(&tok)
	default:
		x.scanIdent(&tok)
	}
	return tok
}

func (x *ScriptLexer) Peek() Token {
	currState := x.Save()
	defer x.Restore(currState)
	return x.Scan()
}

func (x *ScriptLexer) scanNL(tok *Token) {
	x.SkipNL()
	tok.Type = op.Eol
}

func (x *ScriptLexer) scanComment(tok *Token) {
	x.read()
	x.skipBlanks()
	for !x.done() && !isNL(x.char) {
		x.write()
		x.read()
	}
	x.SkipNL()
	tok.Type = op.Comment
	tok.Literal = x.literal()
}

func (x *ScriptLexer) scanDirective(tok *Token) {
	x.read()
	x.read()
	if x.char == space {
		x.skipBlanks()
		tok.Type = op.Pragma
		return
	}
	for !x.done() && !isNL(x.char) {
		x.write()
		x.read()
	}
	tok.Type = op.Directive
	tok.Literal = x.literal()
}

func (x *ScriptLexer) scanIdent(tok *Token) {
	reco := recognizeCell()
	for !x.done() && isAlpha(x.char) {
		reco.Update(x.char)
		x.write()
		x.read()
	}
	tok.Type = op.Ident
	tok.Literal = x.literal()
	if reco.IsCell() {
		tok.Type = op.Cell
	}
	if reco.IsCol() {
		tok.Type = op.Column
	}

	if isKeyword(tok.Literal) {
		tok.Type = op.Keyword
		if tok.Literal == kwAnd {
			tok.Type = op.And
		} else if tok.Literal == kwOr {
			tok.Type = op.Or
		} else if tok.Literal == kwNot {
			tok.Type = op.Not
		}
	}
}

func (x *ScriptLexer) scanNumber(tok *Token) {
	tok.Type = op.Number
	for !x.done() && isDigit(x.char) {
		x.write()
		x.read()
	}
	tok.Literal = x.literal()
	if x.char == dot {
		x.write()
		x.read()
		for !x.done() && isDigit(x.char) {
			x.write()
			x.read()
		}
		tok.Literal = x.literal()
	}
	if x.char == 'e' || x.char == 'E' {
		x.write()
		x.read()
		if isSign(x.char) {
			x.write()
			x.read()
		}
		for !x.done() && isDigit(x.char) {
			x.write()
			x.read()
		}
		tok.Literal = x.literal()
	}
}

func (x *ScriptLexer) scanLiteral(tok *Token) {
	quote := x.char
	x.read()
	for !x.done() && !isQuote(x.char) {
		x.write()
		x.read()
		if x.char == dquote && x.peek() == x.char {
			x.write()
			x.read()
			x.read()
		}
	}
	tok.Type = op.Literal
	tok.Literal = x.literal()
	if isQuote(x.char) && quote == x.char {
		x.read()
	} else {
		tok.Type = op.Invalid
	}
}

func (x *ScriptLexer) scanOperator(tok *Token) {
	tok.Type = op.Invalid
	switch x.char {
	case arobase:
		tok.Type = op.Special
	case dot:
		tok.Type = op.Dot
	case pipe:
		tok.Type = op.Union
	case amper:
		tok.Type = op.Concat
		if x.peek() == equal {
			x.read()
			tok.Type = op.ConcatAssign
		}
	case percent:
		tok.Type = op.Percent
	case plus:
		tok.Type = op.Add
		if x.peek() == equal {
			x.read()
			tok.Type = op.AddAssign
		}
	case minus:
		tok.Type = op.Sub
		if x.peek() == equal {
			x.read()
			tok.Type = op.SubAssign
		}
	case star:
		tok.Type = op.Mul
		if x.peek() == equal {
			x.read()
			tok.Type = op.MulAssign
		}
	case slash:
		tok.Type = op.Div
		if x.peek() == equal {
			x.read()
			tok.Type = op.DivAssign
		}
	case caret:
		tok.Type = op.Pow
		if x.peek() == equal {
			x.read()
			tok.Type = op.PowAssign
		}
	case langle:
		tok.Type = op.Lt
		if k := x.peek(); k == equal {
			tok.Type = op.Le
		} else if k == rangle {
			tok.Type = op.Ne
		}
		if tok.Type != op.Lt {
			x.read()
		}
	case rangle:
		tok.Type = op.Gt
		if x.peek() == equal {
			x.read()
			tok.Type = op.Ge
		}
	case equal:
		tok.Type = op.Eq
	case colon:
		tok.Type = op.RangeRef
		if x.peek() == equal {
			x.read()
			tok.Type = op.Assign
		}
	case bang:
		tok.Type = op.SheetRef
	default:
	}
	x.read()
}

func (x *ScriptLexer) scanDelimiter(tok *Token) {
	tok.Type = op.Invalid
	switch x.char {
	case comma:
		tok.Type = op.Comma
	case semi:
		tok.Type = op.Semi
	case lparen:
		tok.Type = op.BegGrp
	case rparen:
		tok.Type = op.EndGrp
	case lsquare:
		tok.Type = op.BegProp
	case rsquare:
		tok.Type = op.EndProp
	default:
	}
	x.read()
}

type DialectType int8

const (
	TypeOxml DialectType = iota
	TypeOds
)

type Dialect interface {
	Init(*FormulaLexer) error
	Type() DialectType

	IsOperator(rune) bool
	IsQuote(rune) bool
	IsDelimiter(rune) bool
	IsDigit(rune) bool
	IsAddress(rune) bool
	IsError(rune) bool

	ScanOperator(*FormulaLexer, *Token)
	ScanDelimiter(*FormulaLexer, *Token)
	ScanLiteral(*FormulaLexer, *Token)
	ScanNumber(*FormulaLexer, *Token)
	ScanIdentifier(*FormulaLexer, *Token)
	ScanError(*FormulaLexer, *Token)
	ScanAddress(*FormulaLexer, *Token)
}

type oxmlDialect struct{}

func (d oxmlDialect) Init(sc *FormulaLexer) error {
	return nil
}

func (oxmlDialect) Type() DialectType {
	return TypeOxml
}

func (d oxmlDialect) IsOperator(ch rune) bool {
	return ch == plus || ch == minus || ch == slash || ch == star ||
		ch == langle || ch == rangle || ch == equal || ch == caret ||
		ch == amper || ch == percent || ch == bang || ch == colon
}

func (d oxmlDialect) IsQuote(ch rune) bool {
	return isQuote(ch)
}

func (d oxmlDialect) IsDelimiter(ch rune) bool {
	return isDelimiter(ch)
}

func (d oxmlDialect) IsDigit(ch rune) bool {
	return isDigit(ch)
}

func (d oxmlDialect) IsAddress(ch rune) bool {
	return ch == dollar
}

func (d oxmlDialect) IsError(ch rune) bool {
	return ch == pound
}

func (d oxmlDialect) ScanOperator(sc *FormulaLexer, tok *Token) {
	if sc.is(bang) {
		tok.Type = op.SheetRef
		sc.read()
		return
	}
	sc.scanOperator(tok)
}

func (d oxmlDialect) ScanDelimiter(sc *FormulaLexer, tok *Token) {
	sc.scanDelimiter(tok)
}

func (d oxmlDialect) ScanLiteral(sc *FormulaLexer, tok *Token) {
	sc.scanLiteral(tok)
}

func (d oxmlDialect) ScanNumber(sc *FormulaLexer, tok *Token) {
	sc.scanNumber(tok)
}

func (d oxmlDialect) ScanIdentifier(sc *FormulaLexer, tok *Token) {
	sc.scanIdent(tok)
}

func (d oxmlDialect) ScanError(sc *FormulaLexer, tok *Token) {
	sc.scanError(tok)
}

func (d oxmlDialect) ScanAddress(sc *FormulaLexer, tok *Token) {
	sc.scanIdent(tok)
}

type odsDialect struct{}

func (d odsDialect) Init(lex *FormulaLexer) error {
	str := lex.readUntil(colon, true)
	if str != "of" {
		return fmt.Errorf("invalid openformula namespace")
	}
	return nil
}

func (odsDialect) Type() DialectType {
	return TypeOds
}

func (d odsDialect) IsOperator(ch rune) bool {
	return ch == plus || ch == minus || ch == slash || ch == star ||
		ch == langle || ch == rangle || ch == equal || ch == caret ||
		ch == amper || ch == percent || ch == dot || ch == colon
}

func (d odsDialect) IsQuote(ch rune) bool {
	return isQuote(ch)
}

func (d odsDialect) IsDelimiter(ch rune) bool {
	return isDelimiter(ch)
}

func (d odsDialect) IsDigit(ch rune) bool {
	return isDigit(ch)
}

func (d odsDialect) IsAddress(ch rune) bool {
	return ch == dot || ch == lsquare || ch == dollar
}

func (d odsDialect) IsError(ch rune) bool {
	return ch == pound
}

func (d odsDialect) ScanOperator(sc *FormulaLexer, tok *Token) {
	var fallback bool
	switch {
	case sc.is(dot):
		tok.Type = op.SheetRef
	default:
		fallback = true
		sc.scanOperator(tok)
	}
	if !fallback {
		sc.read()
	}
}

func (d odsDialect) ScanDelimiter(sc *FormulaLexer, tok *Token) {
	var fallback bool
	switch {
	case sc.is(lsquare):
		tok.Type = op.BegAddr
	case sc.is(rsquare):
		tok.Type = op.EndAddr
	default:
		fallback = true
		sc.scanDelimiter(tok)
	}
	if !fallback {
		sc.read()
	}
}

func (d odsDialect) ScanLiteral(sc *FormulaLexer, tok *Token) {
	sc.scanLiteral(tok)
}

func (d odsDialect) ScanNumber(sc *FormulaLexer, tok *Token) {
	sc.scanNumber(tok)
}

func (d odsDialect) ScanIdentifier(sc *FormulaLexer, tok *Token) {
	sc.scanIdent(tok)
}

func (d odsDialect) ScanError(sc *FormulaLexer, tok *Token) {
	sc.scanError(tok)
}

func (d odsDialect) ScanAddress(sc *FormulaLexer, tok *Token) {

}

var (
	Ods  = odsDialect{}
	Oxml = oxmlDialect{}
)

type FormulaLexer struct {
	dialect Dialect
	*reader
}

func (s *FormulaLexer) Peek() Token {
	currState := s.Save()
	defer s.Restore(currState)
	return s.Scan()
}

func (s *FormulaLexer) Script() bool {
	return false
}

func (s *FormulaLexer) Type() DialectType {
	return s.dialect.Type()
}

func (s *FormulaLexer) Scan() Token {
	s.skipBlanks()

	var tok Token
	tok.Position = s.Position
	if s.done() {
		tok.Type = op.EOF
		return tok
	}
	defer s.reset()
	switch {
	case isNL(s.char):
		s.SkipNL()
		return s.Scan()
	case s.dialect.IsError(s.char):
		s.dialect.ScanError(s, &tok)
	case s.dialect.IsOperator(s.char):
		s.dialect.ScanOperator(s, &tok)
	case s.dialect.IsDelimiter(s.char):
		s.dialect.ScanDelimiter(s, &tok)
	case s.dialect.IsQuote(s.char):
		s.dialect.ScanLiteral(s, &tok)
	case s.dialect.IsDigit(s.char):
		s.dialect.ScanNumber(s, &tok)
	case s.dialect.IsAddress(s.char):
		s.dialect.ScanAddress(s, &tok)
	default:
		s.dialect.ScanIdentifier(s, &tok)
	}
	return tok
}

func (s *FormulaLexer) scanError(tok *Token) {
	s.write()
	s.read()
	accept := func(ch rune) bool {
		return isUpper(ch) || ch == '0' ||
			ch == question || ch == bang || ch == slash
	}
	for !s.done() && accept(s.char) {
		s.write()
		s.read()
	}
	tok.Type = op.Ident
	tok.Literal = s.literal()
}

func (s *FormulaLexer) scanIdent(tok *Token) {
	reco := recognizeCell()
	for !s.done() && isAlpha(s.char) {
		reco.Update(s.char)
		s.write()
		s.read()
	}
	tok.Type = op.Ident
	tok.Literal = s.literal()
	if reco.IsCell() {
		tok.Type = op.Cell
	}
}

func (s *FormulaLexer) scanNumber(tok *Token) {
	tok.Type = op.Number
	for !s.done() && isDigit(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.literal()
	if s.char == dot {
		s.write()
		s.read()
		for !s.done() && isDigit(s.char) {
			s.write()
			s.read()
		}
		tok.Literal = s.literal()
	}
	if s.char == 'e' || s.char == 'E' {
		s.write()
		s.read()
		if isSign(s.char) {
			s.write()
			s.read()
		}
		for !s.done() && isDigit(s.char) {
			s.write()
			s.read()
		}
		tok.Literal = s.literal()
	}
}

func (s *FormulaLexer) scanLiteral(tok *Token) {
	quote := s.char
	s.read()
	for !s.done() && !isQuote(s.char) {
		s.write()
		s.read()
		if s.char == dquote && s.peek() == s.char {
			s.write()
			s.read()
			s.read()
		}
	}
	tok.Type = op.Literal
	tok.Literal = s.literal()
	if isQuote(s.char) && quote == s.char {
		s.read()
	} else {
		tok.Type = op.Invalid
	}
}

func (s *FormulaLexer) scanOperator(tok *Token) {
	tok.Type = op.Invalid
	switch s.char {
	case amper:
		tok.Type = op.Concat
	case percent:
		tok.Type = op.Percent
	case plus:
		tok.Type = op.Add
	case minus:
		tok.Type = op.Sub
	case star:
		tok.Type = op.Mul
	case slash:
		tok.Type = op.Div
	case caret:
		tok.Type = op.Pow
	case langle:
		tok.Type = op.Lt
		if k := s.peek(); k == equal {
			tok.Type = op.Le
		} else if k == rangle {
			tok.Type = op.Ne
		}
		if tok.Type != op.Lt {
			s.read()
		}
	case rangle:
		tok.Type = op.Gt
		if k := s.peek(); k == equal {
			s.read()
			tok.Type = op.Ge
		}
	case equal:
		tok.Type = op.Eq
	case colon:
		tok.Type = op.RangeRef
	default:
		tok.Type = op.Invalid
	}
	s.read()
}

func (s *FormulaLexer) scanDelimiter(tok *Token) {
	tok.Type = op.Invalid
	switch s.char {
	case comma:
		tok.Type = op.Comma
	case semi:
		tok.Type = op.Semi
	case lparen:
		tok.Type = op.BegGrp
	case rparen:
		tok.Type = op.EndGrp
	default:
	}
	s.read()
}

type ScannerState struct {
	pos      int
	next     int
	char     rune
	position Position
}

type reader struct {
	input []byte
	pos   int
	next  int
	char  rune

	Position

	buf bytes.Buffer
}

func prepare(r io.Reader) (*reader, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	rs := reader{
		input: input,
	}
	rs.Position.Line++
	rs.read()
	return &rs, nil
}

func (r *reader) Save() ScannerState {
	return ScannerState{
		pos:      r.pos,
		next:     r.next,
		char:     r.char,
		position: r.Position,
	}
}

func (r *reader) Restore(state ScannerState) {
	r.Position = state.position
	r.pos = state.pos
	r.next = state.next
	r.char = state.char
}

func (r *reader) literal() string {
	return r.buf.String()
}

func (r *reader) write() {
	r.buf.WriteRune(r.char)
}

func (r *reader) reset() {
	r.buf.Reset()
}

func (r *reader) read() {
	if r.pos >= len(r.input) {
		r.char = 0
		return
	}
	c, n := utf8.DecodeRune(r.input[r.next:])
	if c == utf8.RuneError {
		r.char = 0
		r.next = len(r.input)
	}
	r.char, r.pos, r.next = c, r.next, r.next+n

	if r.char == nl {
		r.Line += 1
		r.Column = 0
	}
	r.Column++
}

func (r *reader) readUntil(char rune, eat bool) string {
	defer r.reset()
	for !r.done() && r.char != char {
		r.write()
		r.read()
	}
	if r.char == char && eat {
		r.read()
	}
	return strings.ToLower(r.literal())
}

func (r *reader) is(char rune) bool {
	return r.char == char
}

func (r *reader) peek() rune {
	c, _ := utf8.DecodeRune(r.input[r.next:])
	return c
}

func (r *reader) done() bool {
	return r.pos >= len(r.input) || r.char == 0
}

func (r *reader) SkipNL() {
	for isNL(r.char) {
		r.read()
	}
}

func (r *reader) skipBlanks() {
	for isBlank(r.char) {
		r.read()
	}
}

const (
	underscore = '_'
	bang       = '!'
	question   = '?'
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
	pipe       = '|'
	percent    = '%'
	dollar     = '$'
	nl         = '\n'
	cr         = '\r'
	pound      = '#'
	lsquare    = '['
	rsquare    = ']'
	backslash  = '\\'
	arobase    = '@'
)

func isComment(c rune) bool {
	return c == pound
}

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

func isAbsolute(c rune) bool {
	return c == dollar
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

func isNL(c rune) bool {
	return c == nl || c == cr
}

func isDelimiter(c rune) bool {
	return c == semi || c == lparen || c == rparen ||
		c == comma || c == lsquare || c == rsquare
}

func isOperator(c rune) bool {
	return c == plus || c == minus || c == slash || c == star ||
		c == langle || c == rangle || c == colon || c == bang ||
		c == equal || c == caret || c == amper || c == percent ||
		c == dot || c == pipe || c == arobase
}

func isSign(c rune) bool {
	return c == plus || c == minus
}
