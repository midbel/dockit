package ds

type Node[T any] struct {
	name     string
	value    T
	setted   bool
	children map[string]*Node[T]
}

func createNode[T any](name string) *Node[T] {
	return &Node[T]{
		name:     name,
		children: make(map[string]*Node[T]),
	}
}

type Trie[T any] struct {
	root *Node[T]
}

func NewTrie[T any]() *Trie[T] {
	trie := Trie[T]{
		root: createNode[T](""),
	}
	return &trie
}

func (t *Trie[T]) Get(path []string) (T, bool) {
	var (
		node = t.root
		ok   bool
	)
	for _, name := range path {
		node, ok = node.children[name]
		if !ok {
			var z T
			return z, ok
		}
	}
	return node.value, node.setted
}

func (t *Trie[T]) Walk(path []string, fn func(path []string, v T)) {
	node := t.root
	// move to prefix
	for _, name := range path {
		n, ok := node.children[name]
		if !ok {
			return
		}
		node = n
	}

	var walk func(n *Node[T], path []string)

	walk = func(n *Node[T], path []string) {
		if n.setted {
			fn(path, n.value)
		}

		for name, child := range n.children {
			walk(child, append(path, name))
		}
	}

	walk(node, path)
}

func (t *Trie[T]) Register(path []string, value T) {
	node := t.root
	for _, name := range path {
		if node.children[name] == nil {
			node.children[name] = createNode[T](name)
		}
		node = node.children[name]
	}
	node.name = path[len(path)-1]
	node.value = value
	node.setted = true
}
