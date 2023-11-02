package graph

import "fmt"

// checks graph topology
func CheckGraph(g IGraph) {
	explorer := g.GetGraphExplorer()
	for i := 0; i < int(g.NodeCount()); i++ {
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if ref.IsShortcut() {
				return
			}
			edge := g.GetEdge(ref.EdgeID)
			if edge.NodeA != int32(i) {
				fmt.Println("error 81")
			}
			if edge.NodeB != ref.OtherID {
				fmt.Println("error 84")
			}
		})
		explorer.ForAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if ref.IsShortcut() {
				return
			}
			edge := g.GetEdge(ref.EdgeID)
			if edge.NodeB != int32(i) {
				fmt.Println("error 95")
			}
			if edge.NodeA != ref.OtherID {
				fmt.Println("error 98")
			}
		})
	}
}

// checks topology of ch graph
func CheckCHGraph(g ICHGraph) {
	explorer := g.GetGraphExplorer()
	for i := 0; i < int(g.NodeCount()); i++ {
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if ref.IsShortcut() {
				edge := g.GetShortcut(ref.EdgeID)
				if edge.From != int32(i) {
					fmt.Println("error 1")
				}
				if edge.To != ref.OtherID {
					fmt.Println("error 2")
				}
			} else {
				edge := g.GetEdge(ref.EdgeID)
				if edge.NodeA != int32(i) {
					fmt.Println("error 3")
				}
				if edge.NodeB != ref.OtherID {
					fmt.Println("error 4")
				}
			}
		})
		explorer.ForAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if ref.IsShortcut() {
				edge := g.GetShortcut(ref.EdgeID)
				if edge.To != int32(i) {
					fmt.Println("error 5")
				}
				if edge.From != ref.OtherID {
					fmt.Println("error 6")
				}
			} else {
				edge := g.GetEdge(ref.EdgeID)
				if edge.NodeB != int32(i) {
					fmt.Println("error 7")
				}
				if edge.NodeA != ref.OtherID {
					fmt.Println("error 8")
				}
			}
		})
	}
}

// checks graph topology
func CheckTiledGraph(g ITiledGraph) {
	explorer := g.GetGraphExplorer()

	// check edges
	for i := 0; i < int(g.NodeCount()); i++ {
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if ref.IsShortcut() {
				fmt.Println("error 23")
			} else {
				edge := g.GetEdge(ref.EdgeID)
				if g.GetNodeTile(edge.NodeA) != g.GetNodeTile(edge.NodeB) && !ref.IsCrossBorder() {
					fmt.Println("error 24")
				}
			}
		})
		explorer.ForAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if ref.IsShortcut() {
				fmt.Println("error 25")
			} else {
				edge := g.GetEdge(ref.EdgeID)
				if g.GetNodeTile(edge.NodeA) != g.GetNodeTile(edge.NodeB) && !ref.IsCrossBorder() {
					fmt.Println("error 26")
				}
			}
		})
	}

	// check skip
	for i := 0; i < int(g.NodeCount()); i++ {
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_SKIP, func(ref EdgeRef) {
			if ref.IsShortcut() {
				edge := g.GetShortcut(ref.EdgeID)
				if g.GetNodeTile(edge.From) != g.GetNodeTile(edge.To) {
					fmt.Println("error 33")
				}
			} else {
				edge := g.GetEdge(ref.EdgeID)
				if g.GetNodeTile(edge.NodeA) != g.GetNodeTile(edge.NodeB) && !ref.IsCrossBorder() {
					fmt.Println("error 34")
				}
			}
		})
		explorer.ForAdjacentEdges(int32(i), BACKWARD, ADJACENT_SKIP, func(ref EdgeRef) {
			if ref.IsShortcut() {
				edge := g.GetShortcut(ref.EdgeID)
				if g.GetNodeTile(edge.From) != g.GetNodeTile(edge.To) {
					fmt.Println("error 35")
				}
			} else {
				edge := g.GetEdge(ref.EdgeID)
				if g.GetNodeTile(edge.NodeA) != g.GetNodeTile(edge.NodeB) && !ref.IsCrossBorder() {
					fmt.Println("error 36")
				}
			}
		})
	}
}
