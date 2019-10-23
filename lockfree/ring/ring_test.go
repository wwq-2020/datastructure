package ring

import "testing"

func TestRing(t *testing.T) {
	r := New(4)
	err := r.Push("1")
	if err != nil {
		t.Fatalf("expected err:nil,got:%+v", err)
	}
	err = r.Push("2")
	if err != nil {
		t.Fatalf("expected err:nil,got:%+v", err)
	}
	err = r.Push("3")
	if err != nil {
		t.Fatalf("expected err:nil,got:%+v", err)
	}
	err = r.Push("4")
	if err != nil {
		t.Fatalf("expected err:nil,got:%+v", err)
	}
	err = r.Push("5")
	if err == nil {
		t.Fatal("expected err not nil,got nil")
	}
	val, err := r.Pop()
	if err != nil {
		t.Fatalf("expected err:nil,got:%+v", err)
	}
	if val.(string) != "1" {
		t.Fatalf("expected val:1,got:%+v", val)
	}
	val, err = r.Pop()
	if err != nil {
		t.Fatalf("expected err:nil,got:%+v", err)
	}
	if val.(string) != "2" {
		t.Fatalf("expected val:2,got:%+v", val)
	}
	val, err = r.Pop()
	if err != nil {
		t.Fatalf("expected err:nil,got:%+v", err)
	}
	if val.(string) != "3" {
		t.Fatalf("expected val:4,got:%+v", val)
	}
	val, err = r.Pop()
	if err != nil {
		t.Fatalf("expected err:nil,got:%+v", err)
	}
	if val.(string) != "4" {
		t.Fatalf("expected val:4,got:%+v", val)
	}
	val, err = r.Pop()
	if err == nil {
		t.Fatal("expected err not nil,got nil")
	}

}
