package layout

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Selection interface {
	Indices(*Range) []int64
}

func SelectionFromString(str string) (Selection, error) {
	var (
		list  []Selection
		parts = strings.Split(str, ";")
	)
	for _, str := range parts {
		parts := strings.Split(strings.TrimSpace(str), ":")
		switch n := len(parts); n {
		case 1:
			ix, _ := ParseIndex(parts[0])
			list = append(list, columnRef{
				Index: ix,
			})
		case 2, 3:
			lo, _ := ParseIndex(parts[0])
			hi, _ := ParseIndex(parts[1])
			ref := columnSpan{
				Starts: lo,
				Ends:   hi,
				Step:   1,
			}
			if n == 3 && parts[2] != "" {
				st, err := strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					return nil, err
				}
				ref.Step = st
			}
			list = append(list, ref)
		default:
			return nil, fmt.Errorf("selection: invalid syntax")
		}
	}
	if len(list) == 1 {
		return list[0], nil
	}
	combined := combinedRef{
		list: list,
	}
	return combined, nil
}

type columnRef struct {
	Index int64
}

func SelectSingle(ix int64) Selection {
	return columnRef{
		Index: ix,
	}
}

func (c columnRef) Indices(rg *Range) []int64 {
	if rg == nil {
		return nil
	}
	if c.Index >= rg.Starts.Column && c.Index <= rg.Ends.Column {
		return []int64{c.Index - 1}
	}
	return nil
}

type columnSpan struct {
	Starts int64
	Ends   int64
	Step   int64
}

func SelectSpan(from, to, step int64) Selection {
	if step == 0 {
		step++
	}
	return columnSpan{
		Starts: from,
		Ends:   to,
		Step:   step,
	}
}

func (c columnSpan) Indices(rg *Range) []int64 {
	if rg == nil {
		return nil
	}
	var (
		all     []int64
		step    = c.Step
		starts  = c.Starts
		ends    = c.Ends
		forward bool
	)
	if step == 0 {
		step = 1
	}
	forward = step > 0

	if starts == 0 {
		if forward {
			starts = rg.Starts.Column
		} else {
			starts = rg.Ends.Column
		}
	}
	if ends == 0 {
		if forward {
			ends = rg.Ends.Column
		} else {
			ends = rg.Starts.Column
		}
	}

	if forward {
		starts = max(starts, rg.Starts.Column)
		ends = min(ends, rg.Ends.Column)

		for i := starts; i <= ends; i += step {
			all = append(all, i-1)
		}
	} else {
		starts = min(starts, rg.Ends.Column)
		ends = max(ends, rg.Starts.Column)

		for i := starts; i >= ends; i += step {
			all = append(all, i-1)
		}
	}
	return all
}

type combinedRef struct {
	list []Selection
}

func (r combinedRef) Indices(rg *Range) []int64 {
	var all []int64
	for i := range r.list {
		all = slices.Concat(all, r.list[i].Indices(rg))
	}
	return all
}
