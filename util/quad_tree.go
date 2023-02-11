package util

type QuadNode[T any] struct {
	X      int32
	Y      int32
	Value  T
	child1 *QuadNode[T]
	child2 *QuadNode[T]
	child3 *QuadNode[T]
	child4 *QuadNode[T]
}

func _GetNode[T any](node *QuadNode[T], x, y int32) *QuadNode[T] {
	if node == nil {
		return nil
	}
	if x == node.X && y == node.Y {
		return node
	}
	if x >= node.X && y >= node.Y {
		return _GetNode(node.child1, x, y)
	} else if x < node.X && y >= node.Y {
		return _GetNode(node.child2, x, y)
	} else if x < node.X && y < node.Y {
		return _GetNode(node.child3, x, y)
	} else {
		return _GetNode(node.child4, x, y)
	}
}

func _InsertNode[T any](node *QuadNode[T], new_node *QuadNode[T]) {
	if new_node.X == node.X && new_node.Y == node.Y {
		node.Value = new_node.Value
		node.child1 = new_node.child1
		node.child2 = new_node.child2
		node.child3 = new_node.child3
		node.child4 = new_node.child4
	} else if new_node.X >= node.X && new_node.Y >= node.Y {
		if node.child1 == nil {
			node.child1 = new_node
		} else {
			_InsertNode(node.child1, new_node)
		}
	} else if new_node.X < node.X && new_node.Y >= node.Y {
		if node.child2 == nil {
			node.child2 = new_node
		} else {
			_InsertNode(node.child2, new_node)
		}
	} else if new_node.X < node.X && new_node.Y < node.Y {
		if node.child3 == nil {
			node.child3 = new_node
		} else {
			_InsertNode(node.child3, new_node)
		}
	} else if new_node.X >= node.X && new_node.Y < node.Y {
		if node.child4 == nil {
			node.child4 = new_node
		} else {
			_InsertNode(node.child4, new_node)
		}
	}
}

func _UpdateNode[T any](node *QuadNode[T], x, y int32, value T, calc func(T, T) T) {
	if x == node.X && y == node.Y {
		node.Value = calc(node.Value, value)
	} else if x >= node.X && y >= node.Y {
		if node.child1 == nil {
			return
		} else {
			_UpdateNode(node.child1, x, y, value, calc)
		}
	} else if x < node.X && y >= node.Y {
		if node.child2 == nil {
			return
		} else {
			_UpdateNode(node.child2, x, y, value, calc)
		}
	} else if x < node.X && y < node.Y {
		if node.child3 == nil {
			return
		} else {
			_UpdateNode(node.child3, x, y, value, calc)
		}
	} else if x >= node.X && y < node.Y {
		if node.child4 == nil {
			return
		} else {
			_UpdateNode(node.child4, x, y, value, calc)
		}
	}
}

func _InsertOrUpdateNode[T any](node *QuadNode[T], new_node *QuadNode[T], calc func(T, T) T) {
	if new_node.X == node.X && new_node.X == node.Y {
		node.Value = calc(node.Value, new_node.Value)
	} else if new_node.X >= node.X && new_node.Y >= node.Y {
		if node.child1 == nil {
			node.child1 = new_node
		} else {
			_InsertOrUpdateNode(node.child1, new_node, calc)
		}
	} else if new_node.X < node.X && new_node.Y >= node.Y {
		if node.child2 == nil {
			node.child2 = new_node
		} else {
			_InsertOrUpdateNode(node.child2, new_node, calc)
		}
	} else if new_node.X < node.X && new_node.Y < node.Y {
		if node.child3 == nil {
			node.child3 = new_node
		} else {
			_InsertOrUpdateNode(node.child3, new_node, calc)
		}
	} else if new_node.X >= node.X && new_node.Y < node.Y {
		if node.child4 == nil {
			node.child4 = new_node
		} else {
			_InsertOrUpdateNode(node.child4, new_node, calc)
		}
	}
}

