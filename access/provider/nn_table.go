package provider

type INNTable interface {
	GetNearest(destination int) int

	GetNearestRange(destination int) float32
}

type NNTable struct {
	nearest []int
	ranges  []float32
}

func NewNNTable(nearest []int, ranges []float32) NNTable {
	return NNTable{
		nearest: nearest,
		ranges:  ranges,
	}
}

func (self *NNTable) getNearest(destination int) int {
	return self.nearest[destination]
}

func (self *NNTable) getNearestRange(destination int) float32 {
	return self.ranges[destination]
}
