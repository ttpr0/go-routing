package geo

// Checks if a point lies within a polygon using ray-casting algorithm.
func SimplePointInPolygon(point Coord, polygon [][]Coord) bool {
	intersect := false
	for _, p := range polygon {
		j := len(p) - 1
		for i := 0; i < len(p); i++ {
			if (p[i][1] > point[1]) != (p[j][1] > point[1]) &&
				point[0] < (p[j][0]-p[i][0])*(point[1]-p[i][1])/(p[j][1]-p[i][1])+p[i][0] {
				intersect = !intersect
			}
			j = i
		}
	}
	return intersect
}
