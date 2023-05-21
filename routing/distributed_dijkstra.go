package routing

import (
	"fmt"
	"math/rand"
	"sync"

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
	draw        bool
}

type flag_dd struct {
	curr_node   int32
	path_length float64
	prev_edge   int32
}

type IDistributedRunner interface {
	AddRoutingRequest(request DD_RoutingRequest)
	AddRetrivelRequest(request int32)
	AddStopRequest(request bool)
	AddExitRequest(request bool)

	SetMaxLength(length float64)
}

type DistributedRoutingRunner struct {
	routing_chan  *BlockQueue[DD_RoutingRequest]
	retrivel_chan chan int32
	stop_chan     chan bool
	exit_chan     chan bool
	path_chan     chan int32
	edge_queue    *Queue[int32]
	handler       IDistributedHandler

	key        int
	tile_id    int16
	skip       bool
	draw       bool
	heap       SafePriorityQueue[flag_dd, float64]
	start_id   int32
	end_id     int32
	graph      graph.ITiledGraph
	geom       graph.IGeometry
	weight     graph.IWeighting
	flags      SafeDict[int32, flag_dd]
	finished   bool
	is_end     bool
	block      *Block
	max_length float64
}

func NewDistributedRoutingRunner(key int, handler IDistributedHandler, path_chan chan int32, edge_queue *Queue[int32], tile_id int16, skip bool, draw bool, graph graph.ITiledGraph, start, end int32) *DistributedRoutingRunner {
	d := DistributedRoutingRunner{
		routing_chan:  NewBlockQueue[DD_RoutingRequest](),
		retrivel_chan: make(chan int32),
		stop_chan:     make(chan bool),
		exit_chan:     make(chan bool),
		path_chan:     path_chan,
		edge_queue:    edge_queue,
		handler:       handler,

		key:        key,
		tile_id:    tile_id,
		skip:       skip,
		draw:       draw,
		start_id:   start,
		end_id:     end,
		graph:      graph,
		geom:       graph.GetGeometry(),
		weight:     graph.GetWeighting(),
		finished:   false,
		is_end:     false,
		block:      NewBlock(),
		max_length: 10000000000,
	}

	flags := NewSafeDict[int32, flag_dd](100)
	d.flags = flags

	heap := NewSafePriorityQueue[flag_dd, float64](100)
	d.heap = heap

	go d.HandleRetrivalRequest()
	go d.HandleStopRequest()
	go d.HandleExitRequest()

	return &d
}

func (self *DistributedRoutingRunner) HandleRetrivalRequest() {
	for {
		request, ok := <-self.retrivel_chan
		if !ok {
			return
		}
		curr_id := request
		var edge int32
		for {
			if curr_id == self.start_id {
				close(self.path_chan)
				break
			}
			curr_flag := self.flags.Get(curr_id)
			edge = curr_flag.prev_edge
			self.path_chan <- edge
			curr_id, _ = self.graph.GetOtherNode(edge, curr_id)
			if self.graph.GetNodeTile(curr_id) != self.tile_id {
				self.handler.SendRetrivelRequest(MakeTuple(self.key, curr_id))
				break
			}
		}
	}
}
func (self *DistributedRoutingRunner) HandleStopRequest() {
	<-self.stop_chan
	self.finished = true
	self.block.Release()
	if self.is_end {
		self.retrivel_chan <- self.end_id
	}
}
func (self *DistributedRoutingRunner) HandleExitRequest() {
	<-self.exit_chan
	self.finished = true
	close(self.retrivel_chan)
}

func (self *DistributedRoutingRunner) RunRouting() {
	for !self.finished {
		curr_flag, ok := self.heap.Dequeue()
		if !ok {
			self.handler.SetIdle(self.key)
			self.block.Take()
			continue
		}
		if curr_flag.path_length > self.max_length {
			continue
		}
		curr_id := curr_flag.curr_node
		if self.flags.ContainsKey(curr_id) {
			temp_flag := self.flags.Get(curr_id)
			if temp_flag.path_length < curr_flag.path_length {
				continue
			}
		}
		if curr_id == self.end_id {
			fmt.Println("end found")
			self.is_end = true
			self.max_length = curr_flag.path_length
			self.handler.SendStopRequest(self.key, curr_flag.path_length)
		}
		self.flags.Set(curr_id, curr_flag)
		edges := self.graph.GetAdjacentEdges(curr_id, graph.FORWARD)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if self.skip && !ref.IsCrossBorder() && !ref.IsSkip() {
				continue
			}
			edge_id := ref.EdgeID
			new_length := curr_flag.path_length + float64(self.weight.GetEdgeWeight(edge_id))
			if new_length > self.max_length {
				continue
			}
			other_id, _ := self.graph.GetOtherNode(edge_id, curr_id)
			var other_flag flag_dd
			if self.flags.ContainsKey(other_id) {
				other_flag = self.flags.Get(other_id)
			} else {
				other_flag = flag_dd{curr_node: other_id, path_length: 1000000, prev_edge: -1}
			}
			if new_length < other_flag.path_length {
				if self.draw {
					self.edge_queue.Push(edge_id)
				}
				other_flag.curr_node = other_id
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				if ref.IsCrossBorder() {
					request := DD_RoutingRequest{
						key:         self.key,
						start_id:    self.start_id,
						end_id:      self.end_id,
						begin_id:    other_id,
						path_length: new_length,
						prev_edge:   edge_id,
						draw:        self.draw,
					}
					self.handler.SendRoutingRequest(request)
				} else {
					self.heap.Enqueue(other_flag, new_length)
				}
			}
			self.flags.Set(other_id, other_flag)
		}
	}
}

