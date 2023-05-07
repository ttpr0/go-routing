package main

import (
	"fmt"
	"net/http"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/routing"
)

var GRAPH graph.ITiledGraph
var MANAGER *routing.DistributedManager

func main() {
	fmt.Println("hello world")

	GRAPH = graph.LoadOrCreate("./data/niedersachsen.pbf", "./data/landkreise.json", "./data")
	MANAGER = routing.NewDistributedManager(GRAPH)

	http.HandleFunc("/v0/routing", HandleRoutingRequest)
	http.HandleFunc("/v0/routing/draw/create", HandleCreateContextRequest)
	http.HandleFunc("/v0/routing/draw/step", HandleRoutingStepRequest)
	http.HandleFunc("/v0/isoraster", HandleIsoRasterRequest)
	http.HandleFunc("/v0/fca", HandleFCARequest)
}
