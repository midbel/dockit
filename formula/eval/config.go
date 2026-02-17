package eval

import "fmt"

var directiveTrie *Trie

func init() {
	directiveTrie = NewTrie()
	directiveTrie.Register([]string{
		"print",
		"format",
		"rows",
	}, configurePrintRows)
	directiveTrie.Register([]string{
		"print",
		"format",
		"cols",
	}, configurePrintCols)
	directiveTrie.Register([]string{
		"print",
		"mode",
		"debug",
	}, configurePrintDebug)

	directiveTrie.Register([]string{
		"print",
		"format",
		"date",
	}, configureFormatDate)
	directiveTrie.Register([]string{
		"print",
		"format",
		"boolean",
	}, configureFormatBoolean)
	directiveTrie.Register([]string{
		"print",
		"format",
		"number",
	}, configureFormatNumber)

	directiveTrie.Register([]string{
		"import",
		"directory",
	}, configureContextDir)
}

type ConfigFunc func(*EngineConfig, any) error

type TrieNode struct {
	Name     string
	cmd      ConfigFunc
	Children map[string]*TrieNode
}

func createNode(name string) *TrieNode {
	return &TrieNode{
		Children: make(map[string]*TrieNode),
	}
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	trie := Trie{
		root: createNode(""),
	}
	return &trie
}

func (t *Trie) Configure(path []string, value any, cfg *EngineConfig) error {
	node := t.root
	for _, name := range path {
		child, ok := node.Children[name]
		if !ok {
			return fmt.Errorf("%s: option not found", name)
		}
		node = child
	}
	if node != nil && node.cmd != nil {
		return node.cmd(cfg, value)
	}
	return nil
}

func (t *Trie) Register(path []string, cmd ConfigFunc) error {
	node := t.root
	for _, name := range path {
		if node.Children[name] == nil {
			node.Children[name] = createNode(name)
		}
		node = node.Children[name]
	}
	node.cmd = cmd
	return nil
}

func configureContextDir(cfg *EngineConfig, value any) error {
	return nil
}

func configureFormatDate(cfg *EngineConfig, value any) error {
	return nil
}

func configureFormatNumber(cfg *EngineConfig, value any) error {
	return nil
}

func configureFormatBoolean(cfg *EngineConfig, value any) error {
	return nil
}

func configurePrintRows(cfg *EngineConfig, value any) error {
	switch v := value.(type) {
	case int:
		cfg.Print.Rows = v
	case int64:
		cfg.Print.Rows = int(v)
	case float64:
		cfg.Print.Rows = int(v)
	default:
	}
	return nil
}

func configurePrintCols(cfg *EngineConfig, value any) error {
	switch v := value.(type) {
	case int:
		cfg.Print.Cols = v
	case int64:
		cfg.Print.Cols = int(v)
	case float64:
		cfg.Print.Cols = int(v)
	default:
	}
	return nil
}

func configurePrintDebug(cfg *EngineConfig, value any) error {
	if b, ok := value.(bool); ok {
		cfg.Print.Debug = b
	}
	return nil
}
