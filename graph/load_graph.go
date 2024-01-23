package graph

//*******************************************
// graph io
//*******************************************

type ILoadable[T any] interface {
	_New() T
	_Load(path string)
}

func Load[T ILoadable[T]](path string) T {
	var comp T
	comp = comp._New()
	comp._Load(path)
	return comp
}
