package routing

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type DD_RoutingRequest struct {
	key         int
	start_id    int32
	end_id      int32
	begin_id    int32
	path_length float64
	prev_edge   int32
}

type DistributedRoutingRunner struct {
	routing_chan  Queue[DD_RoutingRequest]
	retrivel_chan chan int32
	stop_chan     chan bool
	exit_chan     chan bool
	path_chan     chan int32
	manager       *DistributedRoutingManager

	key      int
	tile_id  int16
	heap     PriorityQueue[int32, float64]
	start_id int32
	end_id   int32
	graph    graph.ITiledGraph
	geom     graph.IGeometry
	weight   graph.IWeighting
	flags    Dict[int32, flag_d]
	finished bool
}

func NewDistributedRunner(key int, manager *DistributedRoutingManager, path_chan chan int32, tile_id int16, graph graph.ITiledGraph, start, end int32, begin int32, path_length float64, prev_edge int32) *DistributedRoutingRunner {
	d := DistributedRoutingRunner{
		routing_chan:  NewQueue[DD_RoutingRequest](),
		retrivel_chan: make(chan int32),
		stop_chan:     make(chan bool),
		exit_chan:     make(chan bool),
		path_chan:     path_chan,
		manager:       manager,

		key:      key,
		tile_id:  tile_id,
		start_id: start,
		end_id:   end,
		graph:    graph,
		geom:     graph.GetGeometry(),
		weight:   graph.GetWeighting(),
		finished: false,
	}

	flags := NewDict[int32, flag_d](100)
	flags[begin] = flag_d{path_length: path_length, prev_edge: prev_edge, visited: false}
	d.flags = flags

	heap := NewPriorityQueue[int32, float64](100)
	heap.Enqueue(begin, 0)
	d.heap = heap

	go d.HandleRetrivalRequest()
	go d.HandleStopRequest()
	go d.HandleExitRequest()

	return &d
}

func (self *DistributedRoutingRunner) HandleRetrivalRequest() {
	for {
		curr_id, ok := <-self.retrivel_chan
		if !ok {
			return
		}
		var edge int32
		for {
			if curr_id == self.start_id {
				close(self.path_chan)
				break
			}
			edge = self.flags[curr_id].prev_edge
			self.path_chan <- edge
			curr_id, _ = self.graph.GetOtherNode(edge, curr_id)
			if self.graph.GetNodeTile(curr_id) != self.tile_id {
				self.manager.retrivel_chan <- MakeTuple(self.key, curr_id)
				break
			}
		}
	}
}
func (self *DistributedRoutingRunner) HandleStopRequest() {
	<-self.stop_chan
	if self.finished {
		self.retrivel_chan <- self.end_id
	}
	self.finished = true
}
func (self *DistributedRoutingRunner) HandleExitRequest() {
	<-self.exit_chan
	self.finished = true
	close(self.retrivel_chan)
}
func (self *DistributedRoutingRunner) RunRouting() {
	for !self.finished {
		for self.routing_chan.Size() > 0 {
			request, _ := self.routing_chan.Pop()
			var flag flag_d
			if self.flags.ContainsKey(request.begin_id) {
				flag = self.flags[request.begin_id]
			} else {
				flag = flag_d{path_length: 1000000, visited: false, prev_edge: -1}
			}
			if flag.path_length > request.path_length {
				flag.prev_edge = request.prev_edge
				flag.path_length = request.path_length
				flag.visited = false
				self.heap.Enqueue(request.begin_id, request.path_length)
			}
			self.flags[request.begin_id] = flag
		}

		curr_id, ok := self.heap.Dequeue()
		if !ok {
			time.Sleep(1000)
			continue
		}
		if curr_id == self.end_id {
			self.finished = true
			self.manager.stop_chan <- self.key
			break
		}
		curr_flag := self.flags[curr_id]
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		edges := self.graph.GetAdjacentEdges(curr_id)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if ref.IsReversed() {
				continue
			}
			edge_id := ref.EdgeID
			other_id, _ := self.graph.GetOtherNode(edge_id, curr_id)
			var other_flag flag_d
			if self.flags.ContainsKey(other_id) {
				other_flag = self.flags[other_id]
			} else {
				other_flag = flag_d{path_length: 1000000, visited: false, prev_edge: -1}
			}
			new_length := curr_flag.path_length + float64(self.weight.GetEdgeWeight(edge_id))
			if ref.IsCrossBorder() {
				request := DD_RoutingRequest{
					key:         self.key,
					start_id:    self.start_id,
					end_id:      self.end_id,
					begin_id:    other_id,
					path_length: new_length,
					prev_edge:   edge_id,
				}
				self.manager.routing_chan <- request
				continue
			}
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				other_flag.visited = false
				self.heap.Enqueue(other_id, new_length)
			}
			self.flags[other_id] = other_flag
		}
		self.flags[curr_id] = curr_flag
	}
}

