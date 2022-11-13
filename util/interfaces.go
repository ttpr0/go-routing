package util

type IIterator[T any] interface {
	Next() (T, bool)
}