func _RemoveNode[T any](node *QuadNode[T], x, y int32) *QuadNode[T] {
	if node == nil {
		return nil
	}
	if x >= node.X && y >= node.Y {
		node.child1 = _RemoveNode(node.child1, x, y)
		return node
	} else if x < node.X && y >= node.Y {
		node.child2 = _RemoveNode(node.child2, x, y)
		return node
	} else if x < node.X && y < node.Y {
		node.child3 = _RemoveNode(node.child3, x, y)
		return node
	} else if x >= node.X && y < node.Y {
		node.child4 = _RemoveNode(node.child4, x, y)
		return node
	}

	if node.child2 == nil && node.child3 == nil && node.child4 == nil {
		return node.child1
	} else if node.child1 == nil && node.child3 == nil && node.child4 == nil {
		return node.child2
	} else if node.child1 == nil && node.child2 == nil && node.child4 == nil {
		return node.child3
	} else if node.child1 == nil && node.child2 == nil && node.child3 == nil {
		return node.child4
	}

	if node.child1 != nil {
		parent := node
		succ := node.child1
		for succ.child3 != nil {
			parent = succ
			succ = parent.child3
		}
		parent.child3 = nil
		child1 := succ.child1
		child2 := succ.child2
		child4 := succ.child4
		succ.child1 = node.child1
		succ.child2 = node.child2
		succ.child3 = node.child3
		succ.child4 = node.child4
		_InsertNode(succ, child1)
		_InsertNode(succ, child2)
		_InsertNode(succ, child4)
		return succ
	} else if node.child2 != nil {
		parent := node
		succ := node.child2
		for succ.child4 != nil {
			parent = succ
			succ = parent.child4
		}
		parent.child4 = nil
		child1 := succ.child1
		child2 := succ.child2
		child3 := succ.child3
		succ.child1 = node.child1
		succ.child2 = node.child2
		succ.child3 = node.child3
		succ.child4 = node.child4
		_InsertNode(succ, child1)
		_InsertNode(succ, child2)
		_InsertNode(succ, child3)
		return succ
	} else if node.child3 != nil {
		parent := node
		succ := node.child3
		for succ.child1 != nil {
			parent = succ
			succ = parent.child1
		}
		parent.child1 = nil
		child2 := succ.child2
		child3 := succ.child3
		child4 := succ.child4
		succ.child1 = node.child1
		succ.child2 = node.child2
		succ.child3 = node.child3
		succ.child4 = node.child4
		_InsertNode(succ, child2)
		_InsertNode(succ, child3)
		_InsertNode(succ, child4)
		return succ
	} else {
		parent := node
		succ := node.child4
		for succ.child2 != nil {
			parent = succ
			succ = parent.child2
		}
		parent.child2 = nil
		child1 := succ.child1
		child3 := succ.child3
		child4 := succ.child4
		succ.child1 = node.child1
		succ.child2 = node.child2
		succ.child3 = node.child3
		succ.child4 = node.child4
		_InsertNode(succ, child1)
		_InsertNode(succ, child3)
		_InsertNode(succ, child4)
		return succ
	}
}

type QuadTree[T any] struct {
	root *QuadNode[T]
	calc func(T, T) T
}

func (self *QuadTree[T]) Get(x int32, y int32) (T, bool) {
	node := _GetNode(self.root, x, y)
	if node == nil {
		var t T
		return t, false
	} else {
		return node.Value, true
	}
}

func (self *QuadTree[T]) Insert(x, y int32, value T) {
	if self.root == nil {
		self.root = &QuadNode[T]{X: x, Y: y, Value: value}
	} else {
		_InsertOrUpdateNode(self.root, &QuadNode[T]{X: x, Y: y, Value: value}, self.calc)
	}
}

func (self *QuadTree[T]) Remove(x int32, y int32) {
	self.root = _RemoveNode(self.root, x, y)
}

func (self *QuadTree[T]) ToSlice() []*QuadNode[T] {
	nodes := make([]*QuadNode[T], 0, 10)
	self._Traverse(self.root, &nodes)
	return nodes
}

func (self *QuadTree[T]) _Traverse(node *QuadNode[T], nodes *[]*QuadNode[T]) {
	if node == nil {
		return
	}
	*nodes = append(*nodes, node)
	self._Traverse(node.child1, nodes)
	self._Traverse(node.child2, nodes)
	self._Traverse(node.child3, nodes)
	self._Traverse(node.child4, nodes)
}

func (self *QuadTree[T]) MergeQuadTree(tree *QuadTree[T]) {
	nodes := tree.ToSlice()
	for _, node := range nodes {
		self.Insert(node.X, node.Y, node.Value)
	}
}

func NewQuadTree[T any](calc func(T, T) T) *QuadTree[T] {
	return &QuadTree[T]{calc: calc}
}
