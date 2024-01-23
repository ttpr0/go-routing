package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ttpr0/go-routing/algorithm"
	"github.com/ttpr0/go-routing/algorithm/partitioning"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

// var GRAPH graph.ICHGraph
// var GRAPH2 graph.ICHGraph
var GRAPH graph.IGraph
var GRAPH3 *graph.TransitGraph

// var GRAPH graph.IGraph

// var MANAGER *routing.DistributedManager

type Dummy struct {
	Val string `json:"val"`
}

func main() {
	fmt.Println("hello world")

	base := graph.Load[*graph.GraphBase]("./graphs/niedersachsen/base")
	weight := graph.Load[*graph.DefaultWeighting]("./graphs/niedersachsen/default")
	GRAPH = graph.BuildBaseGraph(base, weight)

	pedestrian_weight := graph.BuildPedestrianWeighting(base)
	transit_data := graph.LoadTransitData("./graphs/test/transit_graph")
	fmt.Println("start buidling transit-graph")
	GRAPH3 = graph.BuildTransitGraph(base, pedestrian_weight, transit_data)

	// GRAPH = graph.GetBaseGraph(dir, "default")

	// GRAPH = graph.LoadOrCreate("./graphs/default", "./data/niedersachsen.pbf", "./data/landkreise.json")
	// MANAGER = routing.NewDistributedManager(GRAPH)
	// GRAPH = graph.LoadCHGraph("./graphs/test_ch")
	// GRAPH = graph.LoadGraph("./graphs/niedersachsen_sub")
	// GRAPH = graph.LoadCHGraph("./graphs/niedersachsen_ch_3")

	http.HandleFunc("/v0/routing", HandleRoutingRequest)
	http.HandleFunc("/v0/routing/draw/create", HandleCreateContextRequest)
	http.HandleFunc("/v0/routing/draw/step", HandleRoutingStepRequest)
	http.HandleFunc("/v0/isoraster", HandleIsoRasterRequest)
	http.HandleFunc("/v0/fca", HandleFCARequest)
	http.HandleFunc("/v1/accessibility/fca", HandleFCARequest)
	http.HandleFunc("/v1/matrix", HandleMatrixRequest)

	app := http.DefaultServeMux

	MapGet(app, "/v1/test", func(v Dummy) Result2[string] {
		return OK("hello world" + v.Val)
	})
	MapGet(app, "/test", func(none) Result2[int] {
		return OK(1)
	})

	http.ListenAndServe(":5002", nil)
}

// parse OSM to graph and remove unconnected components
func main2() {
	base := graph.ParseGraph("./data/saarland.pbf")
	weight := graph.BuildDefaultWeighting(&base)
	g := graph.BuildBaseGraph(&base, weight)

	// compute closely connected components
	groups := algorithm.ConnectedComponents(g)

	// get largest group
	max_group := GetMostCommon(groups)

	// get nodes to be removed
	remove := NewList[int32](100)
	for i := 0; i < g.NodeCount(); i++ {
		if groups[i] != max_group {
			remove.Add(int32(i))
		}
	}
	fmt.Println("remove", remove.Length(), "nodes")

	// remove nodes from graph
	graph.RemoveNodes(&base, remove)

	graph.Store(&base, "./graphs/niedersachsen/base")
	weight = graph.BuildDefaultWeighting(&base)
	graph.Store(weight, "./graphs/niedersachsen/default")
}

// create grasp graph
func main3() {
	base := graph.Load[*graph.GraphBase]("./graphs/niedersachsen/base")
	weight := graph.Load[*graph.DefaultWeighting]("./graphs/niedersachsen/default")
	g := graph.BuildBaseGraph(base, weight)

	partition := graph.Load[*graph.Partition]("./graphs/niedersachsen/partition")

	td := graph.PrepareOverlay(g, partition)
	ci := graph.PrepareGRASPCellIndex(g, partition)

	graph.Store(td, "./graphs/niedersachsen/grasp-overlay")
	graph.Store(ci, "./graphs/niedersachsen/grasp-index")
}

// create ch_graph
func main4() {
	base := graph.Load[*graph.GraphBase]("./graphs/niedersachsen/base")
	weight := graph.Load[*graph.DefaultWeighting]("./graphs/niedersachsen/default")
	g := graph.BuildBaseGraph(base, weight)

	ch := graph.CalcContraction6(g)
	// ci := graph.PreparePHASTIndex(g, cd)

	// mapping := graph.ComputeLevelOrdering(g, ch)

	graph.Store(ch, "./graphs/niedersachsen/ch")
	// graph.Store(ci, "./graphs/niedersachsen/ch-index")
}

