package algorithm

import (
	"fmt"

	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

// Group closely connected components.
func ConnectedComponents(graph graph.IGraph) Array[int] {
	// compute closely connected components
	groups := NewArray[int](graph.NodeCount())
	group := 1
	for i := 0; i < graph.NodeCount(); i++ {
		if groups[i] != 0 {
			continue
		}
		slog.Debug(fmt.Sprintf("iteration: %v", group))
		start := int32(i)
		visited := CalcBidirectBFS(graph, start)
		for i := 0; i < graph.NodeCount(); i++ {
			if visited[i] {
				if groups[i] != 0 {
					slog.Debug("failure 1")
				}
				groups[i] = group
			}
		}
		group += 1
	}
	return groups
}
