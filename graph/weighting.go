package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
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

type IWeightHandler interface {
	Load(dir string, name string) IWeighting
	Store(dir string, name string, weight IWeighting)
	Remove(dir string, name string)
	_ReorderNodes(weight IWeighting, mapping Array[int32]) // Reorder nodes of base-graph
}

var WEIGHTING_HANDLERS = Dict[WeightType, IWeightHandler]{
	DEFAULT_WEIGHT:   _DefaultWeightingHandler{},
	TURN_COST_WEIGHT: _TCWeightingHandler{},
}

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

type _DefaultWeightingHandler struct{}

func (self _DefaultWeightingHandler) Load(dir string, name string) IWeighting {
	return _LoadDefaultWeighting(dir + name + "-weight")
}
func (self _DefaultWeightingHandler) Store(dir string, name string, weight IWeighting) {
	_StoreDefaultWeighting(weight.(*DefaultWeighting), dir+name+"-weight")
}
func (self _DefaultWeightingHandler) Remove(dir string, name string) {
	os.Remove(dir + name + "-weight")
}
func (self _DefaultWeightingHandler) _ReorderNodes(weight IWeighting, mapping Array[int32]) {
}

func _StoreDefaultWeighting(weight *DefaultWeighting, filename string) {
	weightbuffer := bytes.Buffer{}

	edgecount := len(weight.edge_weights)
	binary.Write(&weightbuffer, binary.LittleEndian, int32(edgecount))
	for i := 0; i < edgecount; i++ {
		edge_weight := weight.GetEdgeWeight(int32(i))
		binary.Write(&weightbuffer, binary.LittleEndian, uint8(edge_weight))
	}

	weightfile, _ := os.Create(filename)
	defer weightfile.Close()
	weightfile.Write(weightbuffer.Bytes())
}

func _LoadDefaultWeighting(file string) *DefaultWeighting {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	nodedata, _ := os.ReadFile(file)
	nodereader := bytes.NewReader(nodedata)

	var edgecount int32
	binary.Read(nodereader, binary.LittleEndian, &edgecount)

	weights := make([]int32, edgecount)
	for i := 0; i < int(edgecount); i++ {
		var w uint8
		binary.Read(nodereader, binary.LittleEndian, &w)
		weights[i] = int32(w)
	}

	return &DefaultWeighting{
		edge_weights: weights,
	}
}

func BuildDefaultWeighting(base GraphBase) IWeighting {
	edges := base.store.edges

	weights := NewArray[int32](edges.Length())
	for id, edge := range edges {
		w := edge.Length * 3.6 / float32(edge.Maxspeed)
		if w < 1 {
			w = 1
		}
		weights[id] = int32(w)
	}

	return &DefaultWeighting{
		edge_weights: weights,
	}
}

func BuildEqualWeighting(base GraphBase) IWeighting {
	count := base.EdgeCount()

	weights := NewArray[int32](count)
	for i := 0; i < count; i++ {
		weights[i] = 1
	}

	return &DefaultWeighting{
		edge_weights: weights,
	}
}

func BuildPedestrianWeighting(base GraphBase) IWeighting {
	edges := base.store.edges

	weights := NewArray[int32](edges.Length())
	for id, edge := range edges {
		w := edge.Length * 3.6 / 3
		if w < 1 {
			w = 1
		}
		weights[id] = int32(w)
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

type _TCWeightingHandler struct{}

func (self _TCWeightingHandler) Load(dir string, name string) IWeighting {
	return _LoadTCWeighting(dir + name + "-weight")
}
func (self _TCWeightingHandler) Store(dir string, name string, weight IWeighting) {
	_StoreTCWeighting(weight.(*TCWeighting), dir+name+"-weight")
}
func (self _TCWeightingHandler) Remove(dir string, name string) {
	os.Remove(dir + name + "-weight")
}
func (self _TCWeightingHandler) _ReorderNodes(weight IWeighting, mapping Array[int32]) {
	panic("not implemented")
}

func BuildTCWeighting(base GraphBase) IWeighting {
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

func _StoreTCWeighting(weight *TCWeighting, filename string) {
	writer := NewBufferWriter()

	edgecount := len(weight.edge_weights)
	Write(writer, int32(edgecount))
	nodecount := len(weight.turn_refs)
	Write(writer, int32(nodecount))

	for i := 0; i < edgecount; i++ {
		edge_weight := weight.GetEdgeWeight(int32(i))
		Write(writer, uint8(edge_weight))
		edge_indices := weight.edge_indices[i]
		Write(writer, uint8(edge_indices.A))
		Write(writer, uint8(edge_indices.B))
	}
	for i := 0; i < nodecount; i++ {
		tc_ref := weight.turn_refs[i]
		Write(writer, int32(tc_ref.A))
		Write(writer, uint8(tc_ref.B))
		Write(writer, uint8(tc_ref.C))
	}
	WriteArray(writer, weight.turn_weights)

	weightfile, _ := os.Create(filename)
	defer weightfile.Close()
	weightfile.Write(writer.Bytes())
}

func _LoadTCWeighting(file string) *TCWeighting {
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

	return &TCWeighting{
		edge_weights: List[int32](edge_weights),
		edge_indices: List[Tuple[byte, byte]](edge_indices),
		turn_refs:    List[Triple[int, byte, byte]](turn_refs),
		turn_weights: turn_weights,
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
