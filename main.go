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

	GRAPH = graph.LoadTiledGraph("./data/niedersachsen")
	MANAGER = routing.NewDistributedManager(GRAPH)

	http.HandleFunc("/v0/routing", HandleRoutingRequest)
	http.HandleFunc("/v0/routing/draw/create", HandleCreateContextRequest)
	http.HandleFunc("/v0/routing/draw/step", HandleRoutingStepRequest)
	http.HandleFunc("/v0/isoraster", HandleIsoRasterRequest)
	http.ListenAndServe(":5000", nil)
}