func (self *DistributedRoutingRunner) AddRoutingRequest(request DD_RoutingRequest) {
	if request.path_length > self.max_length {
		return
	}
	if self.flags.ContainsKey(request.begin_id) {
		flag := self.flags.Get(request.begin_id)
		if request.path_length > flag.path_length {
			return
		}
	}
	flag := flag_dd{curr_node: request.begin_id, prev_edge: request.prev_edge, path_length: request.path_length}
	self.heap.Enqueue(flag, request.path_length)
	if self.block.IsTaken() {
		self.handler.SetRunning(self.key)
		self.block.Release()
	}
}
func (self *DistributedRoutingRunner) AddRetrivelRequest(request int32) {
	self.retrivel_chan <- request
}
func (self *DistributedRoutingRunner) AddStopRequest(request bool) {
	self.stop_chan <- request
}
func (self *DistributedRoutingRunner) AddExitRequest(request bool) {
	self.exit_chan <- request
}
func (self *DistributedRoutingRunner) SetMaxLength(length float64) {
	self.max_length = length
}

type IDistributedHandler interface {
	AddRoutingRequest(request DD_RoutingRequest)
	AddRetrivelRequest(request Tuple[int, int32])
	AddStopRequest(request int)
	AddExitRequest(request int)
	AddMaxLengthRequest(request Tuple[int, float64])

	SendRoutingRequest(request DD_RoutingRequest)
	SendRetrivelRequest(request Tuple[int, int32])
	SendStopRequest(key int, path_length float64)

	SetIdle(key int)
	SetRunning(key int)
}

type DistributedHandler struct {
	tile_id     int16
	runners     SafeDict[int, IDistributedRunner]
	max_lengths SafeDict[int, float64]
	manager     IDistributedManager
	graph       graph.ITiledGraph
}

func NewDistributedHandler(tile_id int16, manager IDistributedManager, graph graph.ITiledGraph) *DistributedHandler {
	runners := NewSafeDict[int, IDistributedRunner](10)
	max_lengths := NewSafeDict[int, float64](10)

	h := &DistributedHandler{
		tile_id:     tile_id,
		runners:     runners,
		max_lengths: max_lengths,
		manager:     manager,
		graph:       graph,
	}

	return h
}

func (self *DistributedHandler) AddRoutingRequest(request DD_RoutingRequest) {
	if self.runners.ContainsKey(request.key) {
		runner := self.runners.Get(request.key)
		runner.AddRoutingRequest(request)
	} else if self.manager.IsStoped(request.key) {
		return
	} else {
		var skip bool
		if self.graph.GetNodeTile(request.end_id) == self.tile_id || self.graph.GetNodeTile(request.start_id) == self.tile_id {
			skip = false
		} else {
			skip = true
		}
		runner := NewDistributedRoutingRunner(request.key, self, self.manager.GetPathChannel(request.key), self.manager.GetEdgeQueue(request.key), self.tile_id, skip, request.draw, self.graph, request.start_id, request.end_id)
		if self.max_lengths.ContainsKey(request.key) {
			runner.max_length = self.max_lengths.Get(request.key)
		}
		self.runners.Set(request.key, runner)
		runner.AddRoutingRequest(request)
		self.SetRunning(request.key)
		go runner.RunRouting()
	}
}
func (self *DistributedHandler) AddRetrivelRequest(request Tuple[int, int32]) {
	runner := self.runners.Get(request.A)
	runner.AddRetrivelRequest(request.B)
}
func (self *DistributedHandler) AddStopRequest(request int) {
	if self.runners.ContainsKey(request) {
		runner := self.runners.Get(request)
		runner.AddStopRequest(true)
	}
}
func (self *DistributedHandler) AddExitRequest(request int) {
	if self.runners.ContainsKey(request) {
		runner := self.runners.Get(request)
		runner.AddExitRequest(true)
		self.runners.Delete(request)
		self.max_lengths.Delete(request)
	}
}
func (self *DistributedHandler) AddMaxLengthRequest(request Tuple[int, float64]) {
	self.max_lengths.Set(request.A, request.B)
	if self.runners.ContainsKey(request.A) {
		runner := self.runners.Get(request.A)
		runner.SetMaxLength(request.B)
	}
}
func (self *DistributedHandler) SendRoutingRequest(request DD_RoutingRequest) {
	self.manager.AddRoutingRequest(request)
}
func (self *DistributedHandler) SendRetrivelRequest(request Tuple[int, int32]) {
	self.manager.AddRetrivelRequest(request)
}
func (self *DistributedHandler) SendStopRequest(key int, path_length float64) {
	self.manager.AddStopRequest(key, path_length)
}
func (self *DistributedHandler) SetRunning(key int) {
	self.manager.IncrementRunningCount(key)
}
func (self *DistributedHandler) SetIdle(key int) {
	self.manager.DecrementRunningCount(key)
}

