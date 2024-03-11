package knearest

type DistFlag struct {
	Dist   int32
	Source int32
}

func (self *DistFlag) GetDist() int32 {
	return self.Dist
}

type PQItem struct {
	item int32
	dist int32
}
