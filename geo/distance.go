package geo

import (
	"math"
)

func HaversineDistance(from, to Coord) float64 {
	r := 6365000.0
	lat1 := from.Lat * math.Pi / 180
	lat2 := to.Lat * math.Pi / 180
	lon1 := from.Lon * math.Pi / 180
	lon2 := to.Lon * math.Pi / 180
	a := math.Pow(math.Sin(float64((lat2-lat1)/2)), 2)
	b := math.Pow(math.Sin(float64((lon2-lon1)/2)), 2)
	return 2.0 * r * math.Asin(math.Sqrt(a+math.Cos(float64(lat1))*math.Cos(float64(lat2))*b))
}

func EuclideanDistance(a, b Coord) float64 {
	d_lon := float64(a.Lon) - float64(b.Lon)
	d_lat := float64(a.Lat) - float64(b.Lat)
	return math.Sqrt(math.Pow(d_lon, 2) + math.Pow(d_lat, 2))
}
