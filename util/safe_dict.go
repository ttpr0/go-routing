package util

import (
	"sync"
)

type SafeDict[K comparable, V any] struct {
	dict map[K]V
	lock *sync.Mutex
}

// Returns the number of elements in the Dict
func (self *SafeDict[K, V]) Length() int {
	return len(self.dict)
}

// Clears the Dict.
func (self *SafeDict[K, V]) Clear() {
	self.lock.Lock()
	defer self.lock.Unlock()
	for k := range self.dict {
		delete(self.dict, k)
	}
}

// Returns the value of the key in the Dict.
func (self *SafeDict[K, V]) Get(key K) V {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.dict[key]
}

// Sets the value of the key in the Dict.
func (self *SafeDict[K, V]) Set(key K, value V) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.dict[key] = value
}

// Deletes a key from the Dict.
func (self *SafeDict[K, V]) Delete(key K) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.dict, key)
}

// Checks if a key exists in the Dict.
func (self *SafeDict[K, V]) ContainsKey(key K) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	_, ok := self.dict[key]
	return ok
}

// Creates and returns a new Dict with the initial capacity cap.
// Dict with the key K and value V.
func NewSafeDict[K comparable, V any](cap int) SafeDict[K, V] {
	return SafeDict[K, V]{
		dict: make(map[K]V, cap),
		lock: &sync.Mutex{},
	}
}
