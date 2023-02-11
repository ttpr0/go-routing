package util

import (
	"math/rand"
	"testing"
)

func _Check[T any](node *QuadNode[T]) bool {
	if node.child1 != nil {
		if node.child1.X < node.X || node.child1.Y < node.Y {
			return false
		}
		res := _Check(node.child1)
		if res != true {
			return false
		}
	}
	if node.child2 != nil {
		if node.child2.X < node.X || node.child2.Y < node.Y {
			return false
		}
		res := _Check(node.child2)
		if res != true {
			return false
		}
	}
	if node.child3 != nil {
		if node.child3.X < node.X || node.child3.Y < node.Y {
			return false
		}
		res := _Check(node.child3)
		if res != true {
			return false
		}
	}
	if node.child4 != nil {
		if node.child4.X < node.X || node.child4.Y < node.Y {
			return false
		}
		res := _Check(node.child4)
		if res != true {
			return false
		}
	}

	return true
}

func TestRemove(t *testing.T) {
	tree := NewQuadTree(func(a, b int) int { return a })

	pairs := NewList[Tuple[int32, int32]](10)
	for i := 0; i < 100; i++ {
		x := rand.Int31n(100)
		y := rand.Int31n(100)
		tree.Insert(x, y, i)
		if i%10 == 0 {
			pairs.Add(MakeTuple(x, y))
		}
	}

	for _, pair := range pairs {
		tree.Remove(pair.A, pair.B)
	}

	if !_Check(tree.root) {
		t.Errorf("tree.Check() = false; want true")
	}
}
