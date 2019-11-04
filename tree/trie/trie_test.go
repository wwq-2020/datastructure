package trie

import "testing"

func TestTrie(t *testing.T) {

	tests := []struct {
		given    string
		expected string
	}{
		{given: "test1", expected: "hello1"},
		{given: "test2", expected: "hello2"},
		{given: "test3", expected: "hello3"},
	}
	trie := New()

	for _, test := range tests {
		trie.Put(test.given, test.expected)
		got, exist := trie.Get(test.given)
		if !exist {
			t.Fatalf("given:%+v,exist expected:true,got:false", test.given)
		}
		if got != test.expected {
			t.Fatalf("data expected:%+v,got:%+v", test.expected, got)
		}
		trie.Del(test.given)
		got, exist = trie.Get(test.given)
		if exist {
			t.Fatalf("given:%+v,exist expected:false,got:true", test.given)
		}
	}
}