type DistributedRoutingHandler struct {
	tile_id int16
	runners Dict[int, *DistributedRoutingRunner]
	manager *DistributedRoutingManager
	graph   graph.ITiledGraph

	routing_chan  chan DD_RoutingRequest
	retrivel_chan chan Tuple[int, int32]
	stop_chan     chan int
	exit_chan     chan int
}

func NewDistributedHandler(tile_id int16, manager *DistributedRoutingManager, graph graph.ITiledGraph) *DistributedRoutingHandler {
	routing_chan := make(chan DD_RoutingRequest, 100)
	retrivel_chan := make(chan Tuple[int, int32], 100)
	stop_chan := make(chan int, 10)
	exit_chan := make(chan int, 10)

	runners := NewDict[int, *DistributedRoutingRunner](10)

	h := &DistributedRoutingHandler{
		tile_id: tile_id,
		runners: runners,
		manager: manager,
		graph:   graph,

		routing_chan:  routing_chan,
		retrivel_chan: retrivel_chan,
		stop_chan:     stop_chan,
		exit_chan:     exit_chan,
	}

	go h.HandleRoutingRequest()
	go h.HandleRetrivelRequest()
	go h.HandleStopRequest()
	go h.HandleExitRequest()

	return h
}

func (self *DistributedRoutingHandler) HandleRoutingRequest() {
	for {
		request := <-self.routing_chan
		if self.runners.ContainsKey(request.key) {
			runner := self.runners[request.key]
			runner.routing_chan.Push(request)
		} else if self.manager.stoped.ContainsKey(request.key) {
			continue
		} else {
			runner := NewDistributedRunner(request.key, self.manager, self.manager.paths.Get(request.key), self.tile_id, self.graph, request.start_id, request.end_id, request.begin_id, request.path_length, request.prev_edge)
			self.runners[request.key] = runner
			go runner.RunRouting()
		}
	}
}
func (self *DistributedRoutingHandler) HandleRetrivelRequest() {
	for {
		request := <-self.retrivel_chan
		runner := self.runners.Get(request.A)
		runner.retrivel_chan <- request.B
	}
}
func (self *DistributedRoutingHandler) HandleStopRequest() {
	for {
		request := <-self.stop_chan
		if self.runners.ContainsKey(request) {
			runner := self.runners.Get(request)
			runner.stop_chan <- true
		}
	}
}
func (self *DistributedRoutingHandler) HandleExitRequest() {
	for {
		request := <-self.exit_chan
		if self.runners.ContainsKey(request) {
			runner := self.runners[request]
			runner.exit_chan <- true
			self.runners.Delete(request)
		}
	}
}

type DistributedRoutingManager struct {
	handlers Dict[int16, *DistributedRoutingHandler]
	graph    graph.ITiledGraph

	paths    Dict[int, chan int32]
	finished Dict[int, chan bool]
	stoped   Dict[int, bool]

	routing_chan  chan DD_RoutingRequest
	retrivel_chan chan Tuple[int, int32]
	stop_chan     chan int
	exit_chan     chan int
}

