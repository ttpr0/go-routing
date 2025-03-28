package isochrone

import (
	"math"

	"github.com/ttpr0/go-routing/geo"
)

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
