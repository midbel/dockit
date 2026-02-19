package format

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/midbel/dockit/value"
)

type numberFormatter struct {
	minInt int
	maxInt int
	minDec int
	maxDec int

	signAlways  bool
	hasGrouping bool
	hasDecimal  bool

	decimalSep  byte
	thousandSep byte
}

func ParseNumberFormatter(pattern string) (Formatter, error) {
	if pattern == "" || pattern == "." || pattern == "-" || pattern == "+" {
		return nil, fmt.Errorf("invalid pattern given")
	}
	var (
		nf     numberFormatter
		left   string
		right  string
		zeroes = true
	)
	nf.decimalSep = '.'
	nf.thousandSep = ','

	left, right, nf.hasDecimal = strings.Cut(pattern, ".")

	if left == "" || left == "-" || left == "+" {
		return nil, fmt.Errorf("invalid pattern given")
	}

	for i := 0; i < len(right); i++ {
		if zeroes && right[i] == '0' {
			nf.minDec++
			nf.maxDec++
		} else if right[i] == '#' {
			zeroes = false
			nf.maxDec++
		} else {
			return nil, fmt.Errorf("unexpected character in fractional part pattern")
		}
	}

	if left[0] == '+' {
		nf.signAlways = true
		left = left[1:]
	}

	zeroes = true
	for i := len(left) - 1; i >= 0; i-- {
		if left[i] == ',' {
			nf.hasGrouping = true
			continue
		}

		if zeroes && left[i] == '0' {
			nf.minInt++
			nf.maxInt++
		} else if left[i] == '#' {
			zeroes = false
			nf.maxInt++
		} else {
			return nil, fmt.Errorf("unexpected character in integral part pattern")
		}
	}

	return nf, nil
}

func (nf numberFormatter) Format(v value.Value) (string, error) {
	vf, ok := v.(value.Float)
	if !ok {
		return "", fmt.Errorf("value is not a number")
	}

	var (
		scale      = math.Pow10(nf.maxDec)
		rounded    = math.Round(float64(vf)*scale) / scale
		integral   []byte
		fractional []byte
		str        = strconv.FormatFloat(rounded, 'f', nf.maxDec, 64)
		signed     = math.Signbit(float64(vf))
	)
	left, right, _ := strings.Cut(str, ".")
	if nf.maxDec > 0 {
		fractional = make([]byte, nf.maxDec)
		for i := 0; i < nf.maxDec; i++ {
			fractional[i] = '0'
		}
		copy(fractional, right)
	}
	for len(fractional) > nf.minDec && fractional[len(fractional)-1] == '0' {
		fractional = fractional[:len(fractional)-1]
	}
	if signed {
		left = left[1:]
	}
	integral = []byte(left)
	if z := len(integral); z < nf.minInt {
		tmp := make([]byte, nf.minInt)
		for i := 0; i < nf.minInt; i++ {
			tmp[i] = '0'
		}
		copy(tmp[nf.minInt-z:], integral)
		integral = tmp
	}
	if nf.hasGrouping {
		slices.Reverse(integral)
		var tmp []byte
		for i := 0; i < len(integral); i += 3 {
			if i+3 >= len(integral) {
				tmp = append(tmp, integral[i:]...)
			} else {
				tmp = append(tmp, integral[i:i+3]...)
				tmp = append(tmp, nf.thousandSep)
			}
		}
		slices.Reverse(tmp)
		integral = tmp
	}
	if signed || nf.signAlways {
		tmp := []byte{'+'}
		if signed {
			tmp[0] = '-'
		}
		integral = append(tmp, integral...)
	}
	var all []byte
	if len(fractional) > 0 {
		all = append(integral, nf.decimalSep)
		all = append(all, fractional...)
	} else {
		all = integral
	}
	return string(all), nil
}
