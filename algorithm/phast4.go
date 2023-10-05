package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcPHAST4(g *graph.CHGraph4, start int32, max_range int32) Array[int32] {
	active_tiles := NewArray[bool](g.TileCount())
	visited := NewArray[bool](g.NodeCount())
	dist := NewArray[int32](g.NodeCount())
	for i := 0; i < len(dist); i++ {
		dist[i] = 1000000000
	}
	dist[start] = 0

	heap := NewPriorityQueue[int32, int32](100)
	heap.Enqueue(start, 0)

	explorer := g.GetDefaultExplorer()

	// upwards search
	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			break
		}
		//curr := (*d.graph).GetNode(curr_id)
		if visited[curr_id] {
			continue
		}
		visited[curr_id] = true
		curr_tile := g.GetNodeTile(curr_id)
		active_tiles[curr_tile] = true
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			if g.GetNodeLevel(other_id) <= g.GetNodeLevel(curr_id) {
				return
			}
			if visited[other_id] {
				return
			}
			new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
			if new_length > max_range {
				return
			}
			if dist[other_id] > new_length {
				dist[other_id] = new_length
				heap.Enqueue(other_id, new_length)
			}
		})
	}
	// iterative down-sweep
	down_edges := g.GetDownEdges(graph.FORWARD)
	overlay_dummy := down_edges[0]
	overlay_start := 1
	overlay_end := 1 + overlay_dummy.To
	for i := overlay_start; i < int(overlay_end); i++ {
		edge := down_edges[i]
		curr_len := dist[edge.From]
		new_len := curr_len + edge.Weight
		if new_len > max_range {
			continue
		}
		if dist[edge.To] > new_len {
			dist[edge.To] = new_len
			active_tiles[edge.ToTile] = true
		}
	}
	for i := int(overlay_end); i < down_edges.Length(); i++ {
		curr_dummy := down_edges[i]
		curr_tile := curr_dummy.ToTile
		curr_count := curr_dummy.To
		if active_tiles[curr_tile] {
			tile_start := i + 1
			tile_end := i + 1 + int(curr_count)
			for j := tile_start; j < tile_end; j++ {
				edge := down_edges[j]
				curr_len := dist[edge.From]
				new_len := curr_len + edge.Weight
				if new_len > max_range {
					continue
				}
				if dist[edge.To] > new_len {
					dist[edge.To] = new_len
				}
			}
		}
		i += int(curr_count)
	}

	return dist
}

func CalcPHAST5(g *graph.CHGraph4, start int32, max_range int32) Array[int32] {
	active_tiles := NewArray[bool](g.TileCount())
	visited := NewArray[bool](g.NodeCount())
	dist := NewArray[int32](g.NodeCount())
	for i := 0; i < len(dist); i++ {
		dist[i] = 1000000000
	}
	dist[start] = 0

	heap := NewPriorityQueue[int32, int32](100)
	heap.Enqueue(start, 0)

	explorer := g.GetDefaultExplorer()

	// upwards search
	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			break
		}
		//curr := (*d.graph).GetNode(curr_id)
		if visited[curr_id] {
			continue
		}
		visited[curr_id] = true
		curr_tile := g.GetNodeTile(curr_id)
		active_tiles[curr_tile] = true
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			if g.GetNodeLevel(other_id) <= g.GetNodeLevel(curr_id) {
				return
			}
			if visited[other_id] {
				return
			}
			new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
			if new_length > max_range {
				return
			}
			if dist[other_id] > new_length {
				dist[other_id] = new_length
				heap.Enqueue(other_id, new_length)
			}
		})
	}
	// iterative down-sweep
	down_edges := g.GetDownEdges(graph.FORWARD)
	overlay_dummy := down_edges[0]
	overlay_start := 1
	overlay_end := 1 + int(overlay_dummy.To)
	for i := overlay_start; i < overlay_end; i++ {
		edge := down_edges[i]
		curr_len := dist[edge.From]
		new_len := curr_len + edge.Weight
		if new_len > max_range {
			continue
		}
		if dist[edge.To] > new_len {
			dist[edge.To] = new_len
			active_tiles[edge.ToTile] = true
		}
	}
	tiles_start := overlay_end
	tiles_end := int(down_edges.Length())
	for i := tiles_start; i < tiles_end; i++ {
		edge := down_edges[i]
		if edge.IsDummy {
			curr_tile := edge.ToTile
			if !active_tiles[curr_tile] {
				i += int(edge.To)
			}
			continue
		}
		curr_len := dist[edge.From]
		new_len := curr_len + edge.Weight
		if new_len > max_range {
			continue
		}
		if dist[edge.To] > new_len {
			dist[edge.To] = new_len
		}
	}

	return dist
}
