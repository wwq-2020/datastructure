package queue

import (
	"sync"
	"testing"
)

func TestSimple(t *testing.T) {
	queue := New()
	queue.Push("hello")
	queue.Push("world")
	got, err := queue.Pop()
	if err != nil {
		t.Fatal("mismatch")
	}
	if got.(string) != "hello" {
		t.Fatal("mismatch", got)
	}
	got, err = queue.Pop()
	if err != nil {
		t.Fatal("mismatch")
	}
	if got.(string) != "world" {
		t.Fatal("mismatch")
	}
}

func TestBPop(t *testing.T) {
	queue := New()
	var wg sync.WaitGroup
	wg.Add(4)
	var results []string
	var m sync.Mutex
	go func() {
		defer wg.Done()
		val, _ := queue.BPop(0)
		m.Lock()
		results = append(results, val.(string))
		m.Unlock()
	}()
	go func() {
		defer wg.Done()
		val, _ := queue.BPop(0)
		m.Lock()
		results = append(results, val.(string))
		m.Unlock()
	}()

	go func() {
		defer wg.Done()
		queue.Push("hello")
	}()
	go func() {
		defer wg.Done()
		queue.Push("world")
	}()

	wg.Wait()
	if len(results) != 2 {
		t.Fatal("mismatch")
	}

	if (results[0] != "hello" && results[0] != "world") ||
		(results[1] != "hello" && results[1] != "world") {
		t.Fatal("mismatch")

	}
}
