package util

type Matrix[T any] struct {
	data []T
	rows int
	cols int
}

// Returns the number of rows of the Matrix.
func (self *Matrix[T]) Rows() int {
	return self.rows
}

// Returns the number of cols of the Matrix.
func (self *Matrix[T]) Cols() int {
	return self.cols
}

// Returns the element at index.
func (self *Matrix[T]) Get(row, col int) T {
	return self.data[row*self.cols+col]
}

// Sets the element at index.
func (self *Matrix[T]) Set(row, col int, value T) {
	self.data[row*self.cols+col] = value
}

// Creates and Returns a new Matrix with rows and cols.
func NewMatrix[T any](rows, cols int) Matrix[T] {
	return Matrix[T]{
		data: make([]T, rows*cols),
		rows: rows,
		cols: cols,
	}
}
