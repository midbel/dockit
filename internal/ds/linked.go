package ds

import (
	"iter"
)

type Linked[T any] struct {
	prev *Linked[T]
	item T
}

func EmptyList[T any]() *Linked[T] {
	return nil
}

func (k *Linked[T]) All() iter.Seq[T] {
	return nil
}

func (k *Linked[T]) Each(fn func(T) bool) {
	for curr := k; curr != nil; curr = curr.prev {
		if !fn(curr.item) {
			return
		}
	}
}

func (k *Linked[T]) Item() T {
	return k.item
}

func (k *Linked[T]) RootItem() T {
	i := k.Root()
	return i.Item()
}

func (k *Linked[T]) Root() *Linked[T] {
	if k == nil {
		return nil
	}
	curr := k
	for {
		if curr.prev == nil {
			break
		}
		curr = curr.prev
	}
	return curr
}

func (k *Linked[T]) Push(item T) *Linked[T] {
	i := &Linked[T]{
		prev: k,
		item: item,
	}
	return i
}

func (k *Linked[T]) Parent() *Linked[T] {
	if k == nil {
		return k
	}
	return k.prev
}

func (k *Linked[T]) Len() int {
	var (
		count int
		curr  = k
	)
	for curr != nil {
		count++
		curr = curr.prev
	}
	return count
}

func (k *Linked[T]) Truncate(n int) *Linked[T] {
	size := k.Len()
	if n <= 0 || n >= size {
		return nil
	}
	curr := k
	for curr != nil && size > n {
		curr = curr.prev
		size--
	}
	return curr
}
