package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/decay"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//**************************************************
// simple 2sfca only descending into active cells
//**************************************************

func CalcTiled2SFCA(g graph.ITiledGraph, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// get closest node for every demand point
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	active_tiles := NewArray[bool](int(g.TileCount()) + 2)
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
			tile := g.GetNodeTile(id)
			active_tiles[tile] = true
		} else {
			demand_nodes[i] = -1
		}
	}
	// get closest node for every supply point and write to chan
	supply_chan := make(chan Tuple[int32, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		w := sup.GetWeight(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			supply_chan <- MakeTuple(id, float32(w))
		} else {
			supply_chan <- MakeTuple(int32(-1), float32(w))
		}
	}
	close(supply_chan)

	// create array containing accessibility results
	access := NewArray[float32](dem.PointCount())

	// compute 2sfca
	max_range := int32(dec.GetMaxDistance())
	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			// init routing components
			visited := NewArray[bool](g.NodeCount())
			dist := NewArray[int32](g.NodeCount())
			heap := NewPriorityQueue[int32, int32](100)
			explorer := g.GetDefaultExplorer()

			for {
				// read supply entry from chan
				temp, ok := <-supply_chan
				if !ok {
					break
				}
				s_id := temp.A
				s_weight := temp.B
				if s_id == -1 {
					continue
				}
				s_tile := g.GetNodeTile(s_id)

				// clear for routing
				heap.Clear()
				heap.Enqueue(s_id, 0)
				for i := 0; i < visited.Length(); i++ {
					visited[i] = false
					dist[i] = 1000000000
				}
				dist[s_id] = 0

				// routing loop
				for {
					curr_id, ok := heap.Dequeue()
					if !ok {
						break
					}
					if visited[curr_id] {
						continue
					}
					visited[curr_id] = true
					curr_tile := g.GetNodeTile(curr_id)
					handler := func(ref graph.EdgeRef) {
						other_id := ref.OtherID
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
					}
					if curr_tile == s_tile || active_tiles[curr_tile] {
						explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, handler)
					} else {
						explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_SKIP, handler)
					}
				}

				// compute R-value for facility
				demand_sum := float32(0.0)
				for i, d_node := range demand_nodes {
					if d_node == -1 {
						continue
					}
					if !visited[d_node] {
						continue
					}
					distance_decay := dec.GetDistanceWeight(float32(dist[d_node]))
					demand_sum += float32(dem.GetWeight(i)) * distance_decay
				}
				R := s_weight / demand_sum
				// add new access to reachable demand points
				for i, d_node := range demand_nodes {
					if d_node == -1 {
						continue
					}
					if !visited[d_node] {
						continue
					}
					distance_decay := dec.GetDistanceWeight(float32(dist[d_node]))
					access[i] += R * distance_decay
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return access
}

//*******************************************
// using index within active cells
//*******************************************

func CalcTiled2SFCA2(g *graph.TiledGraph3, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// get closest node for every demand point
	tilecount := g.TileCount() + 2
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	active_tiles := NewArray[bool](int(tilecount))
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
			tile := g.GetNodeTile(id)
			active_tiles[tile] = true
		} else {
			demand_nodes[i] = -1
		}
	}
	// get closest node for every supply point and write to chan
	supply_chan := make(chan Tuple[int32, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		w := sup.GetWeight(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			supply_chan <- MakeTuple(id, float32(w))
		} else {
			supply_chan <- MakeTuple(int32(-1), float32(w))
		}
	}
	close(supply_chan)

	// create array containing accessibility results
	access := NewArray[float32](dem.PointCount())

	// compute 2sfca
	max_range := int32(dec.GetMaxDistance())
	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			// init routing components
			found_tiles := NewArray[bool](int(tilecount))
			visited := NewArray[bool](g.NodeCount())
			dist := NewArray[int32](g.NodeCount())
			heap := NewPriorityQueue[int32, int32](100)
			explorer := g.GetDefaultExplorer()

			for {
				// read supply entry from chan
				temp, ok := <-supply_chan
				if !ok {
					break
				}
				s_id := temp.A
				s_weight := temp.B
				if s_id == -1 {
					continue
				}
				s_tile := g.GetNodeTile(s_id)

				// clear for routing
				heap.Clear()
				heap.Enqueue(s_id, 0)
				for i := 0; i < visited.Length(); i++ {
					visited[i] = false
					dist[i] = 1000000000
				}
				for i := 0; i < int(tilecount); i++ {
					found_tiles[i] = false
				}
				dist[s_id] = 0

				// routing loop
				for {
					curr_id, ok := heap.Dequeue()
					if !ok {
						break
					}
					if visited[curr_id] {
						continue
					}
					visited[curr_id] = true
					curr_tile := g.GetNodeTile(curr_id)
					handler := func(ref graph.EdgeRef) {
						other_id := ref.OtherID
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
					}
					if curr_tile == s_tile {
						explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, handler)
					} else {
						found_tiles[curr_tile] = true
						explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_SKIP, handler)
					}
				}

				for i := 0; i < int(tilecount); i++ {
					if !active_tiles[i] || !found_tiles[i] {
						continue
					}
					down_edges := g.GetDownEdges(int16(i), graph.FORWARD)
					for j := 0; j < down_edges.Length(); j++ {
						edge := down_edges[j]
						curr_len := dist[edge.From]
						new_len := curr_len + edge.Weight
						if dist[edge.To] > new_len {
							dist[edge.To] = new_len
						}
					}
				}

				// compute R-value for facility
				demand_sum := float32(0.0)
				for i, d_node := range demand_nodes {
					if d_node == -1 {
						continue
					}
					d_dist := dist[d_node]
					if d_dist >= 1000000000 {
						continue
					}
					distance_decay := dec.GetDistanceWeight(float32(d_dist))
					demand_sum += float32(dem.GetWeight(i)) * distance_decay
				}
				R := s_weight / demand_sum
				// add new access to reachable demand points
				for i, d_node := range demand_nodes {
					if d_node == -1 {
						continue
					}
					d_dist := dist[d_node]
					if d_dist >= 1000000000 {
						continue
					}
					distance_decay := dec.GetDistanceWeight(float32(d_dist))
					access[i] += R * distance_decay
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return access
}

