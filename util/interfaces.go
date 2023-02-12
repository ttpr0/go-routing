package util

// Iterator interface for iterating over a collection of objects T
type IIterator[T any] interface {
	Next() (T, bool)
}

// Interface representing numeric values
type number interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}
