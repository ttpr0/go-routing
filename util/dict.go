package util

type Dict[K comparable, V comparable] map[K]V

func (self *Dict[K, V]) Length() int {
	return len(*self)
}
func (self *Dict[K, V]) Clear() {
	for k := range *self {
		delete((*self), k)
	}
}
func (self *Dict[K, V]) Get(key K) V {
	return (*self)[key]
}
func (self *Dict[K, V]) Set(key K, value V) {
	(*self)[key] = value
}
func (self *Dict[K, V]) Delete(key K) {
	delete(*self, key)
}
func (self *Dict[K, V]) FindKey(value V) (K, bool) {
	for k, v := range *self {
		if v == value {
			return k, true
		}
	}
	var t K
	return t, false
}
func (self *Dict[K, V]) ContainsKey(key K) bool {
	_, ok := (*self)[key]
	return ok
}

func NewDict[K comparable, V comparable](cap int) Dict[K, V] {
	return make(map[K]V, cap)
}
