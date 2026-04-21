package parse

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
