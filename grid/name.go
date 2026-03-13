package grid

import (
	"fmt"
	"strings"
	"unicode"
)

func CleanName(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			return r
		}
		return -1
	}, str)
}

type NameIndex struct {
	counter map[string]int
	used    map[string]struct{}
}

func NewNameIndex() *NameIndex {
	return &NameIndex{
		counter: make(map[string]int),
		used:    make(map[string]struct{}),
	}
}

func (n *NameIndex) Next(name string) string {
	base := n.cutSuffix(name)
	if _, ok := n.used[base]; !ok {
		n.used[base] = struct{}{}
		n.counter[base] = 0
		return base
	}

	for {
		n.counter[base]++
		name := fmt.Sprintf("%s_%03d", base, n.counter[base])

		if _, ok := n.used[name]; !ok {
			n.used[name] = struct{}{}
			return name
		}
	}
}

func (n *NameIndex) Delete(name string) {
	delete(n.used, name)
}

func (n *NameIndex) cutSuffix(name string) string {
	ix := strings.LastIndexByte(name, '_')
	if ix < 0 {
		return name
	}
	trim := strings.TrimRightFunc(name[ix+1:], unicode.IsDigit)
	if trim != "" {
		return name
	}
	return name[:ix]
}
