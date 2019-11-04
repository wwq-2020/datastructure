package trie

import "sync"

// Node 节点
type Node struct {
	word     rune
	data     interface{}
	valid    bool
	children map[rune]*Node
}

// Trie Trie
type Trie struct {
	sync.RWMutex
	root *Node
}

// New 初始化Trie
func New() *Trie {
	return &Trie{
		root: newNode(),
	}
}

func newNode() *Node {
	return &Node{
		children: make(map[rune]*Node),
	}
}

// Put 添加
func (t *Trie) Put(key string, data interface{}) {
	words := []rune(key)
	t.Lock()
	defer t.Unlock()
	node := t.root
	for _, word := range words {
		child, exist := node.children[word]
		if exist {
			node = child
		} else {
			node = node.addChild(word)
		}
	}
	node.data = data
	node.valid = true
}

// Get 获取
func (t *Trie) Get(key string) (interface{}, bool) {
	words := []rune(key)
	t.RLock()
	defer t.RUnlock()
	node := t.root
	for _, word := range words {
		child, exist := node.children[word]
		if !exist {
			return nil, false
		}
		node = child
	}
	if node.valid {
		return node.data, true
	}
	return nil, false
}

// Del 删除
func (t *Trie) Del(key string) {
	words := []rune(key)
	t.Lock()
	defer t.Unlock()
	node := t.root
	for _, word := range words {
		child, exist := node.children[word]
		if exist {
			node = child
		} else {
			node = newNode()
			node.word = word
		}
	}
	node.data = nil
	node.valid = false
}

func (n *Node) addChild(word rune) *Node {
	child := newNode()
	child.word = word
	n.children[word] = child
	return child
}
