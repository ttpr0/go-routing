package util

type Flags[T any] struct {
	flags       Array[T]
	flag_counts Array[int32]
	_default    T
	_counter    int32
}

func NewFlags[T any](count int32, default_data T) Flags[T] {
	flags := NewArray[T](int(count))
	flag_counts := NewArray[int32](int(count))
	for i := 0; i < flags.Length(); i++ {
		flags[i] = default_data
		flag_counts[i] = 0
	}
	return Flags[T]{
		flags:       flags,
		flag_counts: flag_counts,
		_default:    default_data,
		_counter:    0,
	}
}

func (self *Flags[T]) Get(id int32) *T {
	flag := &(self.flags[id])
	if self.flag_counts[id] != self._counter {
		*flag = self._default
		self.flag_counts[id] = self._counter
	}
	return flag
}
func (self *Flags[T]) Reset() {
	self._counter += 1
	if self._counter > 4000000 {
		for i := 0; i < self.flags.Length(); i++ {
			self.flags[i] = self._default
			self.flag_counts[i] = 0
		}
		self._counter = 0
	}
}
