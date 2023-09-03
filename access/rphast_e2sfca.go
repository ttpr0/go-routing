package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/decay"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

// *******************************************
// simple 2sfca using rphast
// *******************************************

func CalcRPHAST2SFCA(g *graph.CHGraph3, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// node queue for selecting graph subset
	node_queue := NewQueue[int32]()

	// get closest node for every demand point
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
			node_queue.Push(id)
		} else {
			demand_nodes[i] = -1
		}
	}

	// get closest node for every supply point and write to chan
	supply_chan := make(chan Tuple[int32, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		weight := sup.GetWeight(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			supply_chan <- MakeTuple(id, float32(weight))
		} else {
			supply_chan <- MakeTuple(int32(-1), float32(weight))
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
		node_level := g.GetNodeLevel(node)
		edges := explorer.GetAdjacentEdges(node, graph.BACKWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if graph_subset[ref.OtherID] {
				continue
			}
			if node_level >= g.GetNodeLevel(ref.OtherID) {
				continue
			}
			node_queue.Push(ref.OtherID)
		}
	}
	// selecting subset of downward edges for linear sweep
	down_edges_subset := NewList[graph.CHEdge](dem.PointCount())
	down_edges := g.GetDownEdges(graph.FORWARD)
	for i := 0; i < len(down_edges); i++ {
		edge := down_edges[i]
		if !graph_subset[edge.From] {
			continue
		}
		down_edges_subset.Add(edge)
	}

	// create array containing accessibility results
	access := NewArray[float32](dem.PointCount())

	// compute 2sfca
	// max_range := int32(dec.GetMaxDistance())
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
					edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_UPWARDS)
					for {
						ref, ok := edges.Next()
						if !ok {
							break
						}
						other_id := ref.OtherID
						if visited[other_id] {
							continue
						}
						new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
						if dist[other_id] > new_length {
							dist[other_id] = new_length
							heap.Enqueue(other_id, new_length)
						}
					}
				}
				// downwards sweep
				for i := 0; i < len(down_edges_subset); i++ {
					edge := down_edges_subset[i]
					curr_len := dist[edge.From]
					new_len := curr_len + edge.Weight
					if dist[edge.To] > new_len {
						dist[edge.To] = new_len
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

// *******************************************
// phast with range restriction
// *******************************************

func CalcRPHAST2SFCA2(g *graph.CHGraph3, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// node queue for selecting graph subset
	node_queue := NewQueue[int32]()

	// get closest node for every demand point
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
			node_queue.Push(id)
		} else {
			demand_nodes[i] = -1
		}
	}

	// get closest node for every supply point and write to chan
	supply_chan := make(chan Tuple[int32, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		weight := sup.GetWeight(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			supply_chan <- MakeTuple(id, float32(weight))
		} else {
			supply_chan <- MakeTuple(int32(-1), float32(weight))
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
					edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_UPWARDS)
					for {
						ref, ok := edges.Next()
						if !ok {
							break
						}
						other_id := ref.OtherID
						if visited[other_id] {
							continue
						}
						new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
						if new_length > max_range {
							continue
						}
						if dist[other_id] > new_length {
							dist[other_id] = new_length
							heap.Enqueue(other_id, new_length)
						}
					}
				}
				// downwards sweep
				down_edges := g.GetDownEdges(graph.FORWARD)
				for i := 0; i < len(down_edges); i++ {
					edge := down_edges[i]
					curr_len := dist[edge.From]
					if curr_len > max_range {
						continue
					}
					new_len := curr_len + edge.Weight
					if dist[edge.To] > new_len {
						dist[edge.To] = new_len
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

// *******************************************
// rphast with range
// *******************************************

func CalcRPHAST2SFCA3(g *graph.CHGraph3, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// node queue for selecting graph subset
	node_queue := NewQueue[int32]()

	// get closest node for every demand point
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
			node_queue.Push(id)
		} else {
			demand_nodes[i] = -1
		}
	}

	// get closest node for every supply point and write to chan
	supply_chan := make(chan Tuple[int32, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		weight := sup.GetWeight(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			supply_chan <- MakeTuple(id, float32(weight))
		} else {
			supply_chan <- MakeTuple(int32(-1), float32(weight))
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
		node_level := g.GetNodeLevel(node)
		edges := explorer.GetAdjacentEdges(node, graph.BACKWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if graph_subset[ref.OtherID] {
				continue
			}
			if node_level >= g.GetNodeLevel(ref.OtherID) {
				continue
			}
			node_queue.Push(ref.OtherID)
		}
	}
	// selecting subset of downward edges for linear sweep
	down_edges_subset := NewList[graph.CHEdge](dem.PointCount())
	down_edges := g.GetDownEdges(graph.FORWARD)
	for i := 0; i < len(down_edges); i++ {
		edge := down_edges[i]
		if !graph_subset[edge.From] {
			continue
		}
		down_edges_subset.Add(edge)
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
					edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_UPWARDS)
					for {
						ref, ok := edges.Next()
						if !ok {
							break
						}
						other_id := ref.OtherID
						if visited[other_id] {
							continue
						}
						new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
						if new_length > max_range {
							continue
						}
						if dist[other_id] > new_length {
							dist[other_id] = new_length
							heap.Enqueue(other_id, new_length)
						}
					}
				}
				// downwards sweep
				for i := 0; i < len(down_edges_subset); i++ {
					edge := down_edges_subset[i]
					curr_len := dist[edge.From]
					if curr_len > max_range {
						continue
					}
					new_len := curr_len + edge.Weight
					if dist[edge.To] > new_len {
						dist[edge.To] = new_len
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

// *******************************************
// rphast with range selected edge subset
// *******************************************

func CalcRPHAST2SFCA4(g *graph.CHGraph3, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	// node queue for selecting graph subset
	node_queue := NewPriorityQueue[int32, int32](10000)

	// get closest node for every demand point
	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
			node_queue.Enqueue(id, 0)
		} else {
			demand_nodes[i] = -1
		}
	}

	// get closest node for every supply point and write to chan
	supply_chan := make(chan Tuple[int32, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		weight := sup.GetWeight(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			supply_chan <- MakeTuple(id, float32(weight))
		} else {
			supply_chan <- MakeTuple(int32(-1), float32(weight))
		}
	}
	close(supply_chan)

	// set max range
	max_range := int32(dec.GetMaxDistance())

	// select graph subset by marking visited nodes
	explorer := g.GetDefaultExplorer()
	lengths := NewArray[int32](g.NodeCount())
	graph_subset := NewArray[bool](g.NodeCount())
	for {
		node, ok := node_queue.Dequeue()
		if !ok {
			break
		}
		if graph_subset[node] {
			continue
		}
		graph_subset[node] = true
		node_level := g.GetNodeLevel(node)
		node_len := lengths[node]
		edges := explorer.GetAdjacentEdges(node, graph.BACKWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if graph_subset[ref.OtherID] {
				continue
			}
			if node_level >= g.GetNodeLevel(ref.OtherID) {
				continue
			}
			new_len := node_len + explorer.GetEdgeWeight(ref)
			if new_len > max_range {
				continue
			}
			if new_len < lengths[ref.OtherID] {
				lengths[ref.OtherID] = new_len
				node_queue.Enqueue(ref.OtherID, new_len)
			}
		}
	}
	// selecting subset of downward edges for linear sweep
	down_edges_subset := NewList[graph.CHEdge](dem.PointCount())
	down_edges := g.GetDownEdges(graph.FORWARD)
	for i := 0; i < len(down_edges); i++ {
		edge := down_edges[i]
		if !graph_subset[edge.From] {
			continue
		}
		down_edges_subset.Add(edge)
	}

	// create array containing accessibility results
	access := NewArray[float32](dem.PointCount())

	// compute 2sfca
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
					edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_UPWARDS)
					for {
						ref, ok := edges.Next()
						if !ok {
							break
						}
						other_id := ref.OtherID
						if visited[other_id] {
							continue
						}
						new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
						if new_length > max_range {
							continue
						}
						if dist[other_id] > new_length {
							dist[other_id] = new_length
							heap.Enqueue(other_id, new_length)
						}
					}
				}
				// downwards sweep
				for i := 0; i < len(down_edges_subset); i++ {
					edge := down_edges_subset[i]
					curr_len := dist[edge.From]
					if curr_len > max_range {
						continue
					}
					new_len := curr_len + edge.Weight
					if dist[edge.To] > new_len {
						dist[edge.To] = new_len
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
