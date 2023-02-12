package util

import (
	"testing"
)

func TestDequeBack(t *testing.T) {
	deque := NewDeque[int]()

	for i := 0; i < 100; i++ {
		deque.PushBack(i)
	}

	v := 99
	for {
		if deque.Size() == 0 {
			_, ok := deque.PopBack()
			if ok {
				t.Errorf("expected false, got true")
			}
			break
		}
		dv, _ := deque.PopBack()
		if dv != v {
			t.Errorf("expected %d, got %d", v, dv)
		}
		v -= 1
	}
}

func TestDequeFront(t *testing.T) {
	deque := NewDeque[int]()

	for i := 0; i < 100; i++ {
		deque.PushFront(i)
	}

	v := 99
	for {
		if deque.Size() == 0 {
			_, ok := deque.PopFront()
			if ok {
				t.Errorf("expected false, got true")
			}
			break
		}
		dv, _ := deque.PopFront()
		if dv != v {
			t.Errorf("expected %d, got %d", v, dv)
		}
		v -= 1
	}
}

func TestDequeGetSet(t *testing.T) {
	deque := NewDeque[int]()

	for i := 0; i < 100; i++ {
		deque.PushBack(i)
	}

	if v, _ := deque.GetAt(10); v != 10 {
		t.Errorf("expected %d, got %d", 10, v)
	}
	if _, ok := deque.GetAt(31); !ok {
		t.Errorf("expected true, got false")
	}
	if _, ok := deque.GetAt(100); ok {
		t.Errorf("expected false, got true")
	}

	ok := deque.SetAt(11, 9999)
	if v, _ := deque.GetAt(11); v != 9999 {
		t.Errorf("expected %d, got %d", 9999, v)
	}
	if !ok {
		t.Errorf("expected true, got false")
	}
	deque.SetAt(0, -10)
	if v, _ := deque.GetAt(0); v != -10 {
		t.Errorf("expected %d, got %d", -10, v)
	}
	ok = deque.SetAt(-10, -10)
	if ok {
		t.Errorf("expected false, got true")
	}
}

func TestDequeAddRemove(t *testing.T) {
	deque := NewDeque[int]()
	for i := 0; i < 100; i++ {
		deque.PushBack(i)
	}

	ok := deque.AddAt(10, 9999)
	if !ok {
		t.Errorf("expected true, got false")
	}
	if v, _ := deque.GetAt(10); v != 9999 {
		t.Errorf("expected %d, got %d", 9999, v)
	}
	if deque.Size() != 101 {
		t.Errorf("expected 101, got %d", deque.Size())
	}
	ok = deque.AddAt(-10, 9999)
	if ok {
		t.Errorf("expected false, got true")
	}
	if deque.Size() != 101 {
		t.Errorf("expected 101, got %d", deque.Size())
	}

	ok = deque.RemoveAt(20)
	if !ok {
		t.Errorf("expected true, got false")
	}
	if v, _ := deque.GetAt(20); v != 20 {
		t.Errorf("expected %d, got %d", 20, v)
	}
	if deque.Size() != 100 {
		t.Errorf("expected 100, got %d", deque.Size())
	}
	ok = deque.RemoveAt(-10)
	if ok {
		t.Errorf("expected false, got true")
	}
	if deque.Size() != 100 {
		t.Errorf("expected 100, got %d", deque.Size())
	}
}
