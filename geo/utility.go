package geo

func CalcEnvelope(geom Geometry) Envelope {
	if geom.Type() == "Point" {
		point := geom.(*Point)
		coord := point.Coordinates()
		return Envelope{coord[0], coord[1], coord[0], coord[1]}
	} else if geom.Type() == "MultiPoint" || geom.Type() == "LineString" {
		var coords []Coord
		if geom.Type() == "MultiPoint" {
			points := geom.(*MultiPoint)
			coords = points.Coordinates()
		}
		if geom.Type() == "LineString" {
			line := geom.(*LineString)
			coords = line.Coordinates()
		}
		envelope := Envelope{coords[0][0], coords[0][1], coords[0][0], coords[0][1]}
		for i := 1; i < len(coords); i++ {
			coord := coords[i]
			if coord[0] < envelope[0] {
				envelope[0] = coord[0]
			} else if coord[0] > envelope[2] {
				envelope[2] = coord[0]
			}
			if coord[1] < envelope[1] {
				envelope[1] = coord[1]
			} else if coord[1] > envelope[3] {
				envelope[3] = coord[1]
			}
		}
		return envelope
	} else if geom.Type() == "MultiLineString" || geom.Type() == "Polygon" {
		var coords [][]Coord
		if geom.Type() == "MultiLineString" {
			lines := geom.(*MultiLineString)
			coords = lines.Coordinates()
		}
		if geom.Type() == "Polygon" {
			polygon := geom.(*Polygon)
			coords = polygon.Coordinates()
		}
		envelope := Envelope{coords[0][0][0], coords[0][0][1], coords[0][0][0], coords[0][0][1]}
		for i := 0; i < len(coords); i++ {
			c := coords[i]
			for j := 0; j < len(c); j++ {
				coord := c[j]
				if coord[0] < envelope[0] {
					envelope[0] = coord[0]
				} else if coord[0] > envelope[2] {
					envelope[2] = coord[0]
				}
				if coord[1] < envelope[1] {
					envelope[1] = coord[1]
				} else if coord[1] > envelope[3] {
					envelope[3] = coord[1]
				}
			}
		}
		return envelope
	} else {
		polygons := geom.(*MultiPolygon)
		coords := polygons.Coordinates()
		envelope := Envelope{coords[0][0][0][0], coords[0][0][0][1], coords[0][0][0][0], coords[0][0][0][1]}
		for i := 0; i < len(coords); i++ {
			c1 := coords[i]
			for j := 0; j < len(c1); j++ {
				c2 := c1[j]
				for k := 0; k < len(c2); k++ {
					coord := c2[k]
					if coord[0] < envelope[0] {
						envelope[0] = coord[0]
					} else if coord[0] > envelope[2] {
						envelope[2] = coord[0]
					}
					if coord[1] < envelope[1] {
						envelope[1] = coord[1]
					} else if coord[1] > envelope[3] {
						envelope[3] = coord[1]
					}
				}
			}
		}
		return envelope
	}
}
