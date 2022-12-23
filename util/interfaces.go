package util

type IIterator[T any] interface {
	Next() (T, bool)
}

type number interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}
