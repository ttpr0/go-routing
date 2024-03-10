package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/ttpr0/go-routing/algorithm"
	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/parser"
	"github.com/ttpr0/go-routing/preproc"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

var MANAGER *RoutingManager

func main() {
	fmt.Println("hello world")

	config := ReadConfig("./config.yaml")
	MANAGER = NewRoutingManager("./graphs/test", config)

	http.HandleFunc("/v0/routing", HandleRoutingRequest)
	http.HandleFunc("/v0/routing/draw/create", HandleCreateContextRequest)
	http.HandleFunc("/v0/routing/draw/step", HandleRoutingStepRequest)
	http.HandleFunc("/v0/isoraster", HandleIsoRasterRequest)
	http.HandleFunc("/v1/matrix", HandleMatrixRequest)

	app := http.DefaultServeMux
	MapGet(app, "/test", func(none) Result {
		return OK(1)
	})
	MapPost(app, "/test1", func(a struct{ a string }) Result {
		return OK(a)
	})

	http.ListenAndServe(":5002", nil)
}

// parse OSM to graph and remove unconnected components
func main2() {
	base, attributes := parser.ParseGraph("./data/saarland.pbf")
	weight := BuildFastestWeighting(base, attributes)
	g := graph.BuildGraph(base, weight, None[comps.IGraphIndex]())

	// compute closely connected components
	groups := algorithm.ConnectedComponents(g)

	// get largest group
	max_group := GetMostCommon(groups)

	// get nodes and edges to be removed
	e := g.GetGraphExplorer()
	nodes_remove := NewArray[bool](g.NodeCount())
	edges_remove := NewArray[bool](g.EdgeCount())
	for i := 0; i < g.NodeCount(); i++ {
		if groups[i] == max_group {
			continue
		}
		// mark nodes not part of largest group
		nodes_remove[i] = true
		// mark all in- and out-going edges of those nodes
		e.ForAdjacentEdges(int32(i), graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			edges_remove[ref.EdgeID] = true
		})
		e.ForAdjacentEdges(int32(i), graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			edges_remove[ref.EdgeID] = true
		})
	}

	// get remove lists
	remove_nodes := NewList[int32](100)
	remove_edges := NewList[int32](100)
	for i := 0; i < g.NodeCount(); i++ {
		if nodes_remove[i] {
			remove_nodes.Add(int32(i))
		}
	}
	for i := 0; i < g.EdgeCount(); i++ {
		if edges_remove[i] {
			remove_edges.Add(int32(i))
		}
	}
	fmt.Println("remove", remove_nodes.Length(), "nodes")

	// remove nodes from graph
	comps.RemoveNodes(base, remove_nodes)
	attributes.RemoveNodes(remove_nodes)
	attributes.RemoveEdges(remove_edges)

	comps.Store(base, "./graphs/niedersachsen/base")
	attr.Store(attributes, "./graphs/niedersachsen/attr")
	weight = BuildFastestWeighting(base, attributes)
	comps.Store(weight, "./graphs/niedersachsen/default")
}

// create grasp graph
func main3() {
	base := comps.Load[*comps.GraphBase]("./graphs/niedersachsen/base")
	weight := comps.Load[*comps.DefaultWeighting]("./graphs/niedersachsen/default")
	g := graph.BuildGraph(base, weight, None[comps.IGraphIndex]())

	partition := comps.Load[*comps.Partition]("./graphs/niedersachsen/partition")

	td := preproc.PrepareOverlay(g, partition)
	ci := preproc.PrepareGRASPCellIndex(g, partition)

	comps.Store(td, "./graphs/niedersachsen/grasp-overlay")
	comps.Store(ci, "./graphs/niedersachsen/grasp-index")
}

// create ch_graph
func main4() {
	base := comps.Load[*comps.GraphBase]("./graphs/niedersachsen/base")
	weight := comps.Load[*comps.DefaultWeighting]("./graphs/niedersachsen/default")

	ch := preproc.CalcContraction6(base, weight)
	// ci := graph.PreparePHASTIndex(g, cd)

	// mapping := graph.ComputeLevelOrdering(g, ch)

	comps.Store(ch, "./graphs/niedersachsen/ch")
	// graph.Store(ci, "./graphs/niedersachsen/ch-index")
}

// create ch_graph with node_tile
func main5() {
	base := comps.Load[*comps.GraphBase]("./graphs/niedersachsen/base")
	weight := comps.Load[*comps.DefaultWeighting]("./graphs/niedersachsen/default")

	partition := comps.Load[*comps.Partition]("./graphs/niedersachsen/partition")

	ch := preproc.CalcContraction5(base, weight, partition)
	// ci := graph.PrepareGSPHASTIndex(g, cd, partition)

	// mapping := graph.ComputeTileLevelOrdering(g, partition, ch)

	comps.Store(ch, "./graphs/niedersachsen/ch_2")
	// graph.Store(ci, "./graphs/niedersachsen/ch-index")
}

