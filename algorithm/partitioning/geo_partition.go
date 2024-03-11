package partitioning

import (
	"fmt"
	"strconv"

	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

// computes node tiles based on geo-polygons
func GeometricPartitioning(g graph.IGraph, features []geo.Feature) Array[int16] {
	node_tiles := make([]int16, g.NodeCount())

	c := 0
	for i := 0; i < int(g.NodeCount()); i++ {
		node := g.GetNodeGeom(int32(i))
		if c%1000 == 0 {
			slog.Debug(fmt.Sprintf("finished node: %v", c))
		}
		point := geo.NewPoint(node)
		node_tiles[i] = -1
		for _, feature := range features {
			polygon := feature.Geometry()
			if polygon.Contains(&point) {
				tile_id := feature.Properties()["TileID"]
				id, _ := strconv.Atoi(tile_id.(string))
				node_tiles[i] = int16(id)
				break
			}
		}
		c += 1
	}

	return node_tiles
}
