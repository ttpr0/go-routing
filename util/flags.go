package util

type _Flag[T any] struct {
	data     T
	_counter int32
}

type Flags[T any] struct {
	flags    Array[_Flag[T]]
	_default T
	_counter *int32
}

func NewFlags[T any](count int32, default_data T) Flags[T] {
	flags := NewArray[_Flag[T]](int(count))
	for i := 0; i < flags.Length(); i++ {
		flags[i] = _Flag[T]{
			data:     default_data,
			_counter: 0,
		}
	}
	counter := int32(0)
	return Flags[T]{
		flags:    flags,
		_default: default_data,
		_counter: &counter,
	}
}

func (self *Flags[T]) Get(id int32) *T {
	flag := &(self.flags[id])
	if flag._counter != *self._counter {
		*flag = _Flag[T]{
			data:     self._default,
			_counter: *self._counter,
		}
	}
	return &flag.data
}
func (self *Flags[T]) Reset() {
	*self._counter += 1
	if *self._counter > 4000000 {
		for i := 0; i < self.flags.Length(); i++ {
			self.flags[i] = _Flag[T]{
				data:     self._default,
				_counter: 0,
			}
		}
		*self._counter = 0
	}
}
