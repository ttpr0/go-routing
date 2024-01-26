package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"

	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

func run_benchmark() {
	// // load locations and weights
	// fmt.Println("Loading location data ...")
	// demand_locs, demand_weights, supply_locs, supply_weights := load_data("./data/population_wittmund.json", "./data/physicians_wittmund.json")

	// // load graphs
	// fmt.Println("Loading graphs ...")
	// dir := graph.OpenGraphDir("./graphs/niedersachsen")
	// dir.LoadGraphBase("base")
	// dir.LoadWeighting("default")
	// dir.LoadSpeedUp("ch")
	// dir.LoadSpeedUp("tiled")

	// g1 := graph.BuildBaseGraph(dir.GetGraphBase("base"), dir.GetWeighting("default"))
	// graph.PreparePHASTIndex2(g1, dir.GetSpeedUp("ch"))
	// g2 := graph.BuildCHGraph2(dir.GetGraphBase("base"), dir.GetWeighting("default"), dir.GetSpeedUp("ch"))
	// g3 := graph.BuildTiledGraph2(dir.GetGraphBase("base"), dir.GetWeighting("default"), dir.GetSpeedUp("tiled"))

	// // prepare benchmark data
	// fmt.Println("Preparing benchmark data ...")
	// const VALUE_COUNT = 11
	// const REPEAT_COUNT = 1
	// distance_decay := decay.NewLinearDecay(1800)
	// demand_view := view.NewPointView(demand_locs, demand_weights)
	// supply_views := NewDict[int, List[view.IPointView]](VALUE_COUNT)
	// counts := []int{50, 100, 150, 200, 250, 300, 350, 400, 450, 500, 550}
	// for _, count := range counts {
	// 	views := NewList[view.IPointView](REPEAT_COUNT)
	// 	for i := 0; i < REPEAT_COUNT; i++ {
	// 		locs, weights := select_random(supply_locs, supply_weights, count)
	// 		views.Add(view.NewPointView(locs, weights))
	// 	}
	// 	supply_views[count] = views
	// }

	// // run benchmarks
	// fmt.Println("Running benchmarks ...")
	// results := NewArray[Result](VALUE_COUNT)
	// for i := 0; i < VALUE_COUNT; i++ {
	// 	count := counts[i]
	// 	fmt.Println("    running value:", count)
	// 	views := supply_views[count]

	// 	t := time.Now()
	// 	for j := 0; j < REPEAT_COUNT; j++ {
	// 		view := views[j]
	// 		access.CalcDijkstra2SFCA(g1, demand_view, view, distance_decay)
	// 	}
	// 	dt := time.Since(t)
	// 	r1 := int(dt.Milliseconds() / REPEAT_COUNT)

	// 	t = time.Now()
	// 	for j := 0; j < REPEAT_COUNT; j++ {
	// 		view := views[j]
	// 		access.CalcRPHAST2SFCA(g2, demand_view, view, distance_decay)
	// 	}
	// 	dt = time.Since(t)
	// 	r2 := int(dt.Milliseconds() / REPEAT_COUNT)

	// 	t = time.Now()
	// 	for j := 0; j < REPEAT_COUNT; j++ {
	// 		view := views[j]
	// 		access.CalcRPHAST2SFCA2(g2, demand_view, view, distance_decay)
	// 	}
	// 	dt = time.Since(t)
	// 	r3 := int(dt.Milliseconds() / REPEAT_COUNT)

	// 	t = time.Now()
	// 	for j := 0; j < REPEAT_COUNT; j++ {
	// 		view := views[j]
	// 		access.CalcRPHAST2SFCA3(g2, demand_view, view, distance_decay)
	// 	}
	// 	dt = time.Since(t)
	// 	r4 := int(dt.Milliseconds() / REPEAT_COUNT)

	// 	t = time.Now()
	// 	for j := 0; j < REPEAT_COUNT; j++ {
	// 		view := views[j]
	// 		access.CalcRPHAST2SFCA4(g2, demand_view, view, distance_decay)
	// 	}
	// 	dt = time.Since(t)
	// 	r5 := int(dt.Milliseconds() / REPEAT_COUNT)

	// 	t = time.Now()
	// 	for j := 0; j < REPEAT_COUNT; j++ {
	// 		view := views[j]
	// 		access.CalcTiled2SFCA(g3, demand_view, view, distance_decay)
	// 	}
	// 	dt = time.Since(t)
	// 	r6 := int(dt.Milliseconds() / REPEAT_COUNT)

	// 	t = time.Now()
	// 	for j := 0; j < REPEAT_COUNT; j++ {
	// 		view := views[j]
	// 		access.CalcTiled2SFCA2(g3, demand_view, view, distance_decay)
	// 	}
	// 	dt = time.Since(t)
	// 	r7 := int(dt.Milliseconds() / REPEAT_COUNT)

	// 	results[i] = Result{
	// 		value: count,
	// 		times: []int{r1, r2, r3, r4, r5, r6, r7},
	// 	}
	// }

	// // write results to csv file
	// fmt.Println("Write results to file ...")
	// headers := []string{"Standortanzahl (Angebot)", "Range-Dijkstra", "RPHAST", "Range-PHAST", "Range-RPHAST", "Range-RPHAST2", "Tiled-Dijkstra", "Index-Dijkstra"}
	// write_results("./results.csv", results, headers)
}

