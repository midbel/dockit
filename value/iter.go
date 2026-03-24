package value

import (
	"iter"
)

type ValueIterator interface {
	Values() iter.Seq[ScalarValue]
}

func Each(args []Value, fn func(Value)) Value {
	for _, a := range args {
		if IsError(a) {
			return a
		}
		if IsScalar(a) {
			fn(a)
		} else if IsArray(a) {
			it, ok := a.(ValueIterator)
			if !ok {
				continue
			}
			var dat []Value
			for v := range it.Values() {
				dat = append(dat, v)
			}
			Each(dat, fn)
		}
	}
	return nil
}

func Collect[T any](args []Value, do func(v Value) (T, bool)) []T {
	all := make([]T, 0, len(args))
	Each(args, func(v Value) {
		r, ok := do(v)
		if ok {
			all = append(all, r)
		}
	})
	return all
}

func Map[T any](args []Value, do func(v Value) T) []T {
	all := make([]T, 0, len(args))
	Each(args, func(v Value) {
		all = append(all, do(v))
	})
	return all
}

func Reduce[T any](args []Value, init T, do func(acc T, v Value) T) T {
	acc := init
	Each(args, func(v Value) {
		acc = do(acc, v)
	})
	return acc
}

func Filter(args []Value, keep func(v Value) bool) []Value {
	all := make([]Value, 0, len(args))
	Each(args, func(v Value) {
		if keep(v) {
			all = append(all, v)
		}
	})
	return all
}
