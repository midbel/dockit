package slx

func One[T any](v T) []T {
	return []T{v}
}
