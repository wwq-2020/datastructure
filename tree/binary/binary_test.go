package binary

import (
	"bytes"
	"testing"
)

func TestBinary(t *testing.T) {
	tree := &Node{Data: "A"}
	tree.Left = &Node{Data: "B"}
	tree.Right = &Node{Data: "C"}
	tree.Left.Left = &Node{Data: "D"}
	tree.Left.Right = &Node{Data: "E"}
	tree.Right.Left = &Node{Data: "F"}
	buffer := bytes.NewBuffer(nil)
	expected := `A
B
D
E
C
F
`
	display(tree, buffer)
	got := buffer.String()
	if got != expected {
		t.Fatalf("expected:%s,got:%s", expected, got)
	}

}
