package comps

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
}

type ITCWeighting interface {
	GetEdgeWeight(edge int32) int32
	GetTurnCost(from, via, to int32) int32
}

type ITransitWeighting interface {
	GetNextWeight(connection int32, from int32) Optional[ConnectionWeight]
	GetWeightsInRange(connection int32, from, to int32) []ConnectionWeight
}

//*******************************************
// default weighting without turn costs
//*******************************************

type DefaultWeighting struct {
	edge_weights []int32
}

func NewDefaultWeighting(base IGraphBase) *DefaultWeighting {
	return &DefaultWeighting{
		edge_weights: make([]int32, base.EdgeCount()),
	}
}

func (self *DefaultWeighting) GetEdgeWeight(edge int32) int32 {
	return self.edge_weights[edge]
}
func (self *DefaultWeighting) SetEdgeWeight(edge int32, weight int32) {
	self.edge_weights[edge] = weight
}
func (self *DefaultWeighting) GetTurnCost(from, via, to int32) int32 {
	return 0
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

//*******************************************
// equal weighting
//*******************************************

type EqualWeighting struct{}

func NewEqualWeighting() *EqualWeighting {
	return &EqualWeighting{}
}

func (self *EqualWeighting) GetEdgeWeight(edge int32) int32 {
	return 1
}
func (self *EqualWeighting) GetTurnCost(from, via, to int32) int32 {
	return 0
}

//*******************************************
// dynamic weighting
//*******************************************

type DynamicWeighting struct {
	weight_func func(int32) int32
}

func NewDynamicWeighting(f func(int32) int32) *DynamicWeighting {
	return &DynamicWeighting{
		weight_func: f,
	}
}

func (self *DynamicWeighting) GetEdgeWeight(edge int32) int32 {
	return self.weight_func(edge)
}
func (self *DynamicWeighting) GetTurnCost(from, via, to int32) int32 {
	return 0
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

func NewTCWeighting(base IGraphBase) *TCWeighting {
	edge_weights := NewArray[int32](int(base.EdgeCount()))
	edge_indices := NewArray[Tuple[byte, byte]](int(base.EdgeCount()))
	turn_cost_ref := NewArray[Triple[int, byte, byte]](int(base.NodeCount()))

	size := 0
	accessor := base.GetAccessor()
	for i := 0; i < int(base.NodeCount()); i++ {
		fwd_index := 0
		accessor.SetBaseNode(int32(i), true)
		for accessor.Next() {
			edge_id := accessor.GetEdgeID()
			edge_indices[int(edge_id)].A = byte(fwd_index)
			fwd_index += 1
		}
		bwd_index := 0
		accessor.SetBaseNode(int32(i), false)
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

func (self *TCWeighting) GetEdgeWeight(edge int32) int32 {
	return self.edge_weights[edge]
}
func (self *TCWeighting) SetEdgeWeight(edge int32, weight int32) {
	self.edge_weights[edge] = weight
}
func (self *TCWeighting) GetTurnCost(from, via, to int32) int32 {
	bwd_index := self.edge_indices[from].B
	fwd_index := self.edge_indices[to].A
	tc_ref := self.turn_refs[via]
	cols := tc_ref.C
	loc := tc_ref.A
	return int32(self.turn_weights[loc+int(cols*bwd_index)+int(fwd_index)])
}
func (self *TCWeighting) SetTurnCost(from, via, to int32, weight int32) {
	bwd_index := self.edge_indices[from].B
	fwd_index := self.edge_indices[to].A
	tc_ref := self.turn_refs[via]
	cols := tc_ref.C
	loc := tc_ref.A
	self.turn_weights[loc+int(cols*bwd_index)+int(fwd_index)] = byte(weight)
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

//*******************************************
// weighting with turn costs
//*******************************************

type TransitWeighting struct {
	transit_weights Array[List[ConnectionWeight]]
}

type ConnectionWeight struct {
	Departure int32
	Arrival   int32
	Trip      int32
}

func NewTransitWeighting(transit *Transit) *TransitWeighting {
	transit_weights := NewArray[List[ConnectionWeight]](transit.ConnectionCount())

	return &TransitWeighting{
		transit_weights: transit_weights,
	}
}

func (self *TransitWeighting) GetNextWeight(connection int32, from int32) Optional[ConnectionWeight] {
	conn_weights := self.transit_weights[connection]
	for i := 0; i < conn_weights.Length(); i++ {
		if conn_weights[i].Departure >= from {
			return Some(conn_weights[i])
		}
	}
	return None[ConnectionWeight]()
}
func (self *TransitWeighting) GetWeightsInRange(connection int32, from, to int32) []ConnectionWeight {
	conn_weights := self.transit_weights[connection]
	start := -1
	end := -1
	for i := 0; i < conn_weights.Length(); i++ {
		if conn_weights[i].Departure >= from && start == -1 {
			start = i
		}
		if start != -1 && conn_weights[i].Departure > to {
			end = i
			break
		}
	}
	if start == -1 {
		return nil
	}
	if end == -1 {
		end = conn_weights.Length()
	}
	return conn_weights[start:end]
}
func (self *TransitWeighting) SetWeights(connection int32, schedule []ConnectionWeight) {
	self.transit_weights[connection] = List[ConnectionWeight](schedule)
}

func (self *TransitWeighting) _New() *TransitWeighting {
	return &TransitWeighting{}
}
func (self *TransitWeighting) _Load(path string) {
	file := path + "-weight"
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	data, _ := os.ReadFile(file)
	reader := NewBufferReader(data)

	conn_count := Read[int32](reader)
	transit_weights := NewArray[List[ConnectionWeight]](int(conn_count))

	for i := 0; i < int(conn_count); i++ {
		schedule_count := Read[int32](reader)
		schedule := NewList[ConnectionWeight](int(schedule_count))
		for j := 0; j < int(schedule_count); j++ {
			departure := Read[int32](reader)
			arrival := Read[int32](reader)
			trip := Read[int32](reader)
			schedule.Add(ConnectionWeight{departure, arrival, trip})
		}
		transit_weights[i] = schedule
	}

	*self = TransitWeighting{
		transit_weights: transit_weights,
	}
}
func (self *TransitWeighting) _Store(path string) {
	filename := path + "-weight"
	writer := NewBufferWriter()

	conn_count := self.transit_weights.Length()
	Write(writer, int32(conn_count))
	for i := 0; i < conn_count; i++ {
		conn_weights := self.transit_weights[i]
		schedule_count := conn_weights.Length()
		Write(writer, int32(schedule_count))
		for j := 0; j < schedule_count; j++ {
			conn_weight := conn_weights[j]
			Write(writer, int32(conn_weight.Departure))
			Write(writer, int32(conn_weight.Arrival))
			Write(writer, int32(conn_weight.Trip))
		}
	}

	weightfile, _ := os.Create(filename)
	defer weightfile.Close()
	weightfile.Write(writer.Bytes())
}
func (self *TransitWeighting) _Remove(path string) {
	os.Remove(path + "-weight")
}
func (self *TransitWeighting) _ReorderNodes(mapping Array[int32]) {
	panic("not implemented")
}
