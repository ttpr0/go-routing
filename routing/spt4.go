package routing

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type DistFlag struct {
	Dist int32
}

type EdgeDistFlag struct {
	Dist  int32
	SDist int32
}

type TransitFlag struct {
	trips List[Tuple[int32, int32]]
}

type ShortestPathTree4 struct {
	heap       PriorityQueue[int32, float64]
	g          *graph.TransitGraph
	node_flags Flags[DistFlag]
	edge_flags Flags[EdgeDistFlag]
	stop_flags Flags[TransitFlag]
	from       int32
	to         int32
}

func NewShortestPathTree4(graph *graph.TransitGraph, from, to int32) *ShortestPathTree4 {
	d := ShortestPathTree4{
		g:    graph,
		from: from,
		to:   to,
	}

	node_flags := NewFlags(int32(graph.NodeCount()), DistFlag{1000000})
	edge_flags := NewFlags[EdgeDistFlag](int32(graph.EdgeCount()), EdgeDistFlag{Dist: 1000000})
	stop_flags := NewFlags[TransitFlag](int32(graph.StopCount()), TransitFlag{})

	d.node_flags = node_flags
	d.edge_flags = edge_flags
	d.stop_flags = stop_flags

	heap := NewPriorityQueue[int32, float64](100)
	d.heap = heap

	return &d
}

func (self *ShortestPathTree4) CalcShortestPathTree(start, max_val int32, consumer ISPTConsumer) {
	self.node_flags.Reset()
	self.edge_flags.Reset()
	self.stop_flags.Reset()
	self.heap.Enqueue(start, 0)
	starts := Array[Tuple[int32, int32]]{MakeTuple(start, int32(0))}
	_CalcTransitDijkstra(self.g, starts, self.node_flags, self.edge_flags, self.stop_flags, max_val, self.from, self.to, consumer)
}

type PQItem struct {
	item int32
	dist int32
}

type TransitItem struct {
	time      int32
	departure int32
	stop      int32
}

func _CalcTransitDijkstra(g *graph.TransitGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], edge_flags Flags[EdgeDistFlag], stop_flags Flags[TransitFlag], max_range int32, from, to int32, consumer ISPTConsumer) {
	// step 1: range-dijkstra from start
	_CalcRangeDijkstraTC(g, starts, node_flags, edge_flags, max_range, consumer)

	// step 2: transit-dijkstra from all found stops
	heap := NewPriorityQueue[TransitItem, int32](100)
	explorer := g.GetTransitExplorer()

	for i := 0; i < g.StopCount(); i++ {
		base_node := g.MapStopToNode(int32(i))
		if base_node == -1 {
			continue
		}
		flag := node_flags.Get(base_node)
		if flag.Dist > max_range {
			continue
		}
		dist := flag.Dist
		time := to + dist
		heap.Enqueue(TransitItem{time, to, int32(i)}, time)
		explorer.ForAdjacentEdges(int32(i), graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if ref.IsShortcut() {
				return
			}
			weights := explorer.GetConnectionWeights(ref, from+dist, to+dist)
			for _, w := range weights {
				heap.Enqueue(TransitItem{w.Departure, w.Departure - dist, int32(i)}, w.Departure)
			}
		})
	}
	for {
		item, ok := heap.Dequeue()
		if !ok {
			break
		}
		curr_flag := stop_flags.Get(item.stop)
		prune := false
		for _, trip := range curr_flag.trips {
			time := trip.A
			departure := trip.B
			if time <= item.time && departure >= item.departure {
				prune = true
				break
			}
		}
		if prune {
			continue
		}
		if curr_flag.trips == nil {
			curr_flag.trips = NewList[Tuple[int32, int32]](4)
		}
		curr_flag.trips.Add(MakeTuple(item.time, item.departure))
		explorer.ForAdjacentEdges(item.stop, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			if ref.IsShortcut() {
				weight := explorer.GetShortcutWeight(ref)
				new_time := item.time + weight
				if new_time-item.departure > max_range {
					return
				}
				heap.Enqueue(TransitItem{new_time, item.departure, other_id}, new_time)
			} else {
				weight := explorer.GetConnectionWeight(ref, item.time)
				if !weight.HasValue() {
					return
				}
				new_time := weight.Value.Arrival
				if new_time-item.departure > max_range {
					return
				}
				heap.Enqueue(TransitItem{new_time, item.departure, other_id}, new_time)
			}
		})
	}

	// step 3: range-dijkstra from all stops
	starts_ := NewList[Tuple[int32, int32]](10)
	for i := 0; i < g.StopCount(); i++ {
		flag := stop_flags.Get(int32(i))
		if flag.trips == nil {
			continue
		}
		dist := int32(100000000)
		for _, trip := range flag.trips {
			time := trip.A
			departure := trip.B
			d := time - departure
			if d < dist {
				dist = d
			}
		}
		if dist > max_range {
			continue
		}
		base_node := g.MapStopToNode(int32(i))
		starts_.Add(MakeTuple(base_node, dist))
	}
	_CalcRangeDijkstraTC(g, Array[Tuple[int32, int32]](starts_), node_flags, edge_flags, max_range, consumer)
}

func _CalcRangeDijkstraTC(g graph.IGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], edge_flags Flags[EdgeDistFlag], max_range int32, consumer ISPTConsumer) {
	heap := NewPriorityQueue[PQItem, int32](100)
	explorer := g.GetGraphExplorer()

	for _, item := range starts {
		start := item.A
		dist := item.B
		start_flag := node_flags.Get(start)
		start_flag.Dist = dist

		explorer.ForAdjacentEdges(start, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			edge_id := ref.EdgeID
			next_node_id := ref.OtherID
			edge_dist := explorer.GetEdgeWeight(ref) + dist
			if edge_dist > max_range {
				return
			}
			edge_flag := edge_flags.Get(edge_id)
			edge_flag.Dist = edge_dist
			edge_flag.SDist = dist
			heap.Enqueue(PQItem{edge_id, edge_dist}, edge_dist)
			node_flag := node_flags.Get(next_node_id)
			if edge_dist < node_flag.Dist {
				node_flag.Dist = edge_dist
			}
		})
	}

	for {
		curr_item, ok := heap.Dequeue()
		if !ok {
			break
		}
		curr_id := curr_item.item
		curr_dist := curr_item.dist
		curr_flag := edge_flags.Get(curr_id)
		if curr_flag.Dist < curr_dist {
			continue
		}
		consumer.ConsumeEdge(curr_id, int(curr_flag.SDist), int(curr_flag.Dist))
		curr_edge := g.GetEdge(curr_id)
		curr_ref := graph.EdgeRef{EdgeID: curr_id, Type: 0, OtherID: curr_edge.NodeB}
		explorer.ForAdjacentEdges(curr_edge.NodeB, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			other_id := ref.EdgeID
			other_node_b := ref.OtherID
			other_flag := edge_flags.Get(other_id)
			new_length := curr_flag.Dist + explorer.GetEdgeWeight(ref) + explorer.GetTurnCost(curr_ref, curr_edge.NodeB, ref)
			if new_length > max_range {
				return
			}
			if other_flag.Dist > new_length {
				edge_flag := edge_flags.Get(other_id)
				edge_flag.Dist = new_length
				edge_flag.SDist = curr_flag.Dist + explorer.GetTurnCost(curr_ref, curr_edge.NodeB, ref)
				heap.Enqueue(PQItem{other_id, new_length}, new_length)
				node_flag := node_flags.Get(other_node_b)
				if new_length < node_flag.Dist {
					node_flag.Dist = new_length
				}
			}
		})
	}
}
