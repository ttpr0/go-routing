package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/decay"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// simple 2sfca using forward dijkstra
//*******************************************

func CalcDijkstra2SFCA(g graph.IGraph, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// get closest node for every demand point
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
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
			explorer := g.GetGraphExplorer()

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
					explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
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
					})
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
// computed demand_sum during routing
//*******************************************

func CalcDijkstra2SFCA2(g graph.IGraph, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// get closest node for every demand point
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	is_demand := NewArray[int32](g.NodeCount())
	demand_count := 0
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
			is_demand[id] = dem.GetWeight(i)
			demand_count += 1
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
			explorer := g.GetGraphExplorer()

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

				// clear for routing
				heap.Clear()
				heap.Enqueue(s_id, 0)
				for i := 0; i < visited.Length(); i++ {
					visited[i] = false
					dist[i] = 1000000000
				}
				dist[s_id] = 0

				// routing loop
				demand_sum := float32(0.0)
				visited_count := 0
				for {
					curr_id, ok := heap.Dequeue()
					if !ok {
						break
					}
					if visited[curr_id] {
						continue
					}
					visited[curr_id] = true
					if is_demand[curr_id] != 0 {
						distance_decay := dec.GetDistanceWeight(float32(dist[curr_id]))
						demand_sum += float32(is_demand[curr_id]) * distance_decay
						visited_count += 1
						if visited_count >= demand_count {
							break
						}
					}
					explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
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
					})
				}

				// compute R-value for facility
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
