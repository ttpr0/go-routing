package isochrone

import (
	"fmt"

	"github.com/ttpr0/go-routing/geo"
	"golang.org/x/exp/slog"
)

//**********************************************************
// marching squares
//**********************************************************

// Returns the type of square at the given location (x,y is lower left corner of the square)
func _GetMarchingSquareIndex(tree *IsoTree[int], x, y int32, max_range int) int {
	range0, active0 := tree.Get(x, y)
	range1, active1 := tree.Get(x+1, y)
	range2, active2 := tree.Get(x+1, y+1)
	range3, active3 := tree.Get(x, y+1)
	key := 0
	if active0 && range0 <= max_range {
		key += 1
	}
	if active1 && range1 <= max_range {
		key += 2
	}
	if active2 && range2 <= max_range {
		key += 4
	}
	if active3 && range3 <= max_range {
		key += 8
	}
	return key
}

// This function processes a single cube
//
// Takes a square index a direction (i.e. the edge of the square clockwise from the bottom) and the location of the square
//
// Returns the next square location and direction as well as a point to add to the polygon and the square index to replace this square with (-1 if remove requested)
func _ProcessSquare(index int, direction int, x, y int32, rasterizer IRasterizer) (int32, int32, int, geo.Coord, int) {
	new_direction := -1
	replace_index := -1
	if direction == -1 {
		direction = _GetDefaultDirection(index)
	}
	switch index {
	case 0:
		new_direction, replace_index = _ProcessSquare0(direction)
	case 1:
		new_direction, replace_index = _ProcessSquare1(direction)
	case 2:
		new_direction, replace_index = _ProcessSquare2(direction)
	case 3:
		new_direction, replace_index = _ProcessSquare3(direction)
	case 4:
		new_direction, replace_index = _ProcessSquare4(direction)
	case 5:
		new_direction, replace_index = _ProcessSquare5(direction)
	case 6:
		new_direction, replace_index = _ProcessSquare6(direction)
	case 7:
		new_direction, replace_index = _ProcessSquare7(direction)
	case 8:
		new_direction, replace_index = _ProcessSquare8(direction)
	case 9:
		new_direction, replace_index = _ProcessSquare9(direction)
	case 10:
		new_direction, replace_index = _ProcessSquare10(direction)
	case 11:
		new_direction, replace_index = _ProcessSquare11(direction)
	case 12:
		new_direction, replace_index = _ProcessSquare12(direction)
	case 13:
		new_direction, replace_index = _ProcessSquare13(direction)
	case 14:
		new_direction, replace_index = _ProcessSquare14(direction)
	case 15:
		new_direction, replace_index = _ProcessSquare15(direction)
	}
	loc := rasterizer.IndexToPoint(x, y)
	switch direction {
	case 0:
		loc[0] += 0.5 * rasterizer.GetCellSize()
	case 1:
		loc[1] += 0.5 * rasterizer.GetCellSize()
	case 2:
		loc[0] += 0.5 * rasterizer.GetCellSize()
		loc[1] += rasterizer.GetCellSize()
	case 3:
		loc[0] += rasterizer.GetCellSize()
		loc[1] += 0.5 * rasterizer.GetCellSize()
	}
	loc[0] += 0.5 * rasterizer.GetCellSize()
	loc[1] += 0.5 * rasterizer.GetCellSize()
	newx := x
	newy := y
	switch new_direction {
	case 0:
		newy += 1
	case 1:
		newx += 1
	case 2:
		newy -= 1
	case 3:
		newx -= 1
	}

	return newx, newy, new_direction, loc, replace_index
}

// Returns the default direction for the given square index (leads to counter-clockwise direction for outer polygons)
func _GetDefaultDirection(index int) int {
	switch index {
	case 0:
		return -1
	case 1:
		return 0
	case 2:
		return 3
	case 3:
		return 3
	case 4:
		return 2
	case 5:
		return 2
	case 6:
		return 2
	case 7:
		return 2
	case 8:
		return 1
	case 9:
		return 0
	case 10:
		return 1
	case 11:
		return 3
	case 12:
		return 1
	case 13:
		return 0
	case 14:
		return 1
	case 15:
		return -1
	}
	return -1
}