//*******************************************
// using index within active cells
//*******************************************

func CalcTiled2SFCA3(g *graph.TiledGraph4, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// node queue for selecting graph subset
	node_queue := NewQueue[int32]()

	// get closest node for every demand point
	tilecount := g.TileCount() + 2
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	active_tiles := NewArray[bool](int(tilecount))
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
			tile := g.GetNodeTile(id)
			active_tiles[tile] = true
			node_queue.Push(id)
		} else {
			demand_nodes[i] = -1
		}
	}
	// get closest node for every supply point and write to chan
	supply_chan := make(chan Tuple[int32, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		w := sup.GetWeight(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			supply_chan <- MakeTuple(id, float32(w))
		} else {
			supply_chan <- MakeTuple(int32(-1), float32(w))
		}
	}
	close(supply_chan)

	// select graph subset by marking visited nodes
	explorer := g.GetDefaultExplorer()
	graph_subset := NewArray[bool](int(g.NodeCount()))
	for {
		if node_queue.Size() == 0 {
			break
		}
		node, _ := node_queue.Pop()
		if graph_subset[node] {
			continue
		}
		graph_subset[node] = true
		explorer.ForAdjacentEdges(node, graph.BACKWARD, graph.ADJACENT_UPWARDS, func(ref graph.EdgeRef) {
			if graph_subset[ref.OtherID] {
				return
			}
			node_queue.Push(ref.OtherID)
		})
	}
	// selecting subset of downward edges for linear sweep
	down_edges_subset := NewDict[int16, List[graph.TiledSHEdge]](int(tilecount))
	for i := 0; i < int(tilecount); i++ {
		if !active_tiles[i] {
			continue
		}
		down_edges := g.GetDownEdges(int16(i), graph.FORWARD)
		de_subset := NewList[graph.TiledSHEdge](10)
		for j := 0; j < down_edges.Length(); j++ {
			edge := down_edges[j]
			if !graph_subset[edge.From] {
				continue
			}
			de_subset.Add(edge)
		}
		down_edges_subset[int16(i)] = de_subset
	}

	// create array containing accessibility results
	access := NewArray[float32](dem.PointCount())

	// compute 2sfca
	max_range := int32(dec.GetMaxDistance())
	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			// init routing components
			found_tiles := NewArray[bool](int(tilecount))
			visited := NewArray[bool](g.NodeCount())
			dist := NewArray[int32](g.NodeCount())
			heap := NewPriorityQueue[int32, int32](100)
			explorer := g.GetDefaultExplorer()

			for {
				// read supply entry from chan
				temp, ok := <-supply_chan
				if !ok {
					break
				}
				s_id := temp.A
				s_weight := temp.B
				if s_id == -1 {
					continue
				}
				s_tile := g.GetNodeTile(s_id)

				// clear for routing
				heap.Clear()
				heap.Enqueue(s_id, 0)
				for i := 0; i < visited.Length(); i++ {
					visited[i] = false
					dist[i] = 1000000000
				}
				for i := 0; i < int(tilecount); i++ {
					found_tiles[i] = false
				}
				dist[s_id] = 0

				// routing loop
				for {
					curr_id, ok := heap.Dequeue()
					if !ok {
						break
					}
					if visited[curr_id] {
						continue
					}
					visited[curr_id] = true
					curr_tile := g.GetNodeTile(curr_id)
					handler := func(ref graph.EdgeRef) {
						other_id := ref.OtherID
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
					}
					if curr_tile == s_tile {
						explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, handler)
					} else {
						found_tiles[curr_tile] = true
						explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_SKIP, handler)
					}
				}

				for i := 0; i < int(tilecount); i++ {
					if !active_tiles[i] || !found_tiles[i] {
						continue
					}
					down_edges := down_edges_subset[int16(i)]
					for j := 0; j < down_edges.Length(); j++ {
						edge := down_edges[j]
						curr_len := dist[edge.From]
						new_len := curr_len + edge.Weight
						if dist[edge.To] > new_len {
							dist[edge.To] = new_len
						}
					}
				}

				// compute R-value for facility
				demand_sum := float32(0.0)
				for i, d_node := range demand_nodes {
					if d_node == -1 {
						continue
					}
					d_dist := dist[d_node]
					if d_dist >= 1000000000 {
						continue
					}
					distance_decay := dec.GetDistanceWeight(float32(d_dist))
					demand_sum += float32(dem.GetWeight(i)) * distance_decay
				}
				R := s_weight / demand_sum
				// add new access to reachable demand points
				for i, d_node := range demand_nodes {
					if d_node == -1 {
						continue
					}
					d_dist := dist[d_node]
					if d_dist >= 1000000000 {
						continue
					}
					distance_decay := dec.GetDistanceWeight(float32(d_dist))
					access[i] += R * distance_decay
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return access
}
