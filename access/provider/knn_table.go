package provider

type IKNNTable interface {
	GetNearest(destination int, k int) int

	GetNearestRange(destination int, k int) float32
}

type KNNTable struct {
	nearest [][]int
	ranges  [][]float32
}

func NewKNNTable(nearest [][]int, ranges [][]float32) KNNTable {
	return KNNTable{
		nearest: nearest,
		ranges:  ranges,
	}
}

func (self *KNNTable) getNearest(destination int, k int) int {
	return self.nearest[destination][k]
}

func (self *KNNTable) getNearestRange(destination int, k int) float32 {
	return self.ranges[destination][k]
}