//**********************************************************
// square functions
//**********************************************************

// each function takes the in direction (i.e. the edge of the square)
// and returns the out direction and the square index to replace this
// square with (-1 if remove requested)

func _ProcessSquare0(direction int) (int, int) {
	panic("this square should never be traversed")
}

func _ProcessSquare1(direction int) (int, int) {
	if direction != 0 && direction != 1 {
		slog.Debug(fmt.Sprintf("invalid direction: %v", direction))
		panic("invalid direction")
	}
	if direction == 0 {
		return 3, -1
	} else {
		return 2, -1
	}
}

func _ProcessSquare2(direction int) (int, int) {
	if direction != 0 && direction != 3 {
		slog.Debug(fmt.Sprintf("invalid direction: %v", direction))
		panic("invalid direction")
	}
	if direction == 0 {
		return 1, -1
	} else {
		return 2, -1
	}
}

func _ProcessSquare3(direction int) (int, int) {
	if direction != 1 && direction != 3 {
		slog.Debug(fmt.Sprintf("invalid direction: %v", direction))
		panic("invalid direction")
	}
	if direction == 1 {
		return 1, -1
	} else {
		return 3, -1
	}
}

func _ProcessSquare4(direction int) (int, int) {
	if direction != 2 && direction != 3 {
		slog.Debug(fmt.Sprintf("invalid direction: %v", direction))
		panic("invalid direction")
	}
	if direction == 2 {
		return 1, -1
	} else {
		return 0, -1
	}
}

func _ProcessSquare5(direction int) (int, int) {
	if direction < 0 || direction > 3 {
		slog.Debug(fmt.Sprintf("invalid direction: %v", direction))
		panic("invalid direction")
	}
	if direction == 0 {
		return 1, 7
	} else if direction == 3 {
		return 2, 7
	} else if direction == 1 {
		return 0, 13
	} else {
		return 3, 13
	}
}

func _ProcessSquare6(direction int) (int, int) {
	if direction != 0 && direction != 2 {
		slog.Debug(fmt.Sprintf("invalid direction: %v", direction))
		panic("invalid direction")
	}
	if direction == 0 {
		return 0, -1
	} else {
		return 2, -1
	}
}

func _ProcessSquare7(direction int) (int, int) {
	if direction != 1 && direction != 2 {
		slog.Debug(fmt.Sprintf("invalid direction: %v", direction))
		panic("invalid direction")
	}
	if direction == 1 {
		return 0, -1
	} else {
		return 3, -1
	}
}

func _ProcessSquare8(direction int) (int, int) {
	return _ProcessSquare7(direction)
}

func _ProcessSquare9(direction int) (int, int) {
	return _ProcessSquare6(direction)
}

func _ProcessSquare10(direction int) (int, int) {
	if direction < 0 || direction > 3 {
		slog.Debug(fmt.Sprintf("invalid direction: %v", direction))
		panic("invalid direction")
	}
	if direction == 0 {
		return 3, 11
	} else if direction == 1 {
		return 2, 11
	} else if direction == 3 {
		return 0, 14
	} else {
		return 1, 14
	}
}

func _ProcessSquare11(direction int) (int, int) {
	return _ProcessSquare4(direction)
}

func _ProcessSquare12(direction int) (int, int) {
	return _ProcessSquare3(direction)
}

func _ProcessSquare13(direction int) (int, int) {
	return _ProcessSquare2(direction)
}

func _ProcessSquare14(direction int) (int, int) {
	return _ProcessSquare1(direction)
}

func _ProcessSquare15(direction int) (int, int) {
	return _ProcessSquare0(direction)
}
