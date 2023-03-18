package util

import (
	"testing"
	"time"
)

func TestBlock(t *testing.T) {
	ch := make(chan int, 10)
	block := NewBlock()

	go func(ch chan int) {
		ch <- 0
		block.Take()
		ch <- 1
	}(ch)

	time.Sleep(1000)
	ch <- 10
	if !block.is_locked {
		t.Errorf("expected true, got false")
	}
	block.Release()
	time.Sleep(1000)
	ch <- 11

	val := <-ch
	if val != 0 {
		t.Errorf("expected 0, got %v", val)
	}
	val = <-ch
	if val != 10 {
		t.Errorf("expected 10, got %v", val)
	}
	val = <-ch
	if val != 1 {
		t.Errorf("expected 1, got %v", val)
	}
	val = <-ch
	if val != 11 {
		t.Errorf("expected 11, got %v", val)
	}
}
