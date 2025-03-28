package isochrone

import (
	"math"

	. "github.com/ttpr0/go-routing/util"
)

type IsoNode[T any] struct {
	Value  T
	level  int32
	x      int32 // center of the cell (index of the lower left cell of the maxlevel closest to the center)
	y      int32 // center of the cell
	cellx  int32 // index of the cell in the cell-level
	celly  int32 // index of the cell in the cell-level
	parent *IsoNode[T]
	child1 *IsoNode[T]
	child2 *IsoNode[T]
	child3 *IsoNode[T]
	child4 *IsoNode[T]
}

/*
[ child4 ] [ child3 ]
[ child1 ] [ child2 ]
*/

func _GetNode[T any](node *IsoNode[T], x, y int32, maxlevel int32) *IsoNode[T] {
	if node == nil {
		return nil
	}
	if node.level == maxlevel {
		return node
	}
	if x <= node.x && y <= node.y {
		return _GetNode(node.child1, x, y, maxlevel)
	} else if x > node.x && y <= node.y {
		return _GetNode(node.child2, x, y, maxlevel)
	} else if x > node.x && y > node.y {
		return _GetNode(node.child3, x, y, maxlevel)
	} else {
		return _GetNode(node.child4, x, y, maxlevel)
	}
}

// x and y is the index of the cell in the cell-level
func _NewIsoNode[T any](parent *IsoNode[T], cellx, celly int32, maxlevel int32) *IsoNode[T] {
	parent_level := int32(-1)
	if parent != nil {
		parent_level = parent.level
	}
	factor := int32(math.Pow(2, float64(maxlevel-(parent_level+1))))
	x := cellx
	y := celly
	if factor > 1 {
		x = cellx*factor + factor/2 - 1
		y = celly*factor + factor/2 - 1
	}
	return &IsoNode[T]{
		x:      x,
		y:      y,
		cellx:  cellx,
		celly:  celly,
		level:  parent_level + 1,
		parent: parent,
	}
}

// x and y should be the indexes of the cell (at maxlevel; starting from 0)
// value function takes the old value and should return the new value
func _InsertOrUpdateNode[T any](node *IsoNode[T], x, y int32, maxlevel int32, value func(T) T) {
	if node.level == maxlevel {
		node.Value = value(node.Value)
		return
	}
	if x <= node.x && y <= node.y {
		if node.child1 == nil {
			node.child1 = _NewIsoNode(node, 2*node.cellx, 2*node.celly, maxlevel)
		}
		_InsertOrUpdateNode(node.child1, x, y, maxlevel, value)
	} else if x > node.x && y <= node.y {
		if node.child2 == nil {
			node.child2 = _NewIsoNode(node, 2*node.cellx+1, 2*node.celly, maxlevel)
		}
		_InsertOrUpdateNode(node.child2, x, y, maxlevel, value)
	} else if x > node.x && y > node.y {
		if node.child3 == nil {
			node.child3 = _NewIsoNode(node, 2*node.cellx+1, 2*node.celly+1, maxlevel)
		}
		_InsertOrUpdateNode(node.child3, x, y, maxlevel, value)
	} else if x <= node.x && y > node.y {
		if node.child4 == nil {
			node.child4 = _NewIsoNode(node, 2*node.cellx, 2*node.celly+1, maxlevel)
		}
		_InsertOrUpdateNode(node.child4, x, y, maxlevel, value)
	}
}

func _UpdateNode[T any](node *IsoNode[T], x, y int32, maxlevel int32, value func(T) T) {
	if node.level == maxlevel {
		node.Value = value(node.Value)
		return
	}
	if x <= node.x && y <= node.y {
		if node.child1 == nil {
			return
		}
		_UpdateNode(node.child1, x, y, maxlevel, value)
	} else if x > node.x && y <= node.y {
		if node.child2 == nil {
			return
		}
		_UpdateNode(node.child2, x, y, maxlevel, value)
	} else if x > node.x && y > node.y {
		if node.child3 == nil {
			return
		}
		_UpdateNode(node.child3, x, y, maxlevel, value)
	} else if x <= node.x && y > node.y {
		if node.child4 == nil {
			return
		}
		_UpdateNode(node.child4, x, y, maxlevel, value)
	}
}

type IsoTree[T any] struct {
	root   *IsoNode[T]
	extent [4]int32 // lower and upper bound in x and y direction ([minx, miny, maxx, maxy])
	depth  int32
}

