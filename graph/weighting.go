package graph

type IWeighting interface {
	GetEdgeWeight(edge int32) int32
	GetTurnCost(from, via, to int32) int32
}

type Weighting struct {
	EdgeWeight []int32
}

func (self *Weighting) GetEdgeWeight(edge int32) int32 {
	return self.EdgeWeight[edge]
}
func (self *Weighting) GetTurnCost(from, via, to int32) int32 {
	return 0
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
