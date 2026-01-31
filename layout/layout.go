package layout

type Dimension struct {
	Lines   int64
	Columns int64
}

func (d Dimension) Max(other Dimension) Dimension {
	if other.Lines > d.Lines {
		d.Lines = other.Lines
	}
	if other.Columns > d.Columns {
		d.Columns = other.Columns
	}
	return d
}
