package ds

import (
	"slices"
)

type Stack[T any] struct {
	values []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		values: make([]T, 0, 16),
	}
}

func (s *Stack[T]) Clone() *Stack[T] {
	return &Stack[T]{
		values: slices.Clone(s.values),
	}
}

func (s *Stack[T]) Len() int {
	return len(s.values)
}

func (s *Stack[T]) Push(v T) {
	s.values = append(s.values, v)
}

func (s *Stack[T]) Peek() (T, bool) {
	n := len(s.values)
	if n == 0 {
		var z T
		return z, false
	}
	return s.values[n-1], true
}

func (s *Stack[T]) Pop() (T, bool) {
	n := len(s.values)
	if n == 0 {
		var z T
		return z, false
	}
	z := s.values[n-1]
	s.values = s.values[:n-1]
	return z, true
}
