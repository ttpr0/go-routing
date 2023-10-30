package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/algorithm"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func prepare() {
	const DATA_DIR = "./data"
	const GRAPH_DIR = "./graphs/niedersachsen/"
	const GRAPH_NAME = "niedersachsen"
	const KAHIP_EXE = "D:/Dokumente/BA/KaHIP/kaffpa"
	var PARTITIONS = []int{1000}
	//*******************************************
	// Parse graph
	//*******************************************
	base := graph.ParseGraph(DATA_DIR + "/" + GRAPH_NAME + ".pbf")
	dir := graph.NewGraphDir(GRAPH_DIR)
	dir.AddGraphBase(GRAPH_NAME+"_pre", base)
	dir.StoreGraphBase(GRAPH_NAME + "_pre")
	dir.UnloadGraphBase(GRAPH_NAME + "_pre")

	//*******************************************
	// Remove unconnected components
	//*******************************************
	// compute closely connected components
	weight := graph.BuildEqualWeighting(base)
	g := graph.BuildBaseGraph(&base, weight)
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
	graph.RemoveBaseNodes(&base, remove)
	dir.AddGraphBase(GRAPH_NAME, base)
	dir.StoreGraphBase(GRAPH_NAME)

	weight = graph.BuildDefaultWeighting(base)
	dir.AddWeighting("fastest", weight)
	dir.StoreWeighting("fastest")

	g = graph.BuildBaseGraph(&base, weight)

	//*******************************************
	// Partition with KaHIP
	//*******************************************
	// transform to metis graph
	txt := graph.GraphToMETIS(g)
	file, _ := os.Create("./" + GRAPH_NAME + "_metis.txt")
	file.Write([]byte(txt))
	file.Close()
	// run commands
	wg := sync.WaitGroup{}
	fmt.Println("start partitioning graph")
	for _, s := range PARTITIONS {
		size := fmt.Sprint(s)
		wg.Add(1)
		go func() {
			cmd := exec.Command(KAHIP_EXE, GRAPH_NAME+"_metis.txt", "--k="+size, "--preconfiguration=eco", "--output_filename=tmp_"+size+".txt")
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
			fmt.Println("	done:", size)
			wg.Done()
		}()
	}
	wg.Wait()

	//*******************************************
	// Create GRASP-Graphs
	//*******************************************
	fmt.Println("start creating grasp-graphs")
	for _, s := range PARTITIONS {
		size := fmt.Sprint(s)
		wg.Add(1)
		go func() {
			create_grasp_graph(dir, GRAPH_NAME, GRAPH_NAME+"_grasp_"+size, "./tmp_"+size+".txt")
			fmt.Println("	done:", size)
			wg.Done()
		}()
	}
	wg.Wait()

	//*******************************************
	// Create isoPHAST-Graphs
	//*******************************************
	fmt.Println("start creating isophast-graphs")
	for _, s := range PARTITIONS {
		size := fmt.Sprint(s)
		wg.Add(1)
		go func() {
			create_isophast_graph(dir, GRAPH_NAME, GRAPH_NAME+"_isophast_"+size, "./tmp_"+size+".txt")
			fmt.Println("	done:", size)
			wg.Done()
		}()
	}
	wg.Wait()

	//*******************************************
	// Create CH-Graph
	//*******************************************
	fmt.Println("start creating ch-graph")
	create_ch_graph(dir, GRAPH_NAME, GRAPH_NAME+"_ch")
	fmt.Println("	done")

	//*******************************************
	// Create Tiled-CH-Graph
	//*******************************************
	fmt.Println("start creating tiled-ch-graphs")
	for _, s := range PARTITIONS {
		size := fmt.Sprint(s)
		wg.Add(1)
		go func() {
			create_tiled_ch_graph(dir, GRAPH_NAME, GRAPH_NAME+"_ch_tiled_"+size, "./tmp_"+size+".txt")
			fmt.Println("	done:", size)
			wg.Done()
		}()
	}
	wg.Wait()
}

func create_grasp_graph(dir graph.GraphDir, graph_name, out_name, tiles_name string) {
	g := graph.BuildBaseGraph(dir.GetGraphBase(graph_name), dir.GetWeighting("fastest"))
	tiles := graph.ReadNodeTiles(tiles_name)

	td := graph.PreprocessTiledGraph3(g, tiles)

	order := graph.ComputeTileOrdering(g, tiles)
	mapping := graph.NodeOrderToNodeMapping(order)
	graph.ReorderSpeedUpNodes(td, mapping, graph.ONLY_TARGET_NODES)

	graph.PrepareGRASPCellIndex2(g, td)

	dir.AddSpeedUp(out_name, td)
	dir.StoreSpeedUp(out_name)
	dir.UnloadSpeedUp(out_name)
}

func create_isophast_graph(dir graph.GraphDir, graph_name, out_name, tiles_name string) {
	g := graph.BuildBaseGraph(dir.GetGraphBase(graph_name), dir.GetWeighting("fastest"))
	tiles := graph.ReadNodeTiles(tiles_name)

	td := graph.PreprocessTiledGraph5(g, tiles)

	dir.AddSpeedUp(out_name, td)
	dir.StoreSpeedUp(out_name)
	dir.UnloadSpeedUp(out_name)
}

func create_ch_graph(dir graph.GraphDir, graph_name, out_name string) {
	g := graph.BuildBaseGraph(dir.GetGraphBase(graph_name), dir.GetWeighting("fastest"))

	cd := graph.CalcContraction6(g)

	graph.PreparePHASTIndex2(g, cd)

	dir.AddSpeedUp(out_name, cd)
	dir.StoreSpeedUp(out_name)
	dir.UnloadSpeedUp(out_name)
}

func create_tiled_ch_graph(dir graph.GraphDir, graph_name, out_name, tiles_name string) {
	g := graph.BuildBaseGraph(dir.GetGraphBase(graph_name), dir.GetWeighting("fastest"))
	tiles := graph.ReadNodeTiles(tiles_name)

	cd := graph.CalcContraction5(g, tiles)

	graph.PrepareGSPHASTIndex2(g, cd)

	dir.AddSpeedUp(out_name, cd)
	dir.StoreSpeedUp(out_name)
	dir.UnloadSpeedUp(out_name)
}

func GetMostCommon[T comparable](arr Array[T]) T {
	var max_val T
	max_count := 0
	counts := NewDict[T, int](10)
	for i := 0; i < arr.Length(); i++ {
		val := arr[i]
		count := counts[val]
		count += 1
		if count > max_count {
			max_count = count
			max_val = val
		}
		counts[val] = count
	}
	return max_val
}