func NewDistributedManager(graph graph.ITiledGraph) *DistributedRoutingManager {
	routing_chan := make(chan DD_RoutingRequest, 100)
	retrivel_chan := make(chan Tuple[int, int32], 100)
	stop_chan := make(chan int, 10)
	exit_chan := make(chan int, 10)

	paths := NewDict[int, chan int32](10)
	finished := NewDict[int, chan bool](10)
	stoped := NewDict[int, bool](100)

	tile_count := graph.TileCount()
	handlers := NewDict[int16, *DistributedRoutingHandler](int(tile_count))

	m := &DistributedRoutingManager{
		handlers: handlers,
		graph:    graph,

		paths:    paths,
		finished: finished,
		stoped:   stoped,

		routing_chan:  routing_chan,
		retrivel_chan: retrivel_chan,
		stop_chan:     stop_chan,
		exit_chan:     exit_chan,
	}

	for i := int16(1); i < tile_count+1; i++ {
		m.handlers[i] = NewDistributedHandler(i, m, graph)
	}

	go m.HandleRoutingRequest()
	go m.HandleRetrivelRequest()
	go m.HandleStopRequest()
	go m.HandleExitRequest()

	return m
}

func (self *DistributedRoutingManager) HandleRoutingRequest() {
	for {
		request := <-self.routing_chan
		tile_id := self.graph.GetNodeTile(request.begin_id)
		if self.handlers.ContainsKey(tile_id) {
			handler := self.handlers.Get(tile_id)
			handler.routing_chan <- request
		} else {
			handler := NewDistributedHandler(tile_id, self, self.graph)
			handler.routing_chan <- request
			self.handlers[tile_id] = handler
		}
	}
}
func (self *DistributedRoutingManager) HandleRetrivelRequest() {
	for {
		request := <-self.retrivel_chan
		tile_id := self.graph.GetNodeTile(request.B)
		handler := self.handlers.Get(tile_id)
		handler.retrivel_chan <- request
	}
}
func (self *DistributedRoutingManager) HandleStopRequest() {
	for {
		request := <-self.stop_chan
		self.finished[request] <- true
		self.stoped[request] = true
		for _, handler := range self.handlers {
			handler.stop_chan <- request
		}
	}
}
func (self *DistributedRoutingManager) HandleExitRequest() {
	for {
		request := <-self.exit_chan
		for _, handler := range self.handlers {
			handler.exit_chan <- request
		}
	}
}

func (self *DistributedRoutingManager) RunRouting(start, end int32) int {
	key := rand.Int()
	for self.paths.ContainsKey(key) {
		key = rand.Int()
	}

	path := make(chan int32)
	finished := make(chan bool)
	self.paths[key] = path
	self.finished[key] = finished

	request := DD_RoutingRequest{
		key:         key,
		start_id:    start,
		end_id:      end,
		begin_id:    start,
		path_length: 0,
		prev_edge:   -1,
	}
	self.routing_chan <- request

	<-self.finished[key]
	return key
}
func (self *DistributedRoutingManager) GetRoutingPath(key int) List[int32] {
	path := NewList[int32](10)

	path_chan := self.paths[key]

	for {
		edge, ok := <-path_chan
		if !ok {
			break
		}
		path.Add(edge)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	self.exit_chan <- key
	self.paths.Delete(key)

	return path
}

type DistributedDijkstra struct {
	manager *DistributedRoutingManager

	start_id int32
	end_id   int32

	key int
}

func NewDistributedDijkstra(manager *DistributedRoutingManager, start, end int32) *DistributedDijkstra {
	return &DistributedDijkstra{
		manager:  manager,
		start_id: start,
		end_id:   end,
		key:      -1,
	}
}

func (self *DistributedDijkstra) CalcShortestPath() bool {
	fmt.Println("Start RunRouting")
	key := self.manager.RunRouting(self.start_id, self.end_id)
	fmt.Println("Finished RunRouting")
	self.key = key
	return true
}

func (self *DistributedDijkstra) Steps(count int, visitededges *List[geo.CoordArray]) bool {
	key := self.manager.RunRouting(self.start_id, self.end_id)
	self.key = key
	return false
}

func (self *DistributedDijkstra) GetShortestPath() Path {
	edges := self.manager.GetRoutingPath(self.key)
	return NewPath(self.manager.graph, edges)
}
