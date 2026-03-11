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
	registry map[string]int
}

func NewNameIndex() *NameIndex {
	return &NameIndex{
		registry: make(map[string]int),
	}
}

func (n *NameIndex) Available(name string) bool {
	_, ok := n.registry[name]
	return ok
}

func (n *NameIndex) Count(name string) int {
	return n.registry[name]
}

func (n *NameIndex) Next(name string) string {
	if !n.Available(name) {
		n.reset(name)
		return name
	}
	ix := n.update(name)
	return fmt.Sprintf("%s_%03d", name, ix)
}

func (n *NameIndex) Delete(name string) {
	ix := strings.LastIndexByte(name, '_')
	if ix < 0 {
		return
	}
	name = name[:ix]
	if c, ok := n.registry[name]; ok {
		c--
		n.registry[name] = c
		if c == 0 {
			delete(n.registry, name)
		}
	}

}

func (n *NameIndex) update(name string) int {
	n.registry[name] += 1
	return n.registry[name]
}

func (n *NameIndex) reset(name string) {
	n.registry[name] = 0
}