// Returns the value from the given x and y location and a bool indicating success
//
// If no value is found, false will be returned else true.
func (self *IsoTree[T]) Get(x int32, y int32) (T, bool) {
	// transform x and y to the maxlevel cell-index
	if x < self.extent[0] || x > self.extent[2] || y < self.extent[1] || y > self.extent[3] {
		var t T
		return t, false
	}
	node := _GetNode(self.root, x-self.extent[0], y-self.extent[1], self.depth)
	if node == nil {
		var t T
		return t, false
	}
	return node.Value, true
}

// Inserts a new node into the Tree.
// If a node at position x and y already exists the node will be updated with calc method.
func (self *IsoTree[T]) Insert(x, y int32, value func(T) T) {
	if x < self.extent[0] || x > self.extent[2] || y < self.extent[1] || y > self.extent[3] {
		return
	}
	cx := x - self.extent[0]
	cy := y - self.extent[1]
	if self.root == nil {
		self.root = _NewIsoNode[T](nil, 0, 0, self.depth)
	}
	_InsertOrUpdateNode(self.root, cx, cy, self.depth, value)
}

// Inserts or updates a node in the Tree.
// If a node at position x and y already exists the node will be updated with calc method.
func (self *IsoTree[T]) InsertValue(x, y int32, value T) {
	if x < self.extent[0] || x > self.extent[2] || y < self.extent[1] || y > self.extent[3] {
		return
	}
	cx := x - self.extent[0]
	cy := y - self.extent[1]
	if self.root == nil {
		self.root = _NewIsoNode[T](nil, 0, 0, self.depth)
	}
	valuefunc := func(old T) T {
		return value
	}
	_InsertOrUpdateNode(self.root, cx, cy, self.depth, valuefunc)
}

// Inserts a new node into the Tree.
// If a node at position x and y already exists the node will be updated with calc method.
func (self *IsoTree[T]) Update(x, y int32, value func(T) T) {
	if x < self.extent[0] || x > self.extent[2] || y < self.extent[1] || y > self.extent[3] {
		return
	}
	cx := x - self.extent[0]
	cy := y - self.extent[1]
	_UpdateNode(self.root, cx, cy, self.depth, value)
}

func (self *IsoTree[T]) UpdateValue(x, y int32, value T) {
	if x < self.extent[0] || x > self.extent[2] || y < self.extent[1] || y > self.extent[3] {
		return
	}
	cx := x - self.extent[0]
	cy := y - self.extent[1]
	valuefunc := func(old T) T {
		return value
	}
	_UpdateNode(self.root, cx, cy, self.depth, valuefunc)
}

func (self *IsoTree[T]) UpdateAll(value T) {
	var traverse func(tree *IsoTree[T], node *IsoNode[T])
	traverse = func(tree *IsoTree[T], node *IsoNode[T]) {
		if node == nil {
			return
		}
		if node.level == tree.depth {
			node.Value = value
			return
		}
		traverse(tree, node.child1)
		traverse(tree, node.child2)
		traverse(tree, node.child3)
		traverse(tree, node.child4)
	}
}

func (self *IsoTree[T]) Traverse() func(yield func(Triple[int32, int32, T]) bool) {
	var traverse func(tree *IsoTree[T], node *IsoNode[T], yield func(Triple[int32, int32, T]) bool) bool
	traverse = func(tree *IsoTree[T], node *IsoNode[T], yield func(Triple[int32, int32, T]) bool) bool {
		if node == nil {
			return true
		}
		if node.level == tree.depth {
			x := node.x + self.extent[0]
			y := node.y + self.extent[1]
			return yield(MakeTriple(x, y, node.Value))
		}
		contin := traverse(tree, node.child1, yield)
		if !contin {
			return false
		}
		contin = traverse(tree, node.child2, yield)
		if !contin {
			return false
		}
		contin = traverse(tree, node.child3, yield)
		if !contin {
			return false
		}
		contin = traverse(tree, node.child4, yield)
		if !contin {
			return false
		}
		return true
	}

	return func(yield func(Triple[int32, int32, T]) bool) {
		traverse(self, self.root, yield)
	}
}

// Creates and returns a new IsoTree.
//
// extent should contain lower and upper bound in x and y direction ([minx, miny, maxx, maxy])
//
// depth is estimated from the extent as the lowest power of 2 that is greater than the extent.
func NewIsoTree[T any](extent [4]int32) *IsoTree[T] {
	// calculate the depth
	dx := extent[2] - extent[0]
	dy := extent[3] - extent[1]
	d := Max(dx, dy)
	depth := int32(math.Ceil(math.Log2(float64(d))))

	return &IsoTree[T]{
		extent: extent,
		depth:  depth,
	}
}
