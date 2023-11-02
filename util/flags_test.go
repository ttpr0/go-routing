package util

import (
	"testing"
)

func Test_FlagsGet(t *testing.T) {
	flags := NewFlags[int](100, 100000)

	for i := 0; i < 100; i++ {
		val := flags.Get(int32(i))
		*val = i
	}

	for i := 0; i < 100; i++ {
		val := flags.Get(int32(i))
		if *val != i {
			t.Errorf("flag value should be %v, but got %v", i, *val)
		}
	}
}

func Test_FlagsReset(t *testing.T) {
	flags := NewFlags[int](100, 100000)

	for i := 0; i < 100; i++ {
		val := flags.Get(int32(i))
		*val = i
	}

	flags.Reset()

	for i := 0; i < 100; i++ {
		val := flags.Get(int32(i))
		if *val != 100000 {
			t.Errorf("flag value should be 100000, but got %v", *val)
		}
	}
}
