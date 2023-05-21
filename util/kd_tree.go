package util

import "math"

type KDNode[T any] struct {
	Location []float32
	Value    T

	less *KDNode[T]
	more *KDNode[T]
}

type KDTree[T any] struct {
	dim  int
	root *KDNode[T]
}

func NewKDTree[T any](dim int) KDTree[T] {
	return KDTree[T]{
		dim: dim,
	}
}

func (self *KDTree[T]) GetClosest(location []float32, max_dist float32) (T, bool) {
	if len(location) < self.dim {
		var t T
		return t, false
	}
	node, _ := _GetClosestKDNode(self.root, 0, self.dim, location, max_dist)
	if node == nil {
		var t T
		return t, false
	} else {
		return node.Value, true
	}
}

func (self *KDTree[T]) Insert(location []float32, value T) {
	if len(location) < self.dim {
		panic("invalid location dimension")
	}
	if self.root == nil {
		self.root = &KDNode[T]{Location: location, Value: value}
	} else {
		_InsertKDNode(self.root, &KDNode[T]{Location: location, Value: value}, 0, self.dim)
	}
}

func _InsertKDNode[T any](node *KDNode[T], new_node *KDNode[T], dim int, max_dim int) {
	val := node.Location[dim]
	new_val := new_node.Location[dim]
	var new_dim int
	if dim >= max_dim-1 {
		new_dim = 0
	} else {
		new_dim = dim + 1
	}

	if val >= new_val {
		if node.less == nil {
			node.less = new_node
		} else {
			_InsertKDNode(node.less, new_node, new_dim, max_dim)
		}
	} else {
		if node.more == nil {
			node.more = new_node
		} else {
			_InsertKDNode(node.more, new_node, new_dim, max_dim)
		}
	}
}

func _GetClosestKDNode[T any](node *KDNode[T], dim int, max_dim int, location []float32, max_dist float32) (*KDNode[T], float64) {
	if node == nil {
		return nil, -1
	}
	closest := node
	dist := 0.0
	for i := 0; i < max_dim; i++ {
		dist += float64((node.Location[i] - location[i]) * (node.Location[i] - location[i]))
	}
	dist = math.Sqrt(dist)

	val := node.Location[dim]
	s_val := location[dim]
	var new_dim int
	if dim >= max_dim-1 {
		new_dim = 0
	} else {
		new_dim = dim + 1
	}

	if (val - s_val) < max_dist {
		new_closest, new_dist := _GetClosestKDNode(node.more, new_dim, max_dim, location, max_dist)
		if new_closest != nil && new_dist < dist {
			closest = new_closest
			dist = new_dist
		}
	}
	if (s_val - val) < max_dist {
		new_closest, new_dist := _GetClosestKDNode(node.less, new_dim, max_dim, location, max_dist)
		if new_closest != nil && new_dist < dist {
			closest = new_closest
			dist = new_dist
		}
	}

	if dist > float64(max_dist) {
		return nil, -1
	} else {
		return closest, dist
	}
}
