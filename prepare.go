package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/ttpr0/go-routing/algorithm"
	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/parser"
	. "github.com/ttpr0/go-routing/util"
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
	base, attributes := parser.ParseGraph(DATA_DIR + "/" + GRAPH_NAME + ".pbf")
	graph.Store(base, GRAPH_DIR+"/"+GRAPH_NAME+"_pre")
	attr.Store(attributes, GRAPH_DIR+"/"+GRAPH_NAME+"_pre")

	//*******************************************
	// Remove unconnected components
	//*******************************************
	// compute closely connected components
	eq_weight := graph.NewEqualWeighting()
	g := graph.BuildGraph(base, eq_weight)
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
	graph.RemoveNodes(base, remove)
	graph.Store(base, GRAPH_DIR+"/"+GRAPH_NAME)

	weight := BuildDefaultWeighting(base, attributes)
	graph.Store(weight, GRAPH_DIR+"/"+GRAPH_NAME+"-fastest")

	g = graph.BuildGraph(base, weight)

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
			create_grasp_graph(base, weight, GRAPH_NAME, GRAPH_NAME+"_grasp_"+size, "./tmp_"+size+".txt")
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
			create_isophast_graph(base, weight, GRAPH_NAME, GRAPH_NAME+"_isophast_"+size, "./tmp_"+size+".txt")
			fmt.Println("	done:", size)
			wg.Done()
		}()
	}
	wg.Wait()

	//*******************************************
	// Create CH-Graph
	//*******************************************
	fmt.Println("start creating ch-graph")
	create_ch_graph(base, weight, GRAPH_NAME, GRAPH_NAME+"_ch")
	fmt.Println("	done")

	//*******************************************
	// Create Tiled-CH-Graph
	//*******************************************
	fmt.Println("start creating tiled-ch-graphs")
	for _, s := range PARTITIONS {
		size := fmt.Sprint(s)
		wg.Add(1)
		go func() {
			create_tiled_ch_graph(base, weight, GRAPH_NAME, GRAPH_NAME+"_ch_tiled_"+size, "./tmp_"+size+".txt")
			fmt.Println("	done:", size)
			wg.Done()
		}()
	}
	wg.Wait()
}

func create_grasp_graph(base graph.IGraphBase, weight graph.IWeighting, graph_name, out_name, tiles_name string) {
	g := graph.BuildGraph(base, weight)
	tiles := graph.ReadNodeTiles(tiles_name)
	partition := graph.NewPartition(tiles)

	td := graph.PrepareOverlay(g, partition)

	mapping := graph.ComputeTileOrdering(g, partition)
	graph.ReorderNodes(td, mapping)

	graph.PrepareGRASPCellIndex(g, partition)
	// TODO
}

func create_isophast_graph(base graph.IGraphBase, weight graph.IWeighting, graph_name, out_name, tiles_name string) {
	g := graph.BuildGraph(base, weight)
	tiles := graph.ReadNodeTiles(tiles_name)
	partition := graph.NewPartition(tiles)

	graph.PrepareIsoPHAST(g, partition)
	// TODO
}

func create_ch_graph(base graph.IGraphBase, weight graph.IWeighting, graph_name, out_name string) {
	g := graph.BuildGraph(base, weight)

	cd := graph.CalcContraction6(g)

	graph.PreparePHASTIndex(g, cd)
	// TODO
}

func create_tiled_ch_graph(base graph.IGraphBase, weight graph.IWeighting, graph_name, out_name, tiles_name string) {
	g := graph.BuildGraph(base, weight)
	tiles := graph.ReadNodeTiles(tiles_name)
	partition := graph.NewPartition(tiles)

	cd := graph.CalcContraction5(g, partition)

	graph.PrepareGSPHASTIndex(g, cd, partition)
	// TODO
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
