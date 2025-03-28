package main

import (
	"fmt"
	"math"

	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/routing"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

//**********************************************************
// isoraster request and response
//**********************************************************

type IsoRasterRequest struct {
	Locations  [][]float32 `json:"locations"`
	Range      int32       `json:"range"`
	Precession int32       `json:"precession"`
	Profile    string      `json:"profile"`
	Metric     string      `json:"metric"`
}

type IsoRasterResponse struct {
	Type     string        `json:"type"`
	Features []geo.Feature `json:"features"`
}

func NewIsoRasterResponse(nodes []*QuadNode[int], rasterizer IRasterizer) IsoRasterResponse {
	resp := IsoRasterResponse{}
	resp.Type = "FeatureCollection"

	resp.Features = make([]geo.Feature, len(nodes))
	for i := 0; i < len(nodes); i++ {
		ul := rasterizer.IndexToPoint(nodes[i].X, nodes[i].Y)
		lr := rasterizer.IndexToPoint(nodes[i].X+1, nodes[i].Y+1)
		line := make([]geo.Coord, 5)
		line[0][0] = ul[0]
		line[0][1] = ul[1]
		line[1][0] = lr[0]
		line[1][1] = ul[1]
		line[2][0] = lr[0]
		line[2][1] = lr[1]
		line[3][0] = ul[0]
		line[3][1] = lr[1]
		line[4][0] = ul[0]
		line[4][1] = ul[1]
		geom := geo.NewPolygon([][]geo.Coord{line})
		props := NewDict[string, any](1)
		props["value"] = nodes[i].Value
		resp.Features[i] = geo.NewFeature(&geom, props)
	}
	return resp
}

//**********************************************************
// isoraster handler
//**********************************************************

func HandleIsoRasterRequest(req IsoRasterRequest) Result {
	// get profile
	profile_, res := GetRequestProfile(MANAGER, req.Profile, req.Metric)
	if !profile_.HasValue() {
		return res
	}
	profile := profile_.Value
	g_ := profile.GetGraph()
	if !g_.HasValue() {
		return BadRequest("Graph not found")
	}
	g := g_.Value
	att := profile.GetAttributes()

	start := geo.Coord{req.Locations[0][0], req.Locations[0][1]}
	consumer := &SPTConsumer{
		points: NewQuadTree(func(val1, val2 int) int {
			if val1 < val2 {
				return val1
			} else {
				return val2
			}
		}),
		rasterizer: NewDefaultRasterizer(req.Precession),
	}
	s_node, _ := att.GetClosestNode(start)
	spt := routing.NewShortestPathTree(g, s_node, req.Range, consumer)

	slog.Debug(fmt.Sprintf("Start Caluclating shortest-path-tree from %v", start))
	spt.CalcShortestPathTree()
	slog.Debug("shortest-path-tree finished")
	slog.Debug("start building response")
	resp := NewIsoRasterResponse(consumer.points.ToSlice(), consumer.rasterizer)
	slog.Debug("reponse build")
	return OK(resp)
}

//**********************************************************
// isoraster builder
//**********************************************************

type SPTConsumer struct {
	points     *QuadTree[int]
	rasterizer IRasterizer
}

func (self *SPTConsumer) ConsumePoint(point geo.Coord, value int) {
	x, y := self.rasterizer.PointToIndex(point)
	self.points.Insert(x, y, value)
}

func (self *SPTConsumer) ConsumeEdge(edge int32, start_value int, end_value int) {}

type IProjection interface {
	Proj(geo.Coord) geo.Coord
	ReProj(geo.Coord) geo.Coord
}

type IRasterizer interface {
	PointToIndex(geo.Coord) (int32, int32)
	IndexToPoint(int32, int32) geo.Coord
}

type DefaultRasterizer struct {
	projection IProjection
	factor     float32
}

func NewDefaultRasterizer(precession int32) *DefaultRasterizer {
	return &DefaultRasterizer{
		factor:     1 / float32(precession),
		projection: &WebMercatorProjection{},
	}
}

func NewDummyRasterizer(precession int32) *DefaultRasterizer {
	return &DefaultRasterizer{
		factor:     1 / float32(precession),
		projection: &NullProjection{},
	}
}

func (self *DefaultRasterizer) PointToIndex(point geo.Coord) (int32, int32) {
	c := self.projection.Proj(point)
	return int32(c[0] * self.factor), int32(c[1] * self.factor)
}
func (self *DefaultRasterizer) IndexToPoint(x, y int32) geo.Coord {
	point := geo.Coord{float32(x) / self.factor, float32(y) / self.factor}
	return self.projection.ReProj(point)
}

type WebMercatorProjection struct{}

func (self *WebMercatorProjection) Proj(point geo.Coord) geo.Coord {
	a := 6378137.0
	c := geo.Coord{}
	c[0] = float32(a * float64(point[0]) * math.Pi / 180)
	c[1] = float32(a * math.Log(math.Tan(math.Pi/4+float64(point[1])*math.Pi/360)))
	return c
}
func (self *WebMercatorProjection) ReProj(point geo.Coord) geo.Coord {
	a := 6378137.0
	c := geo.Coord{}
	c[0] = float32(float64(point[0]) * 180 / (a * math.Pi))
	c[1] = float32(360 * (math.Atan(math.Exp(float64(point[1])/a)) - math.Pi/4) / math.Pi)
	return c
}

type NullProjection struct{}

func (self *NullProjection) Proj(point geo.Coord) geo.Coord {
	return point
}
func (self *NullProjection) ReProj(point geo.Coord) geo.Coord {
	return point
}
