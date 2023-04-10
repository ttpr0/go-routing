package util

import (
	"testing"
)

func TestNDArrayCreate(t *testing.T) {
	arr := NewNDArray[int](2, 2, 4)

	shape := arr.shape
	sizes := arr.strides
	if shape[0] != 2 || shape[1] != 2 || shape[2] != 4 {
		t.Errorf("shape should be (2, 2, 4), but got %d", shape)
	}
	if sizes[0] != 8 || sizes[1] != 4 || sizes[2] != 1 {
		t.Errorf("sizes should be (8, 4, 1), but got %d", sizes)
	}
}

func TestNDArrayReshape(t *testing.T) {
	temp := NewNDArray[int](2, 2, 4)
	arr, ok := temp.Reshape(2, 2, 2, 2)
	if !ok {
		t.Errorf("reshape should return true, but got false")
	}

	shape := arr.shape
	sizes := arr.strides
	if shape[0] != 2 || shape[1] != 2 || shape[2] != 2 || shape[3] != 2 {
		t.Errorf("shape should be (2, 2, 2, 2), but got %d", shape)
	}
	if sizes[0] != 8 || sizes[1] != 4 || sizes[2] != 2 || sizes[3] != 1 {
		t.Errorf("sizes should be (8, 4, 2, 1), but got %d", sizes)
	}
}

func TestNDArrayGetSet(t *testing.T) {
	arr := NewNDArray[int](2, 2, 4)

	arr.Set(10, 1, 1, 1)
	arr.Set(11, 0, 0, 3)
	arr.Set(12, 1, 0, 0)

	data := arr.data
	if data[13] != 10 {
		t.Errorf("value should be 10, but got %d", data[12])
	}
	if data[3] != 11 {
		t.Errorf("value should be 11, but got %d", data[3])
	}
	if data[8] != 12 {
		t.Errorf("value should be 12, but got %d", data[8])
	}

	data[1] = 13
	data[15] = 14
	data[6] = 15

	val := arr.Get(0, 0, 1)
	if val != 13 {
		t.Errorf("value should be 13, but got %d", val)
	}
	val = arr.Get(1, 1, 3)
	if val != 14 {
		t.Errorf("value should be 14, but got %d", val)
	}
	val = arr.Get(0, 1, 2)
	if val != 15 {
		t.Errorf("value should be 15, but got %d", val)
	}
}
