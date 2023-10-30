package main

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func prepare() {
	// const DATA_DIR = "./data"
	// const GRAPH_DIR = "./graphs"
	// const GRAPH_NAME = "niedersachsen"
	// const KAHIP_EXE = "D:/Dokumente/BA/KaHIP/kaffpa"
	// var PARTITIONS = []int{1000}
	// //*******************************************
	// // Parse graph
	// //*******************************************
	// g := graph.ParseGraph(DATA_DIR + "/" + GRAPH_NAME + ".pbf")
	// graph.StoreGraph(g, GRAPH_DIR+"/"+GRAPH_NAME+"_pre")

	// //*******************************************
	// // Remove unconnected components
	// //*******************************************
	// // compute closely connected components
	// groups := algorithm.ConnectedComponents(g)
	// // get largest group
	// max_group := GetMostCommon(groups)
	// // get nodes to be removed
	// remove := NewList[int32](100)
	// for i := 0; i < g.NodeCount(); i++ {
	// 	if groups[i] != max_group {
	// 		remove.Add(int32(i))
	// 	}
	// }
	// fmt.Println("remove", remove.Length(), "nodes")
	// // remove nodes from graph
	// g = graph.RemoveNodes(g, remove)
	// graph.StoreGraph(g, GRAPH_DIR+"/"+GRAPH_NAME)

	// //*******************************************
	// // Partition with KaHIP
	// //*******************************************
	// // transform to metis graph
	// txt := graph.GraphToMETIS(g)
	// file, _ := os.Create("./" + GRAPH_NAME + "_metis.txt")
	// file.Write([]byte(txt))
	// file.Close()
	// // run commands
	// wg := sync.WaitGroup{}
	// fmt.Println("start partitioning graph")
	// for _, s := range PARTITIONS {
	// 	size := fmt.Sprint(s)
	// 	wg.Add(1)
	// 	go func() {
	// 		cmd := exec.Command(KAHIP_EXE, GRAPH_NAME+"_metis.txt", "--k="+size, "--preconfiguration=eco", "--output_filename=tmp_"+size+".txt")
	// 		if err := cmd.Run(); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		fmt.Println("	done:", size)
	// 		wg.Done()
	// 	}()
	// }
	// wg.Wait()

	// //*******************************************
	// // Create GRASP-Graphs
	// //*******************************************
	// fmt.Println("start creating grasp-graphs")
	// for _, s := range PARTITIONS {
	// 	size := fmt.Sprint(s)
	// 	wg.Add(1)
	// 	go func() {
	// 		create_grasp_graph(GRAPH_DIR+"/"+GRAPH_NAME, GRAPH_DIR+"/"+GRAPH_NAME+"_grasp_"+size, "./tmp_"+size+".txt")
	// 		fmt.Println("	done:", size)
	// 		wg.Done()
	// 	}()
	// }
	// wg.Wait()

	// //*******************************************
	// // Create isoPHAST-Graphs
	// //*******************************************
	// fmt.Println("start creating isophast-graphs")
	// for _, s := range PARTITIONS {
	// 	size := fmt.Sprint(s)
	// 	wg.Add(1)
	// 	go func() {
	// 		create_isophast_graph(GRAPH_DIR+"/"+GRAPH_NAME, GRAPH_DIR+"/"+GRAPH_NAME+"_isophast_"+size, "./tmp_"+size+".txt")
	// 		fmt.Println("	done:", size)
	// 		wg.Done()
	// 	}()
	// }
	// wg.Wait()

	// //*******************************************
	// // Create CH-Graph
	// //*******************************************
	// fmt.Println("start creating ch-graph")
	// create_ch_graph(GRAPH_DIR+"/"+GRAPH_NAME, GRAPH_DIR+"/"+GRAPH_NAME+"_ch")
	// fmt.Println("	done")

	// //*******************************************
	// // Create Tiled-CH-Graph
	// //*******************************************
	// fmt.Println("start creating tiled-ch-graphs")
	// for _, s := range PARTITIONS {
	// 	size := fmt.Sprint(s)
	// 	wg.Add(1)
	// 	go func() {
	// 		create_tiled_ch_graph(GRAPH_DIR+"/"+GRAPH_NAME, GRAPH_DIR+"/"+GRAPH_NAME+"_ch_tiled_"+size, "./tmp_"+size+".txt")
	// 		fmt.Println("	done:", size)
	// 		wg.Done()
	// 	}()
	// }
	// wg.Wait()
}

func create_grasp_graph(in_name, out_name, tiles_name string) {
	// g := graph.LoadGraph2(in_name)
	// tiles := graph.ReadNodeTiles(tiles_name)

	// tg := graph.PreprocessTiledGraph3(g, tiles)

	// // order := graph.ComputeTileOrdering(tg)
	// // mapping := graph.NodeOrderToNodeMapping(order)
	// // graph.ReorderTiledGraph(tg, mapping)

	// graph.PrepareGRASPCellIndex(tg)
	// graph.StoreTiledGraph(tg, out_name)
}

func create_isophast_graph(in_name, out_name, tiles_name string) {
	// g := graph.LoadGraph2(in_name)
	// tiles := graph.ReadNodeTiles(tiles_name)

	// tg := graph.PreprocessTiledGraph4(g, tiles)

	// graph.StoreTiledGraph(tg, out_name)
}

func create_ch_graph(in_name, out_name string) {
	// g := graph.LoadGraph2(in_name)

	// dg := graph.TransformToCHPreprocGraph(g)

	// graph.CalcContraction6(dg)

	// cg := graph.TransformFromCHPreprocGraph(dg)

	// // order := graph.ComputeLevelOrdering(cg)
	// // mapping := graph.NodeOrderToNodeMapping(order)
	// // graph.ReorderCHGraph(cg, mapping)
	// graph.SortNodesByLevel(cg)

	// graph.StoreCHGraph(cg, out_name)
}

func create_tiled_ch_graph(in_name, out_name, tiles_name string) {
	// g := graph.LoadGraph2(in_name)
	// tiles := graph.ReadNodeTiles(tiles_name)

	// dg := graph.TransformToCHPreprocGraph(g)

	// graph.CalcContraction5(dg, tiles)

	// cg := graph.TransformFromCHPreprocGraph(dg)

	// graph.PrepareGSPHASTIndex(cg)

	// graph.StoreCHGraph(cg, out_name)
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
