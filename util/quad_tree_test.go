package util

import (
	"fmt"
	"math"
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
		if node.child2.X >= node.X || node.child2.Y < node.Y {
			return false
		}
		res := _Check(node.child2)
		if res != true {
			return false
		}
	}
	if node.child3 != nil {
		if node.child3.X >= node.X || node.child3.Y >= node.Y {
			return false
		}
		res := _Check(node.child3)
		if res != true {
			return false
		}
	}
	if node.child4 != nil {
		if node.child4.X < node.X || node.child4.Y >= node.Y {
			return false
		}
		res := _Check(node.child4)
		if res != true {
			return false
		}
	}

	return true
}

func TestInsert(t *testing.T) {
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

	if !_Check(tree.root) {
		t.Errorf("tree.Check() = false; want true")
	}
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

func TestGet(t *testing.T) {
	tree := NewQuadTree(func(a, b int) int { return a })

	var tx int32
	var ty int32
	var tvalue int
	for i := 0; i < 100; i++ {
		x := rand.Int31n(100)
		y := rand.Int31n(100)
		tree.Insert(x, y, i)
		if i == 50 {
			tx = x
			ty = y
			tvalue = i
		}
	}

	ovalue, ok := tree.Get(tx, ty)
	if !ok || ovalue != tvalue {
		t.Errorf("tree.Check() = false; want true")
	}
}

func TestQuadTreeGetCloset(t *testing.T) {
	tree := NewQuadTree(func(a, b int) int { return a })

	values := NewList[Triple[int32, int32, int]](100)
	for i := 0; i < 100; i++ {
		x := rand.Int31n(100)
		y := rand.Int31n(100)
		tree.Insert(x, y, i)
		values.Add(MakeTriple(x, y, i))
	}

	for i := 0; i < 10; i++ {
		x := rand.Int31n(100)
		y := rand.Int31n(100)
		dist := 10000000.0
		tvalue := -1
		for _, value := range values {
			new_dist := math.Sqrt(float64((value.A-x)*(value.A-x) + (value.B-y)*(value.B-y)))
			if new_dist < dist {
				tvalue = value.C
				dist = new_dist
			}
		}

		ovalue, ok := tree.GetClosest(x, y, 10)

		fmt.Println(ovalue, ok)

		if dist <= 10 {
			if !ok || ovalue != tvalue {
				t.Errorf("tree has not found a closest node, extected %d but got %d", tvalue, ovalue)
			}
		} else {
			if ok {
				t.Errorf("tree found a value but should not")
			}
		}
	}
}
