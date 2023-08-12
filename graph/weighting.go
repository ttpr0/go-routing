package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type IWeighting interface {
	GetEdgeWeight(edge int32) int32
	GetTurnCost(from, via, to int32) int32
}

type DefaultWeighting struct {
	edge_weights []int32
}

func (self *DefaultWeighting) GetEdgeWeight(edge int32) int32 {
	return self.edge_weights[edge]
}
func (self *DefaultWeighting) GetTurnCost(from, via, to int32) int32 {
	return 0
}

// reorders nodes in weightstore,
// mapping: old id -> new id
func (self *DefaultWeighting) _ReorderNodes(mapping Array[int32]) {
	// new_weights := NewArray[int32](len(self.edge_weights))
	// for i, id := range mapping {
	// 	new_weights[id] = self.edge_weights[i]
	// }
	// self.edge_weights = new_weights
}

func _StoreDefaultWeighting(weight *DefaultWeighting, filename string) {
	weightbuffer := bytes.Buffer{}

	edgecount := len(weight.edge_weights)
	for i := 0; i < edgecount; i++ {
		edge_weight := weight.GetEdgeWeight(int32(i))
		binary.Write(&weightbuffer, binary.LittleEndian, uint8(edge_weight))
	}

	weightfile, _ := os.Create(filename)
	defer weightfile.Close()
	weightfile.Write(weightbuffer.Bytes())
}

func _LoadDefaultWeighting(file string, edgecout int) *DefaultWeighting {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	nodedata, _ := os.ReadFile(file)
	nodereader := bytes.NewReader(nodedata)
	weights := make([]int32, edgecout)
	for i := 0; i < int(edgecout); i++ {
		var w uint8
		binary.Read(nodereader, binary.LittleEndian, &w)
		weights[i] = int32(w)
	}

	return &DefaultWeighting{
		edge_weights: weights,
	}
}

type TCWeighting struct {
	edge_weights List[int32]
	edge_indices List[Tuple[byte, byte]]
	turn_refs    List[Triple[int, byte, byte]]
	turn_weights []byte
}

func (self *TCWeighting) GetEdgeWeight(edge int32) int32 {
	return self.edge_weights[edge]
}
func (self *TCWeighting) GetTurnCost(from, via, to int32) int32 {
	bwd_index := self.edge_indices[from].B
	fwd_index := self.edge_indices[to].A
	tc_ref := self.turn_refs[via]
	cols := tc_ref.C
	loc := tc_ref.A
	return int32(self.turn_weights[loc+int(cols*bwd_index)+int(fwd_index)])
}

func _CreateTCWeighting(graph *Graph) *TCWeighting {
	edge_weights := NewArray[int32](int(graph.EdgeCount()))
	edge_indices := NewArray[Tuple[byte, byte]](int(graph.EdgeCount()))
	turn_cost_ref := NewArray[Triple[int, byte, byte]](int(graph.NodeCount()))

	for i := 0; i < int(graph.EdgeCount()); i++ {
		edge := graph.GetEdge(int32(i))
		edge_weights[i] = int32(edge.Length / float32(edge.Maxspeed))
	}
	size := 0
	explorer := graph.GetDefaultExplorer()
	for i := 0; i < int(graph.NodeCount()); i++ {
		fwd_index := 0
		iter := explorer.GetAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL)
		for {
			ref, ok := iter.Next()
			if !ok {
				break
			}
			if !ref.IsEdge() {
				continue
			}
			edge_id := ref.EdgeID
			edge_indices[int(edge_id)].A = byte(fwd_index)
			fwd_index += 1
		}
		bwd_index := 0
		iter = explorer.GetAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL)
		for {
			ref, ok := iter.Next()
			if !ok {
				break
			}
			if !ref.IsEdge() {
				continue
			}
			edge_id := ref.EdgeID
			edge_indices[int(edge_id)].B = byte(bwd_index)
			bwd_index += 1
		}
		turn_cost_ref[i].B = byte(bwd_index)
		turn_cost_ref[i].C = byte(fwd_index)
		turn_cost_ref[i].A = size
		size += bwd_index * fwd_index
	}
	turn_cost_map := NewArray[byte](size)

	return &TCWeighting{
		edge_weights: List[int32](edge_weights),
		edge_indices: List[Tuple[byte, byte]](edge_indices),
		turn_refs:    List[Triple[int, byte, byte]](turn_cost_ref),
		turn_weights: turn_cost_map,
	}
}

func _StoreTCWeighting(weight *TCWeighting, filename string) {
	weightbuffer := bytes.Buffer{}

	edgecount := len(weight.edge_weights)
	for i := 0; i < edgecount; i++ {
		edge_weight := weight.GetEdgeWeight(int32(i))
		binary.Write(&weightbuffer, binary.LittleEndian, uint8(edge_weight))
		edge_indices := weight.edge_indices[i]
		binary.Write(&weightbuffer, binary.LittleEndian, uint8(edge_indices.A))
		binary.Write(&weightbuffer, binary.LittleEndian, uint8(edge_indices.B))
	}
	nodecount := len(weight.turn_refs)
	for i := 0; i < nodecount; i++ {
		tc_ref := weight.turn_refs[i]
		binary.Write(&weightbuffer, binary.LittleEndian, int32(tc_ref.A))
		binary.Write(&weightbuffer, binary.LittleEndian, uint8(tc_ref.B))
		binary.Write(&weightbuffer, binary.LittleEndian, uint8(tc_ref.C))
	}
	binary.Write(&weightbuffer, binary.LittleEndian, weight.turn_weights)

	weightfile, _ := os.Create(filename)
	defer weightfile.Close()
	weightfile.Write(weightbuffer.Bytes())
}

type TrafficWeighting struct {
	EdgeWeight []int32
	Traffic    *TrafficTable
}

func (self *TrafficWeighting) GetEdgeWeight(edge int32) int32 {
	factor := 1 + float32(self.Traffic.GetTraffic(edge))/20
	weight := float32(self.EdgeWeight[edge])
	return int32(weight * factor)
}
func (self *TrafficWeighting) GetTurnCost(from, via, to int32) int32 {
	return 0
}

type TrafficTable struct {
	EdgeTraffic []int32
}

func (self *TrafficTable) AddTraffic(edge int32) {
	self.EdgeTraffic[edge] += 1
}
func (self *TrafficTable) SubTraffic(edge int32) {
	self.EdgeTraffic[edge] -= 1
}
func (self *TrafficTable) GetTraffic(edge int32) int32 {
	return self.EdgeTraffic[edge]
}
