package slx

import (
	"slices"
)

func One[T any](v T) []T {
	return []T{v}
}

func Make[T any](v ...T) []T {
	return slices.Clone(v)
}
