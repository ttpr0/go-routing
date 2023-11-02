package decay

type LinearDecay struct {
	max_distance float32
}

func NewLinearDecay(max_distance float32) LinearDecay {
	return LinearDecay{
		max_distance: max_distance,
	}
}

func (self LinearDecay) GetDistanceWeight(distance float32) float32 {
	if distance >= self.max_distance {
		return 0
	} else {
		return 1 - (distance / self.max_distance)
	}
}

func (self LinearDecay) GetMaxDistance() float32 {
	return self.max_distance
}
