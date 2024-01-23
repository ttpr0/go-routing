package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// weighting interface
//*******************************************

type IWeighting interface {
	GetEdgeWeight(edge int32) int32
	GetTurnCost(from, via, to int32) int32

	Type() WeightType

	HasTurnCosts() bool
	IsDynamic() bool
	IsTimeDependant() bool
}

type WeightType byte

const (
	DEFAULT_WEIGHT   WeightType = 0
	TURN_COST_WEIGHT WeightType = 1
	TRAFFIC_WEIGHT   WeightType = 2
)

//*******************************************
// default weighting without turn costs
//*******************************************

type DefaultWeighting struct {
	edge_weights []int32
}

func (self *DefaultWeighting) GetEdgeWeight(edge int32) int32 {
	return self.edge_weights[edge]
}
func (self *DefaultWeighting) GetTurnCost(from, via, to int32) int32 {
	return 0
}

func (self *DefaultWeighting) Type() WeightType {
	return DEFAULT_WEIGHT
}
func (self *DefaultWeighting) HasTurnCosts() bool {
	return false
}
func (self *DefaultWeighting) IsDynamic() bool {
	return false
}
func (self *DefaultWeighting) IsTimeDependant() bool {
	return false
}

func (self *DefaultWeighting) _New() *DefaultWeighting {
	return &DefaultWeighting{}
}
func (self *DefaultWeighting) _Load(path string) {
	filename := path + "-weight"

	_, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + filename)
	}

	nodedata, _ := os.ReadFile(filename)
	nodereader := bytes.NewReader(nodedata)

	var edgecount int32
	binary.Read(nodereader, binary.LittleEndian, &edgecount)

	weights := make([]int32, edgecount)
	for i := 0; i < int(edgecount); i++ {
		var w uint8
		binary.Read(nodereader, binary.LittleEndian, &w)
		weights[i] = int32(w)
	}

	*self = DefaultWeighting{
		edge_weights: weights,
	}
}
func (self *DefaultWeighting) _Store(path string) {
	filename := path + "-weight"
	weightbuffer := bytes.Buffer{}

	edgecount := len(self.edge_weights)
	binary.Write(&weightbuffer, binary.LittleEndian, int32(edgecount))
	for i := 0; i < edgecount; i++ {
		edge_weight := self.GetEdgeWeight(int32(i))
		binary.Write(&weightbuffer, binary.LittleEndian, uint8(edge_weight))
	}

	weightfile, _ := os.Create(filename)
	defer weightfile.Close()
	weightfile.Write(weightbuffer.Bytes())
}
func (self *DefaultWeighting) _Remove(path string) {
	os.Remove(path + "-weight")
}
func (self *DefaultWeighting) _ReorderNodes(mapping Array[int32]) {
}

func BuildDefaultWeighting(base IGraphBase) *DefaultWeighting {
	weights := NewArray[int32](base.EdgeCount())
	for i := 0; i < base.EdgeCount(); i++ {
		edge := base.GetEdge(int32(i))
		w := edge.Length * 3.6 / float32(edge.Maxspeed)
		if w < 1 {
			w = 1
		}
		weights[i] = int32(w)
	}

	return &DefaultWeighting{
		edge_weights: weights,
	}
}

func BuildEqualWeighting(base IGraphBase) *DefaultWeighting {
	count := base.EdgeCount()

	weights := NewArray[int32](count)
	for i := 0; i < count; i++ {
		weights[i] = 1
	}

	return &DefaultWeighting{
		edge_weights: weights,
	}
}

func BuildPedestrianWeighting(base IGraphBase) *DefaultWeighting {
	weights := NewArray[int32](base.EdgeCount())
	for i := 0; i < base.EdgeCount(); i++ {
		edge := base.GetEdge(int32(i))
		w := edge.Length * 3.6 / 3
		if w < 1 {
			w = 1
		}
		weights[i] = int32(w)
	}

	return &DefaultWeighting{
		edge_weights: weights,
	}
}

//*******************************************
// weighting with turn costs
//*******************************************

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

func (self *TCWeighting) Type() WeightType {
	return TURN_COST_WEIGHT
}
func (self *TCWeighting) HasTurnCosts() bool {
	return true
}
func (self *TCWeighting) IsDynamic() bool {
	return false
}
func (self *TCWeighting) IsTimeDependant() bool {
	return false
}

