package algorithm

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type TransitItem struct {
	time      int32
	departure int32
	stop      int32
}

type TransitFlag struct {
	trips List[Tuple[int32, int32]]
}

// computes one-to-many distances using forward-dijkstra and public-transit
func CalcTransitDijkstra(g *graph.TransitGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], edge_flags Flags[DistFlag], stop_flags Flags[TransitFlag], max_range int32, from, to int32) {
	// step 1: range-dijkstra from start
	CalcRangeDijkstraTC(g, starts, node_flags, edge_flags, max_range)

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
				heap.Enqueue(TransitItem{new_time, item.departure, other_id}, new_time)
			} else {
				weight := explorer.GetConnectionWeight(ref, item.time)
				if !weight.HasValue() {
					return
				}
				new_time := weight.Value.Arrival
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
		base_node := g.MapStopToNode(int32(i))
		starts_.Add(MakeTuple(base_node, dist))
	}
	CalcRangeDijkstraTC(g, Array[Tuple[int32, int32]](starts_), node_flags, edge_flags, max_range)
}