type BenchResult struct {
	value int
	times []int
}

func load_data(demand, supply string) (Array[geo.Coord], Array[int32], Array[geo.Coord], Array[int32]) {
	file_str, _ := os.ReadFile(demand)
	demand_features := geo.FeatureCollection{}
	_ = json.Unmarshal(file_str, &demand_features)
	demand_locs, demand_weights := read_points(demand_features)

	file_str, _ = os.ReadFile(supply)
	supply_features := geo.FeatureCollection{}
	_ = json.Unmarshal(file_str, &supply_features)
	supply_locs, supply_weights := read_points(supply_features)

	return demand_locs, demand_weights, supply_locs, supply_weights
}

func read_points(col geo.FeatureCollection) (Array[geo.Coord], Array[int32]) {
	features := col.Features()

	locations := NewList[geo.Coord](len(features))
	weights := NewList[int32](len(features))

	for _, feat := range features {
		geom := feat.Geometry().(*geo.Point)
		props := feat.Properties()

		locations.Add(geom.Coordinates())
		weights.Add(int32(props["weight"].(float64)))
	}

	return Array[geo.Coord](locations), Array[int32](weights)
}

func select_random(locations Array[geo.Coord], weights Array[int32], count int) (Array[geo.Coord], Array[int32]) {
	new_locs := NewArray[geo.Coord](count)
	new_weights := NewArray[int32](count)

	length := locations.Length()
	perm := NewArray[Tuple[int, int32]](length)
	for i := 0; i < length; i++ {
		perm[i] = MakeTuple(i, rand.Int31n(int32(length)))
	}
	sort.Slice(perm, func(i, j int) bool {
		a := perm[i]
		b := perm[j]
		return a.B < b.B
	})
	for i := 0; i < count; i++ {
		index := perm[i].A
		new_locs[i] = locations[index]
		new_weights[i] = weights[index]
	}

	return new_locs, new_weights
}

func write_results(filename string, results Array[BenchResult], headers []string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("failed to create csv file")
		return
	}
	defer file.Close()

	var builder strings.Builder
	builder.WriteString(strings.Join(headers, ";") + "\n")
	for i := 0; i < results.Length(); i++ {
		res := results[i]
		tokens := make([]string, len(res.times)+1)
		tokens[0] = fmt.Sprint(res.value)
		for j := 1; j <= len(res.times); j++ {
			tokens[j] = fmt.Sprint(res.times[j-1])
		}
		builder.WriteString(strings.Join(tokens, ";") + "\n")
	}
	file.Write([]byte(builder.String()))
}
