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

type QuadTree[T any] struct {
	root *QuadNode[T]
	calc func(T, T) T
}

func (self *QuadTree[T]) Insert(x, y int32, value T) {
	if self.root == nil {
		self.root = &QuadNode[T]{X: x, Y: y, Value: value}
	} else {
		focus := self.root
		for {
			if x == focus.X && y == focus.Y {
				focus.Value = self.calc(focus.Value, value)
				break
			}
			if x >= focus.X && y >= focus.Y {
				if focus.child1 == nil {
					focus.child1 = &QuadNode[T]{X: x, Y: y, Value: value}
					break
				} else {
					focus = focus.child1
					continue
				}
			}
			if x < focus.X && y >= focus.Y {
				if focus.child2 == nil {
					focus.child2 = &QuadNode[T]{X: x, Y: y, Value: value}
					break
				} else {
					focus = focus.child2
					continue
				}
			}
			if x < focus.X && y < focus.Y {
				if focus.child3 == nil {
					focus.child3 = &QuadNode[T]{X: x, Y: y, Value: value}
					break
				} else {
					focus = focus.child3
					continue
				}
			}
			if x >= focus.X && y < focus.Y {
				if focus.child4 == nil {
					focus.child4 = &QuadNode[T]{X: x, Y: y, Value: value}
					break
				} else {
					focus = focus.child4
					continue
				}
			}
		}
	}
}
func (self *QuadTree[T]) ToSlice() []*QuadNode[T] {
	nodes := make([]*QuadNode[T], 0, 10)
	self.traverse(self.root, &nodes)
	return nodes
}
func (self *QuadTree[T]) traverse(node *QuadNode[T], nodes *[]*QuadNode[T]) {
	if node == nil {
		return
	}
	*nodes = append(*nodes, node)
	self.traverse(node.child1, nodes)
	self.traverse(node.child2, nodes)
	self.traverse(node.child3, nodes)
	self.traverse(node.child4, nodes)
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
