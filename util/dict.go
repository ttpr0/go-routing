package util

type Dict[K comparable, V any] map[K]V

// Returns the number of elements in the Dict
func (self *Dict[K, V]) Length() int {
	return len(*self)
}

// Clears the Dict.
func (self *Dict[K, V]) Clear() {
	for k := range *self {
		delete((*self), k)
	}
}

// Returns the value of the key in the Dict.
func (self *Dict[K, V]) Get(key K) V {
	return (*self)[key]
}

// Sets the value of the key in the Dict.
func (self *Dict[K, V]) Set(key K, value V) {
	(*self)[key] = value
}

// Deletes a key from the Dict.
func (self *Dict[K, V]) Delete(key K) {
	delete(*self, key)
}

// Checks if a key exists in the Dict.
func (self *Dict[K, V]) ContainsKey(key K) bool {
	_, ok := (*self)[key]
	return ok
}

// Creates and returns a new Dict with the initial capacity cap.
// Dict with the key K and value V.
func NewDict[K comparable, V any](cap int) Dict[K, V] {
	return make(map[K]V, cap)
}

// Returns the key for a given value and a bool indicating success.
// If the value is not found false is returned.
func GetKeyOf[K comparable, V comparable](dict Dict[K, V], value V) (K, bool) {
	for k, v := range dict {
		if v == value {
			return k, true
		}
	}
	var t K
	return t, false
}
