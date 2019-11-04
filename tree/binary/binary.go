package binary

import (
	"fmt"
	"io"
)

// Node 节点
type Node struct {
	Left  *Node
	Right *Node
	Data  interface{}
}

// HeightOfTree 树高度
func HeightOfTree(node *Node) int64 {
	if node == nil {
		return 0
	}
	leftHeight := HeightOfTree(node.Left)
	rightHeight := HeightOfTree(node.Right)
	height := leftHeight
	if rightHeight > height {
		height = rightHeight
	}
	return height + 1
}

// IsFullBinaryTree 判断是否是满二叉树
func IsFullBinaryTree(node *Node) bool {
	if node == nil {
		return true
	}
	if node.Left == nil && node.Right == nil {
		return true
	}
	if node.Left != nil && node.Right != nil {
		return IsFullBinaryTree(node.Left) && IsFullBinaryTree(node.Right)
	}
	return false
}

func display(node *Node, w io.Writer) {
	PreOrderVisit(node, func(node *Node) {
		fmt.Fprintln(w, node.Data)
	})
}

// PostOrderVisit 后序遍历
func PostOrderVisit(node *Node, fn func(*Node)) {
	if node == nil {
		return
	}
	if node.Left != nil {
		PostOrderVisit(node.Left, fn)
	}
	if node.Right != nil {
		PostOrderVisit(node.Right, fn)
	}
	fn(node)
}

// PreOrderVisit 先序遍历
func PreOrderVisit(node *Node, fn func(*Node)) {
	if node == nil {
		return
	}
	fn(node)
	if node.Left != nil {
		PreOrderVisit(node.Left, fn)
	}
	if node.Right != nil {
		PreOrderVisit(node.Right, fn)
	}
}

// MidOrderVisit 中序遍历
func MidOrderVisit(node *Node, fn func(*Node)) {
	if node == nil {
		return
	}
	if node.Left != nil {
		MidOrderVisit(node.Left, fn)
	}
	fn(node)
	if node.Right != nil {
		MidOrderVisit(node.Right, fn)
	}
}
