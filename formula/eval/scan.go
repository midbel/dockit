package eval

import (
	"bytes"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/midbel/dockit/formula/op"
)

const (
	kwImport  = "import"
	kwUse     = "use"
	kwUsing   = "using"
	kwWith    = "with"
	kwPrint   = "print"
	kwSave    = "save"
	kwExport  = "export"
	kwDefault = "default"
	kwFrom    = "from"
	kwIn      = "in"
	kwAs      = "as"
	kwTo      = "to"
	kwEnd     = "end"
	kwRo      = "ro"
	kwRw      = "rw"
	kwLock    = "lock"
	kwUnlock  = "unlock"
	kwPush    = "push"
	kwPop     = "pop"
	kwClear   = "clear"
)

func isKeyword(str string) bool {
	switch str {
	case kwImport:
	case kwUse:
	case kwUsing:
	case kwWith:
	case kwFrom:
	case kwPrint:
	case kwSave:
	case kwExport:
	case kwDefault:
	case kwIn:
	case kwAs:
	case kwTo:
	case kwEnd:
	case kwRo:
	case kwRw:
	case kwLock:
	case kwUnlock:
	case kwPush:
	case kwPop:
	case kwClear:
	default:
		return false
	}
	return true
}

type ScanMode int8

const (
	ModeFormula ScanMode = 1 << iota
	ModeScript
)

type ScannerState struct {
	pos      int
	next     int
	char     rune
	position Position
}

type Scanner struct {
	input []byte
	pos   int
	next  int
	char  rune

	Position

	buf  bytes.Buffer
	mode ScanMode
}

func Scan(r io.Reader, mode ScanMode) (*Scanner, error) {
	var (
		scan Scanner
		err  error
	)
	scan.mode = mode
	scan.input, err = io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	scan.Position.Line = 1
	scan.read()
	if scan.char == equal {
		scan.read()
	}
	return &scan, nil
}

func (s *Scanner) Save() ScannerState {
	return ScannerState{
		pos:      s.pos,
		next:     s.next,
		char:     s.char,
		position: s.Position,
	}
}

func (s *Scanner) Restore(state ScannerState) {
	s.Position = state.position
	s.pos = state.pos
	s.next = state.next
	s.char = state.char
}

func (s *Scanner) Peek() Token {
	currState := s.Save()
	defer s.Restore(currState)
	return s.Scan()
}

func (s *Scanner) Value() any {
	tok := s.Scan()
	switch tok.Type {
	case op.Number:
		f, _ := strconv.ParseFloat(tok.Literal, 64)
		return f
	case op.Ident:
		if tok.Literal == "true" {
			return true
		}
		return false
	case op.Literal:
		return tok.Literal
	default:
		return nil
	}
}