type IDistributedManager interface {
	AddRoutingRequest(request DD_RoutingRequest)
	AddRetrivelRequest(request Tuple[int, int32])
	AddStopRequest(key int, path_length float64)
	AddExitRequest(request int)
	IncrementRunningCount(key int)
	DecrementRunningCount(key int)
	IsStoped(key int) bool
	GetPathChannel(key int) chan int32
	GetEdgeQueue(key int) *Queue[int32]
}

type DistributedManager struct {
	handlers Dict[int16, IDistributedHandler]
	graph    graph.ITiledGraph

	paths       SafeDict[int, chan int32]
	finished    SafeDict[int, chan bool]
	edge_queues SafeDict[int, *Queue[int32]]
	stoped      SafeDict[int, bool]
	req_count   SafeDict[int, int]
	run_count   SafeDict[int, int]
	run_lock    sync.Mutex
	req_lock    sync.Mutex

	routing_chan  *BlockQueue[DD_RoutingRequest]
	retrivel_chan chan Tuple[int, int32]
	stop_chan     chan int
	exit_chan     chan int
}

func NewDistributedManager(graph graph.ITiledGraph) *DistributedManager {
	routing_chan := NewBlockQueue[DD_RoutingRequest]()
	retrivel_chan := make(chan Tuple[int, int32], 100)
	stop_chan := make(chan int, 10)
	exit_chan := make(chan int, 10)

	paths := NewSafeDict[int, chan int32](10)
	finished := NewSafeDict[int, chan bool](10)
	edge_queues := NewSafeDict[int, *Queue[int32]](10)
	stoped := NewSafeDict[int, bool](100)
	req_count := NewSafeDict[int, int](10)
	run_count := NewSafeDict[int, int](10)

	tile_count := graph.TileCount()
	handlers := NewDict[int16, IDistributedHandler](int(tile_count))

	m := &DistributedManager{
		handlers: handlers,
		graph:    graph,

		paths:       paths,
		finished:    finished,
		edge_queues: edge_queues,
		stoped:      stoped,
		req_count:   req_count,
		run_count:   run_count,
		run_lock:    sync.Mutex{},
		req_lock:    sync.Mutex{},

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

func (self *DistributedManager) AddRoutingRequest(request DD_RoutingRequest) {
	self.IncrementRequestCount(request.key)
	self.routing_chan.Push(request)
}
func (self *DistributedManager) AddRetrivelRequest(request Tuple[int, int32]) {
	self.retrivel_chan <- request
}
func (self *DistributedManager) AddStopRequest(key int, path_length float64) {
	if self.run_count.Get(key) == 1 && self.req_count.Get(key) == 0 {
		fmt.Println("stoped1")
		self.stop_chan <- key
	}
	for _, handler := range self.handlers {
		handler.AddMaxLengthRequest(MakeTuple(key, path_length))
	}
}
func (self *DistributedManager) AddExitRequest(request int) {
	self.exit_chan <- request
}

func (self *DistributedManager) DecrementRunningCount(key int) {
	self.run_lock.Lock()
	fmt.Println(self.run_count.Get(key), self.req_count.Get(key))
	if self.run_count.Get(key) == 1 && self.req_count.Get(key) == 0 {
		fmt.Println("stoped2")
		self.stop_chan <- key
	}
	self.run_count.Set(key, self.run_count.Get(key)-1)
	self.run_lock.Unlock()
}
func (self *DistributedManager) IncrementRunningCount(key int) {
	self.run_lock.Lock()
	self.run_count.Set(key, self.run_count.Get(key)+1)
	self.run_lock.Unlock()
}
func (self *DistributedManager) DecrementRequestCount(key int) {
	self.req_lock.Lock()
	if self.run_count.Get(key) == 0 && self.req_count.Get(key) == 1 {
		self.stop_chan <- key
	}
	self.req_count.Set(key, self.req_count.Get(key)-1)
	self.req_lock.Unlock()
}
func (self *DistributedManager) IncrementRequestCount(key int) {
	self.req_lock.Lock()
	self.req_count.Set(key, self.req_count.Get(key)+1)
	self.req_lock.Unlock()
}

func (self *DistributedManager) IsStoped(key int) bool {
	return self.stoped.ContainsKey(key)
}
func (self *DistributedManager) GetPathChannel(key int) chan int32 {
	return self.paths.Get(key)
}
func (self *DistributedManager) GetEdgeQueue(key int) *Queue[int32] {
	return self.edge_queues.Get(key)
}

func (self *DistributedManager) HandleRoutingRequest() {
	for {
		request := self.routing_chan.Pop()
		if self.req_count.Get(request.key) == 0 {
			self.stoped.Delete(request.key)
		}
		tile_id := self.graph.GetNodeTile(request.begin_id)
		if self.handlers.ContainsKey(tile_id) {
			handler := self.handlers.Get(tile_id)
			handler.AddRoutingRequest(request)
		} else {
			handler := NewDistributedHandler(tile_id, self, self.graph)
			handler.AddRoutingRequest(request)
			self.handlers[tile_id] = handler
		}
		self.DecrementRequestCount(request.key)
	}
}
func (self *DistributedManager) HandleRetrivelRequest() {
	for {
		request := <-self.retrivel_chan
		tile_id := self.graph.GetNodeTile(request.B)
		handler := self.handlers.Get(tile_id)
		handler.AddRetrivelRequest(request)
	}
}
func (self *DistributedManager) HandleStopRequest() {
	for {
		request := <-self.stop_chan
		if self.stoped.Get(request) {
			continue
		}
		self.finished.Get(request) <- true
		self.stoped.Set(request, true)
		for _, handler := range self.handlers {
			handler.AddStopRequest(request)
		}
	}
}
func (self *DistributedManager) HandleExitRequest() {
	for {
		request := <-self.exit_chan
		for _, handler := range self.handlers {
			handler.AddExitRequest(request)
		}
	}
}

func (self *DistributedManager) RunRouting(start, end int32) int {
	key := rand.Int()
	for self.paths.ContainsKey(key) {
		key = rand.Int()
	}

	path := make(chan int32)
	finished := make(chan bool)
	self.paths.Set(key, path)
	self.finished.Set(key, finished)
	self.run_count.Set(key, 0)
	self.req_count.Set(key, 0)

	request := DD_RoutingRequest{
		key:         key,
		start_id:    start,
		end_id:      end,
		begin_id:    start,
		path_length: 0,
		prev_edge:   -1,
	}
	self.AddRoutingRequest(request)

	<-self.finished.Get(key)
	self.finished.Delete(key)
	return key
}
func (self *DistributedManager) RunRoutingDraw(start, end int32) (int, *Queue[int32]) {
	key := rand.Int()
	for self.paths.ContainsKey(key) {
		key = rand.Int()
	}

	path := make(chan int32)
	finished := make(chan bool)
	edge_queue := NewQueue[int32]()
	self.paths.Set(key, path)
	self.finished.Set(key, finished)
	self.edge_queues.Set(key, &edge_queue)
	self.run_count.Set(key, 0)
	self.req_count.Set(key, 0)

	request := DD_RoutingRequest{
		key:         key,
		start_id:    start,
		end_id:      end,
		begin_id:    start,
		path_length: 0,
		prev_edge:   -1,
		draw:        true,
	}
	self.AddRoutingRequest(request)

	<-self.finished.Get(key)
	self.finished.Delete(key)
	return key, &edge_queue
}
func (self *DistributedManager) GetRoutingPath(key int) List[int32] {
	path := NewList[int32](10)

	path_chan := self.paths.Get(key)

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
	manager *DistributedManager

	start_id int32
	end_id   int32

	key        int
	edges      List[int32]
	edge_queue *Queue[int32]
}

func NewDistributedDijkstra(manager *DistributedManager, start, end int32) *DistributedDijkstra {
	fmt.Println(start, end)
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
	if self.edge_queue == nil {
		key, queue := self.manager.RunRoutingDraw(self.start_id, self.end_id)
		self.key = key
		self.edge_queue = queue
		self.edges = self.manager.GetRoutingPath(self.key)
	}
	for i := 0; i < count; i++ {
		edge_id, ok := self.edge_queue.Pop()
		if !ok {
			return false
		}
		visitededges.Add(self.manager.graph.GetGeometry().GetEdge(edge_id))
	}
	return true
}

func (self *DistributedDijkstra) GetShortestPath() Path {
	if self.edges == nil {
		self.edges = self.manager.GetRoutingPath(self.key)
	}
	return NewPath(self.manager.graph, self.edges)
}
