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
	trie := Trie{
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
	for _, name := range prefix {
		n, ok := node.Children[name]
		if !ok {
			return
		}
		node = n
	}

	var walk func(n *Node[T], path []string)

	walk = func(n *Node[T], path []string) {
		if n.setted {
			fn(path, n.Value)
		}

		for name, child := range n.Children {
			walk(child, append(path, name))
		}
	}

	walk(node, prefix)
}

func (t *Trie[T]) Register(path []string, value T) {
	node := t.root
	for _, name := range path {
		if node.children[name] == nil {
			node.Children[name] = createNode(name)
		}
		node = node.children[name]
	}
	node.name = path[len(path)-1]
	node.value = value
	node.setted = true
}
