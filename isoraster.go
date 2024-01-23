package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/routing"
	"github.com/ttpr0/go-routing/util"
)

type IsoRasterRequest struct {
	Locations  [][]float32 `json:"locations"`
	Range      int32       `json:"range"`
	Precession int32       `json:"precession"`
}

type IsoRasterResponse struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

func NewIsoRasterResponse(nodes []*util.QuadNode[int], rasterizer IRasterizer) IsoRasterResponse {
	resp := IsoRasterResponse{}
	resp.Type = "FeatureCollection"

	resp.Features = make([]GeoJSONFeature, len(nodes))
	for i := 0; i < len(nodes); i++ {
		ul := rasterizer.IndexToPoint(nodes[i].X, nodes[i].Y)
		lr := rasterizer.IndexToPoint(nodes[i].X+1, nodes[i].Y+1)
		line := make([][2]float32, 5)
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
		resp.Features[i] = NewGeoJSONFeature()
		resp.Features[i].Geom["type"] = "Polygon"
		resp.Features[i].Geom["coordinates"] = [][][2]float32{line}
		resp.Features[i].Props["value"] = nodes[i].Value
	}
	return resp
}

func HandleIsoRasterRequest(w http.ResponseWriter, r *http.Request) {
	data := make([]byte, r.ContentLength)
	r.Body.Read(data)
	req := IsoRasterRequest{}
	json.Unmarshal(data, &req)

	start := geo.Coord{req.Locations[0][0], req.Locations[0][1]}
	consumer := &SPTConsumer{
		points: util.NewQuadTree(func(val1, val2 int) int {
			if val1 < val2 {
				return val1
			} else {
				return val2
			}
		}),
		rasterizer: NewDefaultRasterizer(req.Precession),
	}
	spt := routing.NewShortestPathTree(GRAPH, GetClosestNode(start, GRAPH), req.Range, consumer)

	fmt.Println("Start Caluclating shortest-path-tree from", start)
	spt.CalcShortestPathTree()
	fmt.Println("shortest-path-tree finished")
	fmt.Println("start building response")
	resp := NewIsoRasterResponse(consumer.points.ToSlice(), consumer.rasterizer)
	fmt.Println("reponse build")
	data, _ = json.Marshal(resp)
	w.Write(data)
}

type SPTConsumer struct {
	points     *util.QuadTree[int]
	rasterizer IRasterizer
}

func (self *SPTConsumer) ConsumePoint(point geo.Coord, value int) {
	x, y := self.rasterizer.PointToIndex(point)
	self.points.Insert(x, y, value)
}

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
