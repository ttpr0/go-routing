package util

import "sync"

type Promise[T any] <-chan T

func Async[T any](fun func() T) Promise[T] {
	val := make(chan T)
	go func(fun func() T, val chan T) {
		val <- fun()
	}(fun, val)
	return val
}

func Await[T any](promise Promise[T]) T {
	return <-promise
}

type _AsyncIterator[T any] struct {
	agg chan T
}

func (self _AsyncIterator[T]) Next() (T, bool) {
	val, ok := <-self.agg
	return val, ok
}

func AwaitList[T any](promises List[Promise[T]]) IIterator[T] {
	agg := make(chan T)
	go func() {
		wg := sync.WaitGroup{}
		for _, p := range promises {
			wg.Add(1)
			go func(p Promise[T]) {
				agg <- Await(p)
				wg.Done()
			}(p)
		}
		wg.Wait()
		close(agg)
	}()
	return _AsyncIterator[T]{agg}
}