func (s *Scanner) Scan() Token {
	s.skipBlanks()

	var tok Token
	tok.Position = s.Position
	if s.done() {
		tok.Type = op.EOF
		return tok
	}
	defer s.reset()
	switch {
	case isNL(s.char) && s.mode == ModeScript:
		s.scanNL(&tok)
	case isComment(s.char) && s.mode == ModeScript:
		s.scanComment(&tok)
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

func (s *Scanner) scanNL(tok *Token) {
	s.skipNL()
	tok.Type = op.Eol
}

func (s *Scanner) scanComment(tok *Token) {
	s.read()
	if s.char == bang {
		s.scanDirective(tok)
		return
	}
	s.skipBlanks()
	for !s.done() && !isNL(s.char) {
		s.write()
		s.read()
	}
	s.skipNL()
	tok.Type = op.Comment
	tok.Literal = s.literal()
}

func (s *Scanner) scanDirective(tok *Token) {
	s.read()
	if s.char == space {
		s.skipBlanks()
		tok.Type = op.Pragma
		return
	}
	for !s.done() && !isNL(s.char) {
		s.write()
		s.read()
	}
	tok.Type = op.Directive
	tok.Literal = s.literal()
}

func (s *Scanner) scanIdent(tok *Token) {
	reco := recognizeCell()
	for !s.done() && isAlpha(s.char) {
		reco.Update(s.char)
		s.write()
		s.read()
	}
	tok.Type = op.Ident
	tok.Literal = s.literal()
	if s.allowKeywords() && isKeyword(tok.Literal) {
		tok.Type = op.Keyword
	}
	if reco.IsCell() {
		tok.Type = op.Cell
	}
}

func (s *Scanner) scanNumber(tok *Token) {
	tok.Type = op.Number
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
	tok.Type = op.Literal
	tok.Literal = s.literal()
	if isQuote(s.char) && quote == s.char {
		s.read()
	} else {
		tok.Type = op.Invalid
	}
}

func (s *Scanner) scanOperator(tok *Token) {
	tok.Type = op.Invalid
	switch s.char {
	case dot:
		tok.Type = op.Dot
	case pipe:
		tok.Type = op.Union
	case amper:
		tok.Type = op.Concat
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = op.ConcatAssign
		}
	case percent:
		tok.Type = op.Percent
	case plus:
		tok.Type = op.Add
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = op.AddAssign
		}
	case minus:
		tok.Type = op.Sub
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = op.SubAssign
		}
	case star:
		tok.Type = op.Mul
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = op.MulAssign
		}
	case slash:
		tok.Type = op.Div
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = op.DivAssign
		}
	case caret:
		tok.Type = op.Pow
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = op.PowAssign
		}
	case langle:
		tok.Type = op.Lt
		s.read()
		if s.char == equal {
			tok.Type = op.Le
		} else if s.char == rangle {
			tok.Type = op.Ne
		}
	case rangle:
		tok.Type = op.Gt
		s.read()
		if s.char == equal {
			tok.Type = op.Ge
		}
	case equal:
		tok.Type = op.Eq
	case colon:
		tok.Type = op.RangeRef
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = op.Assign
		}
	case bang:
		tok.Type = op.SheetRef
	default:
	}
	s.read()
}

func (s *Scanner) scanDelimiter(tok *Token) {
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
	case lsquare:
		tok.Type = op.BegProp
	case rsquare:
		tok.Type = op.EndProp
	default:
	}
	s.read()
}

func (s *Scanner) allowKeywords() bool {
	return s.mode == ModeScript
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

	if s.char == nl {
		s.Line += 1
		s.Column = 0
	}
	s.Column++
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.input[s.next:])
	return r
}

func (s *Scanner) done() bool {
	return s.pos >= len(s.input) || s.char == 0
}

func (s *Scanner) skipNL() {
	for isNL(s.char) {
		s.read()
	}
}

func (s *Scanner) skipBlanks() {
	for isBlank(s.char) {
		s.read()
	}
}

type recoMode int

const (
	cellCol recoMode = iota // reading column (A-Z)
	cellRow                 // reading row (1-9 then 0-9)
	cellAbsCol
	cellAbsRow
	cellDead // invalid
)

type cellRecognizer struct {
	state  recoMode
	hasRow bool
}

func recognizeCell() *cellRecognizer {
	return &cellRecognizer{
		state: cellAbsCol,
	}
}

func (c *cellRecognizer) Update(ch rune) {
	if c.state == cellDead {
		return
	}
	switch c.state {
	case cellAbsCol:
		if ch == dollar {
			break
		}
		if isUpper(ch) {
			c.toCol()
			break
		}
		c.toDead()
	case cellAbsRow:
		if isDigit(ch) && ch != '0' {
			c.toRow()
			break
		}
		c.toDead()
	case cellCol:
		if isUpper(ch) {
			break
		}
		if ch == dollar {
			c.toAbsRow()
			break
		}
		if isDigit(ch) && ch != '0' {
			c.toRow()
			break
		}
		c.toDead()
	case cellRow:
		if isDigit(ch) {
			break
		}
		c.toDead()
	}
}

func (c *cellRecognizer) IsCell() bool {
	return c.state == cellRow
}

func (c *cellRecognizer) toDead() {
	c.state = cellDead
}

func (c *cellRecognizer) toCol() {
	c.state = cellCol
}

func (c *cellRecognizer) toRow() {
	c.state = cellRow
}

func (c *cellRecognizer) toAbsRow() {
	c.state = cellAbsRow
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
	pipe       = '|'
	percent    = '%'
	dollar     = '$'
	nl         = '\n'
	cr         = '\r'
	pound      = '#'
	lsquare    = '['
	rsquare    = ']'
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
		c == dot || c == pipe
}
