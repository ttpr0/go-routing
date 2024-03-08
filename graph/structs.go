package graph

//*******************************************
// edgeref struct
//*******************************************

type EdgeRef struct {
	EdgeID  int32
	Type    byte
	OtherID int32
}

func (self EdgeRef) IsEdge() bool {
	return self.Type < 100
}
func (self EdgeRef) IsCrossBorder() bool {
	return self.Type == 10
}
func (self EdgeRef) IsSkip() bool {
	return self.Type == 20
}
func (self EdgeRef) IsShortcut() bool {
	return self.Type >= 100
}
func (self EdgeRef) IsCHShortcut() bool {
	return self.Type == 100
}

func CreateEdgeRef(edge int32) EdgeRef {
	return EdgeRef{
		EdgeID:  edge,
		Type:    0,
		OtherID: -1,
	}
}
func CreateCHShortcutRef(edge int32) EdgeRef {
	return EdgeRef{
		EdgeID:  edge,
		Type:    100,
		OtherID: -1,
	}
}
