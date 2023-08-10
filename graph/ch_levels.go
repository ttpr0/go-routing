package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type CHLevelStore struct {
	node_levels Array[int16]
}

func (self *CHLevelStore) GetNodeLevel(index int32) int16 {
	return self.node_levels[index]
}
func (self *CHLevelStore) SetNodeLevel(index int32, level int16) {
	self.node_levels[index] = level
}

// reorders nodes in levelstore,
// node mapping: old id -> new id
func (self *CHLevelStore) _ReorderNodes(mapping Array[int32]) {
	new_levels := NewArray[int16](self.node_levels.Length())
	for i, id := range mapping {
		new_levels[id] = self.node_levels[i]
	}
	self.node_levels = new_levels
}

func _StoreCHLevelStore(ch_levels *CHLevelStore, filename string) {
	lvlbuffer := bytes.Buffer{}
	nodecount := ch_levels.node_levels.Length()
	for i := 0; i < nodecount; i++ {
		binary.Write(&lvlbuffer, binary.LittleEndian, ch_levels.node_levels[i])
	}

	lvlfile, _ := os.Create(filename)
	defer lvlfile.Close()
	lvlfile.Write(lvlbuffer.Bytes())
}

func _LoadCHLevelStore(file string, nodecount int) *CHLevelStore {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	leveldata, _ := os.ReadFile(file)
	levelreader := bytes.NewReader(leveldata)
	levels := NewList[int16](int(nodecount))
	for i := 0; i < int(nodecount); i++ {
		var l int16
		binary.Read(levelreader, binary.LittleEndian, &l)
		levels.Add(l)
	}

	return &CHLevelStore{
		node_levels: Array[int16](levels),
	}
}
