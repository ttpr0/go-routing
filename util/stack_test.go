package util

import (
	"testing"
)

func TestStackSize(t *testing.T) {
	stack := NewStack[int]()

	for i := 0; i < 100; i++ {
		stack.Push(i)
	}
	if stack.Size() != 100 {
		t.Errorf("stack size should be 100, but got %d", stack.Size())
	}

	for i := 0; i < 10; i++ {
		stack.Pop()
	}
	if stack.Size() != 90 {
		t.Errorf("stack size should be 90, but got %d", stack.Size())
	}

	for i := 0; i < 90; i++ {
		stack.Pop()
	}
	_, ok := stack.Pop()
	if ok {
		t.Errorf("should return false on empty stack, but got %v", ok)
	}
}

func TestStack(t *testing.T) {
	stack := NewStack[int]()

	for i := 0; i < 100; i++ {
		stack.Push(i)
	}

	val := 99
	for {
		v, ok := stack.Pop()
		if !ok {
			break
		}
		if v != val {
			t.Errorf("stack should return %d, but got %d", val, v)
		}
		val -= 1
	}
}
