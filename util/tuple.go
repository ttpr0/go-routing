package util

type Tuple[V1, V2 any] struct {
	A V1
	B V2
}

type Triple[V1, V2, V3 any] struct {
	A V1
	B V2
	C V3
}

// Creates and returns a new Tuple
func MakeTuple[V1, V2 any](itemA V1, itemB V2) Tuple[V1, V2] {
	return Tuple[V1, V2]{itemA, itemB}
}

// Creates and returns a new Triple
func MakeTriple[V1, V2, V3 any](itemA V1, itemB V2, itemC V3) Triple[V1, V2, V3] {
	return Triple[V1, V2, V3]{itemA, itemB, itemC}
}
