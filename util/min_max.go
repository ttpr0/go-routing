package util

// returns the minimum of the two numbers
func Min[T number](a, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}

// returns the maximum of the two numbers
func Max[T number](a, b T) T {
	if a > b {
		return a
	}
	return b
}