// create isophast graph
func main6() {
	base := comps.Load[*comps.GraphBase]("./graphs/niedersachsen/base")
	weight := comps.Load[*comps.DefaultWeighting]("./graphs/niedersachsen/default")

	partition := comps.Load[*comps.Partition]("./graphs/niedersachsen/partition")

	td, ci := preproc.PrepareIsoPHAST(base, weight, partition)

	comps.Store(td, "./graphs/niedersachsen/isophast-overlay")
	comps.Store(ci, "./graphs/niedersachsen/isophast-index")
}

func main7() {
	base := comps.Load[*comps.GraphBase]("./graphs/niedersachsen/base")
	weight := comps.Load[*comps.DefaultWeighting]("./graphs/niedersachsen/default")

	ch_data := comps.Load[*comps.CH]("./graphs/niedersachsen/ch")
	ch_index := preproc.PreparePHASTIndex(base, weight, ch_data)
	cg := graph.BuildCHGraph(base, weight, None[comps.IGraphIndex](), ch_data, Some(ch_index))

	fmt.Println("finished loading graph")

	count := 0
	down_egdes, _ := cg.GetDownEdges(graph.FORWARD)
	for i := 0; i < len(down_egdes); i++ {
		edge := down_egdes[i]
		count := structs.Shortcut_get_payload[int32](&edge, 0)
		if count <= 1 {
			count += 1
		}
	}
	fmt.Println(count, "/", len(down_egdes))

	const N = 3
	const RANGE = 1000000

	nodes := NewList[int32](N)
	for i := 0; i < N; i++ {
		nodes.Add(rand.Int31n(int32(cg.NodeCount())))
	}

	t1 := time.Now()
	for _, n := range nodes {
		dist := algorithm.CalcAllDijkstra(cg, n, RANGE)
		fmt.Println(dist[100], dist[1000], dist[10000])
	}
	t2 := time.Since(t1)
	fmt.Println("1-to-all Dijkstra:", t2.Milliseconds()/N, "ms")

	// t1 = time.Now()
	// for _, n := range nodes {
	// 	algorithm.CalcBreathFirstSearch(g, n)
	// }
	// t2 = time.Since(t1)
	// fmt.Println("BFS:", t2.Milliseconds()/N, "ms")

	// t1 = time.Now()
	// for _, n := range nodes {
	// 	algorithm.CalcPHAST(g, n)
	// }
	// t2 = time.Since(t1)
	// fmt.Println("PHAST:", t2.Milliseconds()/N, "ms")

	t1 = time.Now()
	for _, n := range nodes {
		dist := algorithm.CalcPHAST2(cg, n, RANGE)
		fmt.Println(dist[100], dist[1000], dist[10000])
	}
	t2 = time.Since(t1)
	fmt.Println("PHAST2:", t2.Milliseconds()/N, "ms")

	t1 = time.Now()
	for _, n := range nodes {
		dist := algorithm.CalcPHAST3(cg, n, RANGE)
		fmt.Println(dist[100], dist[1000], dist[10000])
	}
	t2 = time.Since(t1)
	fmt.Println("PHAST3:", t2.Milliseconds()/N, "ms")

	// t1 = time.Now()
	// for _, n := range nodes {
	// 	algorithm.CalcPHAST4(cg2, n, RANGE)
	// }
	// t2 = time.Since(t1)
	// fmt.Println("PHAST4:", t2.Milliseconds()/N, "ms")

	// t1 = time.Now()
	// for _, n := range nodes {
	// 	algorithm.CalcPHAST5(cg2, n, RANGE)
	// }
	// t2 = time.Since(t1)
	// fmt.Println("PHAST5:", t2.Milliseconds()/N, "ms")
}

func main10() {
	// load location-data
	demand_locs, demand_weights, supply_locs, supply_weights := load_data("./data/population_wittmund.json", "./data/physicians_wittmund.json")

	// write demand to text file
	dem_file, err := os.Create("population_wittmund.txt")
	if err != nil {
		fmt.Println("failed to create file")
		return
	}
	defer dem_file.Close()
	var builder strings.Builder
	for i := 0; i < demand_locs.Length(); i++ {
		loc := demand_locs[i]
		weight := demand_weights[i]
		builder.WriteString(fmt.Sprintln(loc[0], loc[1], weight))
	}
	dem_file.Write([]byte(builder.String()))

	// write supply to text file
	sup_file, err := os.Create("physicians_wittmund.txt")
	if err != nil {
		fmt.Println("failed to create file")
		return
	}
	defer sup_file.Close()
	builder = strings.Builder{}
	for i := 0; i < supply_locs.Length(); i++ {
		loc := supply_locs[i]
		weight := supply_weights[i]
		builder.WriteString(fmt.Sprintln(loc[0], loc[1], weight))
	}
	sup_file.Write([]byte(builder.String()))
}

func main8() {
	base := comps.Load[*comps.GraphBase]("./graphs/niedersachsen/base")
	weight := comps.Load[*comps.DefaultWeighting]("./graphs/niedersachsen/default")

	f, err := os.Create("test.prof")
	if err != nil {
		fmt.Println("failed to create log file")
	}
	defer f.Close()

	pprof.StartCPUProfile(f)

	preproc.CalcContraction6(base, weight)

	pprof.StopCPUProfile()
}