// create ch_graph with node_tile
func main5() {
	base := graph.Load[*graph.GraphBase]("./graphs/niedersachsen/base")
	weight := graph.Load[*graph.DefaultWeighting]("./graphs/niedersachsen/default")
	g := graph.BuildBaseGraph(base, weight)

	partition := graph.Load[*graph.Partition]("./graphs/niedersachsen/partition")

	ch := graph.CalcContraction5(g, partition)
	// ci := graph.PrepareGSPHASTIndex(g, cd, partition)

	// mapping := graph.ComputeTileLevelOrdering(g, partition, ch)

	graph.Store(ch, "./graphs/niedersachsen/ch")
	// graph.Store(ci, "./graphs/niedersachsen/ch-index")
}

// create isophast graph
func main6() {
	base := graph.Load[*graph.GraphBase]("./graphs/niedersachsen/base")
	weight := graph.Load[*graph.DefaultWeighting]("./graphs/niedersachsen/default")
	g := graph.BuildBaseGraph(base, weight)

	partition := graph.Load[*graph.Partition]("./graphs/niedersachsen/partition")

	td, ci := graph.PrepareIsoPHAST(g, partition)

	graph.Store(td, "./graphs/niedersachsen/grasp-overlay")
	graph.Store(ci, "./graphs/niedersachsen/grasp-index")
}

func main7() {
	base := graph.Load[*graph.GraphBase]("./graphs/niedersachsen/base")
	weight := graph.Load[*graph.DefaultWeighting]("./graphs/niedersachsen/default")
	g := graph.BuildBaseGraph(base, weight)

	ch_data := graph.Load[*graph.CH]("./graphs/niedersachsen/ch")
	ch_index := graph.PreparePHASTIndex(g, ch_data)
	cg := graph.BuildCHGraph(base, weight, ch_data, Some(ch_index))

	fmt.Println("finished loading graph")

	count := 0
	down_egdes, _ := cg.GetDownEdges(graph.FORWARD)
	for i := 0; i < len(down_egdes); i++ {
		edge := down_egdes[i]
		count := graph.Shortcut_get_payload[int32](&edge, 0)
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

func main8() {
	// g := graph.LoadGraph("./graphs/saarland")

	// // file_str, _ := os.ReadFile("./data/landkreise.json")
	// // collection := geo.FeatureCollection{}
	// // _ = json.Unmarshal(file_str, &collection)
	// // tiles := graph.CalcNodeTiles(g.GetGeometry(), collection.Features())

	// tiles := partitioning.InertialFlow(g)

	// nodes, edges := graph.GraphToGeoJSON2(g, tiles)
	// node_bytes, _ := json.Marshal(&nodes)
	// edge_bytes, _ := json.Marshal(&edges)

	// node_file, _ := os.Create("./nodes.json")
	// defer node_file.Close()
	// node_file.Write(node_bytes)

	// edge_file, _ := os.Create("./edges.json")
	// defer edge_file.Close()
	// edge_file.Write(edge_bytes)
}

func main9() {
	prepare()
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

func main11() {
	// tg := graph.LoadGraph("./graphs/saarland_sub")

	// nodes, edges := graph.GraphToGeoJSON2(tg, NewArray[int16](tg.NodeCount()))
	// node_bytes, _ := json.Marshal(&nodes)
	// edge_bytes, _ := json.Marshal(&edges)

	// node_file, _ := os.Create("./nodes.json")
	// defer node_file.Close()
	// node_file.Write(node_bytes)

	// edge_file, _ := os.Create("./edges.json")
	// defer edge_file.Close()
	// edge_file.Write(edge_bytes)
}

func main12() {
	base := graph.Load[*graph.GraphBase]("./graphs/niedersachsen/base")
	weight := graph.Load[*graph.DefaultWeighting]("./graphs/niedersachsen/default")
	g := graph.BuildBaseGraph(base, weight)

	tiles := partitioning.InertialFlow(g)

	partition := graph.NewPartition(tiles)
	graph.Store(partition, "./graphs/niedersachsen/partition")

	graph.StoreNodeTiles("./ni_tiles.txt", tiles)
}
