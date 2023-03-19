package geo

import (
	"math"
)

func HaversineDistance(from, to Coord) float64 {
	r := 6365000.0
	lat1 := from[1] * math.Pi / 180
	lat2 := to[1] * math.Pi / 180
	lon1 := from[0] * math.Pi / 180
	lon2 := to[0] * math.Pi / 180
	a := math.Pow(math.Sin(float64((lat2-lat1)/2)), 2)
	b := math.Pow(math.Sin(float64((lon2-lon1)/2)), 2)
	return 2.0 * r * math.Asin(math.Sqrt(a+math.Cos(float64(lat1))*math.Cos(float64(lat2))*b))
}

func HaversineLength(points CoordArray) float64 {
	r := 6365000.0
	length := float64(0)
	for i := 0; i < len(points)-1; i++ {
		lat1 := points[i][1] * math.Pi / 180
		lat2 := points[i+1][1] * math.Pi / 180
		lon1 := points[i][0] * math.Pi / 180
		lon2 := points[i+1][0] * math.Pi / 180
		a := math.Pow(math.Sin(float64(lat2-lat1)/2), 2)
		b := math.Pow(math.Sin(float64(lon2-lon1)/2), 2)
		length += 2 * r * math.Asin(math.Sqrt(a+math.Cos(float64(lat1))*math.Cos(float64(lat2))*b))
	}
	return length
}

func EuclideanDistance(a, b Coord) float64 {
	d_lon := float64(a[0]) - float64(b[0])
	d_lat := float64(a[1]) - float64(b[1])
	return math.Sqrt(math.Pow(d_lon, 2) + math.Pow(d_lat, 2))
}
