package attr

//*******************************************
// graph attributes
//*******************************************

type EdgeAttribs struct {
	Type     RoadType
	Length   float32
	Maxspeed byte
	Oneway   bool
}

type NodeAttribs struct {
	Type int8
}
