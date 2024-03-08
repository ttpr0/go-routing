package comps

//*******************************************
// graph io
//*******************************************

type IStoreable interface {
	_Store(path string)
}

func Store(comp IStoreable, path string) {
	comp._Store(path)
}

type IRemoveable interface {
	_Remove(path string)
}

func Remove[T IRemoveable](path string) {
	var comp T
	comp._Remove(path)
}
