package algorithm

import (
	"sync"

	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

func DijkstraTDMatrix(g graph.IGraph, sources Array[int32], destinations Array[int32], max_range float32) Matrix[float32] {
	source_chan := make(chan Tuple[int, int32], sources.Length())
	for i := 0; i < sources.Length(); i++ {
		source_chan <- MakeTuple(i, sources[i])
	}
	close(source_chan)
	is_destination := NewArray[bool](g.NodeCount())
	destination_count := 0
	for i := 0; i < destinations.Length(); i++ {
		node := destinations[i]
		if node != -1 {
			is_destination[node] = true
			destination_count += 1
		}
	}

	matrix := NewMatrix[float32](sources.Length(), destinations.Length())
	wg := sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			// init routing components
			visited := NewArray[bool](g.NodeCount())
			dist := NewArray[int32](g.NodeCount())
			heap := NewPriorityQueue[int32, int32](100)
			explorer := g.GetGraphExplorer()

			for {
				// read supply entry from chan
				temp, ok := <-source_chan
				if !ok {
					break
				}
				s := temp.A
				s_node := temp.B

				// if no node set all distances to -1
				if s_node == -1 {
					for i := 0; i < destinations.Length(); i++ {
						matrix.Set(s, i, -1)
					}
					continue
				}

				// clear for routing
				heap.Clear()
				heap.Enqueue(s_node, 0)
				for i := 0; i < visited.Length(); i++ {
					visited[i] = false
					dist[i] = 1000000000
				}
				dist[s_node] = 0

				// routing loop
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
					if is_destination[curr_id] {
						visited_count += 1
						if visited_count >= destination_count {
							break
						}
					}
					explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
						other_id := ref.OtherID
						if visited[other_id] {
							return
						}
						new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
						if new_length > int32(max_range) {
							return
						}
						if dist[other_id] > new_length {
							dist[other_id] = new_length
							heap.Enqueue(other_id, new_length)
						}
					})
				}

				// set distances in matrix
				for d, d_node := range destinations {
					if d_node == -1 {
						matrix.Set(s, d, -1)
						continue
					}
					if !visited[d_node] {
						matrix.Set(s, d, -1)
						continue
					}
					matrix.Set(s, d, float32(dist[d_node]))
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return matrix
}
