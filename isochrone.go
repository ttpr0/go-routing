package main

import (
	"fmt"
	"math"

	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/isochrone"
	"github.com/ttpr0/go-routing/routing"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

//**********************************************************
// isochrone request
//**********************************************************

type IsochroneRequest struct {
	Locations [][]float32 `json:"locations"`
	Range     []int32     `json:"range"`
}

//**********************************************************
// isochrone handler
//**********************************************************

func HandleIsochroneRequest(req IsochroneRequest) Result {
	// get profile
	profile_, res := GetRequestProfile(MANAGER, "driving-car", "time")
	if !profile_.HasValue() {
		return res
	}
	profile := profile_.Value
	g_ := profile.GetGraph()
	if !g_.HasValue() {
		return BadRequest("Graph not found")
	}
	g := g_.Value
	att := profile.GetAttributes()
	loc := [2]float32{req.Locations[0][0], req.Locations[0][1]}
	resp := isochrone.ComputeIsochrone(g, att, loc, req.Range)
	return OK(&resp)
}

func _HandleIsochroneRequest(req IsochroneRequest) Result {
	// get profile
	profile_, res := GetRequestProfile(MANAGER, "driving-car", "time")
	if !profile_.HasValue() {
		return res
	}
	profile := profile_.Value
	g_ := profile.GetGraph()
	if !g_.HasValue() {
		return BadRequest("Graph not found")
	}
	g := g_.Value
	att := profile.GetAttributes()

	start := geo.Coord{req.Locations[0][0], req.Locations[0][1]}
	consumer := &SPTIsochroneConsumer{
		points: NewQuadTree(func(val1, val2 int) int {
			if val1 < val2 {
				return val1
			} else {
				return val2
			}
		}),
		rasterizer: NewDummyRasterizer(1000),
		projection: &WebMercatorProjection{},
		att:        att,
	}
	s_node, _ := att.GetClosestNode(start)
	spt := routing.NewShortestPathTree(g, s_node, req.Range[len(req.Range)-1], consumer)
	slog.Debug(fmt.Sprintf("Start Caluclating shortest-path-tree from %v", start))
	spt.CalcShortestPathTree()
	slog.Debug("shortest-path-tree finished")
	slog.Debug("start building isochrone")
	// create tree containing marching square cells
	mq_tree := NewQuadTree(func(val1, val2 int) int {
		return val2
	})
	cells := consumer.points.ToSlice()
	for _, cell := range cells {
		x := cell.X
		y := cell.Y
		_, exist := mq_tree.Get(x, y)
		if !exist {
			mq_tree.Insert(x, y, _GetMarchingSquareIndex(consumer.points, x, y))
		}
		x = cell.X
		y = cell.Y - 1
		_, exist = mq_tree.Get(x, y)
		if !exist {
			mq_tree.Insert(x, y, _GetMarchingSquareIndex(consumer.points, x, y))
		}
		x = cell.X - 1
		y = cell.Y
		_, exist = mq_tree.Get(x, y)
		if !exist {
			mq_tree.Insert(x, y, _GetMarchingSquareIndex(consumer.points, x, y))
		}
		x = cell.X - 1
		y = cell.Y - 1
		_, exist = mq_tree.Get(x, y)
		if !exist {
			mq_tree.Insert(x, y, _GetMarchingSquareIndex(consumer.points, x, y))
		}
	}
	// type Object struct {
	// 	X int32 `json:"x"`
	// 	Y int32 `json:"y"`
	// 	V int   `json:"v"`
	// }
	// object := make([]Object, 0)
	// for _, cell := range cells {
	// 	object = append(object, Object{X: cell.X, Y: cell.Y, V: cell.Value})
	// }
	// WriteJSONToFile(object, "cells.json")
	// object = make([]Object, 0)
	// nodes := mq_tree.ToSlice()
	// for _, node := range nodes {
	// 	object = append(object, Object{X: node.X, Y: node.Y, V: node.Value})
	// }
	// WriteJSONToFile(object, "mq_tree.json")
	// iterate tree to build polygons
	polygons := NewList[geo.CoordArray](0)
	rasterizer := NewDummyRasterizer(1000)
	projection := WebMercatorProjection{}
	nodes := mq_tree.ToSlice()
	for _, node := range nodes {
		if node.Value == -1 {
			continue
		}
		startx := node.X
		starty := node.Y
		currindex := node.Value
		if currindex == 0 || currindex == 15 {
			mq_tree.Update(startx, starty, -1)
			continue
		}
		currx := startx
		curry := starty
		currdirection := -1
		polygon := make(geo.CoordArray, 0)
		for {
			nextx, nexty, nextdirection, coord, replace := _ProcessSquare(currindex, currdirection, currx, curry, rasterizer)
			if replace == -1 {
				mq_tree.Update(currx, curry, -1)
			} else {
				mq_tree.Update(currx, curry, replace)
			}
			// add coordinate to polygon
			polygon = append(polygon, projection.ReProj(coord))
			if nextx == startx && nexty == starty {
				polygons.Add(polygon)
				break
			}
			nextindex, ok := mq_tree.Get(nextx, nexty)
			if !ok || nextindex == -1 {
				break
				panic("not implemented")
			}
			currx = nextx
			curry = nexty
			currdirection = nextdirection
			currindex = nextindex
		}
	}
	slog.Debug("isochrone built")

	slog.Debug("start building response")
	// seperate outer and inner polygons
	outers := NewList[geo.CoordArray](0)
	inners := NewList[Tuple[geo.CoordArray, int]](0)
	for _, polygon := range polygons {
		if _PolygonOrientation(polygon) {
			outers.Add(polygon)
		} else {
			inners.Add(MakeTuple(polygon, -1))
		}
	}
	for i, tuple := range inners {
		inner := tuple.A
		outerid := -1
		outerlength := math.MaxInt
		for j, outer := range outers {
			// inner polygons will always be either fully within or not due to marching squares
			if geo.SimplePointInPolygon(inner[0], [][]geo.Coord{outer}) {
				// take the smallest outer polygons
				if outerid == -1 || len(outer) < outerlength {
					outerid = j
					outerlength = len(outer)
				}
			}
		}
		if outerid != -1 {
			inners[i] = MakeTuple(inner, outerid)
		}
	}
	features := NewList[geo.Feature](outers.Length())
	for _, outer := range outers {
		polygon := [][]geo.Coord{outer}
		// for _, tuple := range inners {
		// 	inner := tuple.A
		// 	outerid := tuple.B
		// 	if outerid == i {
		// 		polygon = append(polygon, inner)
		// 	}
		// }
		geometry := geo.NewPolygon(polygon)
		properties := NewDict[string, any](0)
		features.Add(geo.NewFeature(&geometry, properties))
	}
	resp := geo.NewFeatureCollection(features)
	slog.Debug("reponse build")
	return OK(&resp)
}

//**********************************************************
// isochrone builder
//**********************************************************

type SPTIsochroneConsumer struct {
	points     *QuadTree[int]
	rasterizer IRasterizer
	projection IProjection
	att        attr.IAttributes
}

func (self *SPTIsochroneConsumer) ConsumePoint(point geo.Coord, value int) {
	point = self.projection.Proj(point)
	x, y := self.rasterizer.PointToIndex(point)
	self.points.Insert(x, y, value)
}

func (self *SPTIsochroneConsumer) ConsumeEdge(edge int32, start_value int, end_value int) {
	geom := self.att.GetEdgeGeom(edge)
	for i, _ := range geom {
		geom[i] = self.projection.Proj(geom[i])
	}
	total_length := float32(0)
	for i := 0; i < len(geom)-1; i++ {
		total_length += _Dist(geom[i], geom[i+1])
	}
	callback := func(point geo.Coord, length float32) {
		x, y := self.rasterizer.PointToIndex(point)
		value := start_value + int(float32(end_value-start_value)*length/total_length)
		self.points.Insert(x, y, value)
	}
	_SampleAlongLine(geom, 50, callback)
}

func _Dist(start geo.Coord, end geo.Coord) float32 {
	return float32(math.Sqrt(math.Pow(float64(start[0]-end[0]), 2) + math.Pow(float64(start[1]-end[1]), 2)))
}

func _PointInDist(start geo.Coord, end geo.Coord, dist float32) geo.Coord {
	dx := end[0] - start[0]
	dy := end[1] - start[1]
	d := _Dist(start, end)
	return geo.Coord{start[0] + dx*dist/d, start[1] + dy*dist/d}
}

func _SampleAlongLine(line geo.CoordArray, dist float32, callback func(geo.Coord, float32)) {
	length := float32(0)
	callback(line[0], length)
	for i := 0; i < len(line)-1; i++ {
		curr_start := line[i]
		curr_end := line[i+1]
		curr_len := _Dist(curr_start, curr_end)
		for {
			if dist > curr_len {
				length += curr_len
				callback(curr_end, length)
				break
			}
			curr_start = _PointInDist(curr_start, curr_end, dist)
			length += dist
			callback(curr_start, length)
			curr_len = _Dist(curr_start, curr_end)
		}
	}
}

// Computes the orientation of the polygon (clockwise or counterclockwise)
//
// # Return true if the polygon is counter-clockwise
//
// Note: this function also works robustely for non-convex polygons frequently encountered for isochrones
func _PolygonOrientation(polygon geo.CoordArray) bool {
	orientation := float64(0)
	for i := 0; i < len(polygon); i++ {
		curr := polygon[i]
		last := curr
		next := curr
		if i == 0 {
			last = polygon[len(polygon)-1]
		} else {
			last = polygon[i-1]
		}
		if i == len(polygon)-1 {
			next = polygon[0]
		} else {
			next = polygon[i+1]
		}
		vector1 := [2]float64{float64(curr[0]) - float64(last[0]), float64(curr[1]) - float64(last[1])}
		vector2 := [2]float64{float64(next[0]) - float64(curr[0]), float64(next[1]) - float64(curr[1])}
		// normalize vectors
		length1 := math.Sqrt(math.Pow(vector1[0], 2) + math.Pow(vector1[1], 2))
		length2 := math.Sqrt(math.Pow(vector2[0], 2) + math.Pow(vector2[1], 2))
		vector1 = [2]float64{vector1[0] / length1, vector1[1] / length1}
		vector2 = [2]float64{vector2[0] / length2, vector2[1] / length2}
		// distances along and tangent to first vector
		dot := vector1[0]*vector2[0] + vector1[1]*vector2[1]
		cross := vector1[0]*vector2[1] - vector1[1]*vector2[0]
		// compute angles
		orientation += math.Pi - math.Atan2(-cross, dot)
	}
	innersum := float64((len(polygon) - 2)) * math.Pi
	if orientation-innersum < 0.0001 {
		return false
	}
	return true
}

//**********************************************************
// marching squares
//**********************************************************

func _GetMarchingSquareIndex(tree *QuadTree[int], x, y int32) int {
	_, active0 := tree.Get(x, y)
	_, active1 := tree.Get(x+1, y)
	_, active3 := tree.Get(x, y+1)
	_, active2 := tree.Get(x+1, y+1)
	key := 0
	if active0 {
		key += 1
	}
	if active1 {
		key += 2
	}
	if active2 {
		key += 4
	}
	if active3 {
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
		new_direction, replace_index = _ProcessSquare0(index, direction)
	case 1:
		new_direction, replace_index = _ProcessSquare1(index, direction)
	case 2:
		new_direction, replace_index = _ProcessSquare2(index, direction)
	case 3:
		new_direction, replace_index = _ProcessSquare3(index, direction)
	case 4:
		new_direction, replace_index = _ProcessSquare4(index, direction)
	case 5:
		new_direction, replace_index = _ProcessSquare5(index, direction)
	case 6:
		new_direction, replace_index = _ProcessSquare6(index, direction)
	case 7:
		new_direction, replace_index = _ProcessSquare7(index, direction)
	case 8:
		new_direction, replace_index = _ProcessSquare8(index, direction)
	case 9:
		new_direction, replace_index = _ProcessSquare9(index, direction)
	case 10:
		new_direction, replace_index = _ProcessSquare10(index, direction)
	case 11:
		new_direction, replace_index = _ProcessSquare11(index, direction)
	case 12:
		new_direction, replace_index = _ProcessSquare12(index, direction)
	case 13:
		new_direction, replace_index = _ProcessSquare13(index, direction)
	case 14:
		new_direction, replace_index = _ProcessSquare14(index, direction)
	case 15:
		new_direction, replace_index = _ProcessSquare15(index, direction)
	}
	loc := rasterizer.IndexToPoint(x, y)
	switch direction {
	case 0:
		loc[0] += 50
	case 1:
		loc[1] += 50
	case 2:
		loc[0] += 50
		loc[1] += 100
	case 3:
		loc[0] += 100
		loc[1] += 50
	}
	loc[0] += 50
	loc[1] += 50
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

func _GetDefaultDirection(index int) int {
	// make shure the direction is counter-clockwise for outer polygons
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

func _ProcessSquare0(index int, direction int) (int, int) {
	panic("this square should never be traversed")
}

func _ProcessSquare1(index int, direction int) (int, int) {
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

func _ProcessSquare2(index int, direction int) (int, int) {
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

func _ProcessSquare3(index int, direction int) (int, int) {
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

func _ProcessSquare4(index int, direction int) (int, int) {
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

func _ProcessSquare5(index int, direction int) (int, int) {
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

func _ProcessSquare6(index int, direction int) (int, int) {
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

func _ProcessSquare7(index int, direction int) (int, int) {
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

func _ProcessSquare8(index int, direction int) (int, int) {
	return _ProcessSquare7(index, direction)
}

func _ProcessSquare9(index int, direction int) (int, int) {
	return _ProcessSquare6(index, direction)
}

func _ProcessSquare10(index int, direction int) (int, int) {
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

func _ProcessSquare11(index int, direction int) (int, int) {
	return _ProcessSquare4(index, direction)
}

func _ProcessSquare12(index int, direction int) (int, int) {
	return _ProcessSquare3(index, direction)
}

func _ProcessSquare13(index int, direction int) (int, int) {
	return _ProcessSquare2(index, direction)
}

func _ProcessSquare14(index int, direction int) (int, int) {
	return _ProcessSquare1(index, direction)
}

func _ProcessSquare15(index int, direction int) (int, int) {
	return _ProcessSquare0(index, direction)
}