func (self *TCWeighting) _New() *TCWeighting {
	return &TCWeighting{}
}
func (self *TCWeighting) _Load(path string) {
	file := path + "-weight"
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	data, _ := os.ReadFile(file)
	reader := NewBufferReader(data)

	edgecount := Read[int32](reader)
	nodecount := Read[int32](reader)

	edge_weights := NewArray[int32](int(edgecount))
	edge_indices := NewArray[Tuple[byte, byte]](int(edgecount))
	for i := 0; i < int(edgecount); i++ {
		edge_weight := Read[uint8](reader)
		edge_weights[i] = int32(edge_weight)
		ei_a := Read[uint8](reader)
		ei_b := Read[uint8](reader)
		edge_indices[i] = MakeTuple(ei_a, ei_b)
	}
	turn_refs := NewArray[Triple[int, byte, byte]](int(nodecount))
	for i := 0; i < int(nodecount); i++ {
		ref_a := Read[int32](reader)
		ref_b := Read[uint8](reader)
		ref_c := Read[uint8](reader)
		turn_refs[i] = MakeTriple(int(ref_a), ref_b, ref_c)
	}
	turn_weights := ReadArray[byte](reader)

	*self = TCWeighting{
		edge_weights: List[int32](edge_weights),
		edge_indices: List[Tuple[byte, byte]](edge_indices),
		turn_refs:    List[Triple[int, byte, byte]](turn_refs),
		turn_weights: turn_weights,
	}
}
func (self *TCWeighting) _Store(path string) {
	filename := path + "-weight"
	writer := NewBufferWriter()

	edgecount := len(self.edge_weights)
	Write(writer, int32(edgecount))
	nodecount := len(self.turn_refs)
	Write(writer, int32(nodecount))

	for i := 0; i < edgecount; i++ {
		edge_weight := self.GetEdgeWeight(int32(i))
		Write(writer, uint8(edge_weight))
		edge_indices := self.edge_indices[i]
		Write(writer, uint8(edge_indices.A))
		Write(writer, uint8(edge_indices.B))
	}
	for i := 0; i < nodecount; i++ {
		tc_ref := self.turn_refs[i]
		Write(writer, int32(tc_ref.A))
		Write(writer, uint8(tc_ref.B))
		Write(writer, uint8(tc_ref.C))
	}
	WriteArray(writer, self.turn_weights)

	weightfile, _ := os.Create(filename)
	defer weightfile.Close()
	weightfile.Write(writer.Bytes())
}
func (self *TCWeighting) _Remove(path string) {
	os.Remove(path + "-weight")
}
func (self *TCWeighting) _ReorderNodes(mapping Array[int32]) {
	panic("not implemented")
}

func BuildTCWeighting(base IGraphBase) IWeighting {
	edge_weights := NewArray[int32](int(base.EdgeCount()))
	edge_indices := NewArray[Tuple[byte, byte]](int(base.EdgeCount()))
	turn_cost_ref := NewArray[Triple[int, byte, byte]](int(base.NodeCount()))

	for i := 0; i < int(base.EdgeCount()); i++ {
		edge := base.GetEdge(int32(i))
		edge_weights[i] = int32(edge.Length / float32(edge.Maxspeed))
	}
	size := 0
	accessor := base.GetAccessor()
	for i := 0; i < int(base.NodeCount()); i++ {
		fwd_index := 0
		accessor.SetBaseNode(int32(i), FORWARD)
		for accessor.Next() {
			edge_id := accessor.GetEdgeID()
			edge_indices[int(edge_id)].A = byte(fwd_index)
			fwd_index += 1
		}
		bwd_index := 0
		accessor.SetBaseNode(int32(i), BACKWARD)
		for accessor.Next() {
			edge_id := accessor.GetEdgeID()
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

//*******************************************
// weighting with traffic updates
//*******************************************

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

func (self *TrafficWeighting) Type() WeightType {
	return TRAFFIC_WEIGHT
}
func (self *TrafficWeighting) HasTurnCosts() bool {
	return false
}
func (self *TrafficWeighting) IsDynamic() bool {
	return true
}
func (self *TrafficWeighting) IsTimeDependant() bool {
	return false
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
