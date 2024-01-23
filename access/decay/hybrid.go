package decay

import (
	"sort"

	. "github.com/ttpr0/go-routing/util"
)

type HybridDecay struct {
	distances []float32
	factors   []float32
}

func NewHybridDecay(distances []float32, factors []float32) HybridDecay {
	// sort two arrays using Pair helper array
	// sorted based on distances (lowest first)
	pairs := NewArray[Tuple[float32, float32]](len(distances))
	for i := 0; i < len(distances); i++ {
		pairs[i] = MakeTuple(distances[i], factors[i])
	}
	sort.Slice(pairs, func(i, j int) bool {
		item_i := pairs[i]
		item_j := pairs[j]
		return item_i.A < item_j.A
	})
	for i := 0; i < len(pairs); i++ {
		distances[i] = pairs[i].A
		factors[i] = pairs[i].B
	}

	return HybridDecay{
		distances: distances,
		factors:   factors,
	}
}

func (self HybridDecay) GetDistanceWeight(distance float32) float32 {
	for i := 0; i < len(self.distances); i++ {
		if distance <= self.distances[i] {
			return self.factors[i]
		}
	}
	return 0
}

func (self HybridDecay) GetMaxDistance() float32 {
	return self.distances[len(self.distances)-1]
}
