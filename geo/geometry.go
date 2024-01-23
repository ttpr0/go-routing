package geo

import (
	"encoding/json"
	"errors"

	. "github.com/ttpr0/go-routing/util"
)

type FeatureCollection struct {
	features List[Feature]
}

func NewFeatureCollection(features List[Feature]) FeatureCollection {
	return FeatureCollection{features: features}
}

func (self *FeatureCollection) Features() []Feature {
	return self.features
}
func (self *FeatureCollection) SetFeatures(features []Feature) {
	self.features = features
}

func (self *FeatureCollection) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Type     string    `json:"type"`
		Features []Feature `json:"features"`
	}{
		Type:     "FeatureCollection",
		Features: self.features,
	})
	return data, err
}
func (self *FeatureCollection) UnmarshalJSON(data []byte) error {
	v := struct {
		Type     string    `json:"type"`
		Features []Feature `json:"features"`
	}{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	self.features = v.Features
	return nil
}

type Feature struct {
	geometry   Geometry
	properties Dict[string, any]
}

func NewFeature(geometry Geometry, properties Dict[string, any]) Feature {
	return Feature{geometry: geometry, properties: properties}
}

func (self *Feature) Geometry() Geometry {
	return self.geometry
}
func (self *Feature) SetGeometry(geometry Geometry) {
	self.geometry = geometry
}
func (self *Feature) Properties() map[string]any {
	return self.properties
}
func (self *Feature) SetProperties(props map[string]any) {
	self.properties = props
}

func (self *Feature) MarshalJSON() ([]byte, error) {
	v := struct {
		Type       string          `json:"type"`
		Geometry   json.RawMessage `json:"geometry"`
		Properties map[string]any  `json:"properties"`
	}{
		Type:       "Feature",
		Geometry:   nil,
		Properties: self.properties,
	}
	switch self.geometry.Type() {
	case "Point":
		point := self.geometry.(*Point)
		v.Geometry, _ = json.Marshal(struct {
			Type        string `json:"type"`
			Coordinates Coord  `json:"coordinates"`
		}{
			Type:        "Point",
			Coordinates: point.coordinates,
		})
	case "MultiPoint":
		point := self.geometry.(*MultiPoint)
		v.Geometry, _ = json.Marshal(struct {
			Type        string  `json:"type"`
			Coordinates []Coord `json:"coordinates"`
		}{
			Type:        "MultiPoint",
			Coordinates: point.coordinates,
		})
	case "LineString":
		point := self.geometry.(*LineString)
		v.Geometry, _ = json.Marshal(struct {
			Type        string  `json:"type"`
			Coordinates []Coord `json:"coordinates"`
		}{
			Type:        "LineString",
			Coordinates: point.coordinates,
		})
	case "MultiLineString":
		point := self.geometry.(*MultiLineString)
		v.Geometry, _ = json.Marshal(struct {
			Type        string    `json:"type"`
			Coordinates [][]Coord `json:"coordinates"`
		}{
			Type:        "MultiLineString",
			Coordinates: point.coordinates,
		})
	case "Polygon":
		point := self.geometry.(*Polygon)
		v.Geometry, _ = json.Marshal(struct {
			Type        string    `json:"type"`
			Coordinates [][]Coord `json:"coordinates"`
		}{
			Type:        "Polygon",
			Coordinates: point.coordinates,
		})
	case "MultiPolygon":
		point := self.geometry.(*MultiPolygon)
		v.Geometry, _ = json.Marshal(struct {
			Type        string      `json:"type"`
			Coordinates [][][]Coord `json:"coordinates"`
		}{
			Type:        "MultiPolygon",
			Coordinates: point.coordinates,
		})
	}
	return json.Marshal(v)
}

func (self *Feature) UnmarshalJSON(data []byte) error {
	v := struct {
		Type     string `json:"type"`
		Geometry struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"`
		} `json:"geometry"`
		Properties map[string]any `json:"properties"`
	}{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	switch v.Geometry.Type {
	case "Point":
		coords := Coord{}
		err := json.Unmarshal(v.Geometry.Coordinates, &coords)
		if err != nil {
			return errors.New("invalid point coordinates")
		}
		self.geometry = &Point{
			coordinates: coords,
		}
	case "MultiPoint":
		coords := []Coord{}
		err := json.Unmarshal(v.Geometry.Coordinates, &coords)
		if err != nil {
			return errors.New("invalid multipoint coordinates")
		}
		self.geometry = &MultiPoint{
			coordinates: coords,
		}
	case "LineString":
		coords := []Coord{}
		err := json.Unmarshal(v.Geometry.Coordinates, &coords)
		if err != nil {
			return errors.New("invalid linestring coordinates")
		}
		self.geometry = &LineString{
			coordinates: coords,
		}
	case "MultiLineString":
		coords := [][]Coord{}
		err := json.Unmarshal(v.Geometry.Coordinates, &coords)
		if err != nil {
			return errors.New("invalid multilinestring coordinates")
		}
		self.geometry = &MultiLineString{
			coordinates: coords,
		}
	case "Polygon":
		coords := [][]Coord{}
		err := json.Unmarshal(v.Geometry.Coordinates, &coords)
		if err != nil {
			return errors.New("invalid polygon coordinates")
		}
		self.geometry = &Polygon{
			coordinates: coords,
		}
	case "MultiPolygon":
		coords := [][][]Coord{}
		err := json.Unmarshal(v.Geometry.Coordinates, &coords)
		if err != nil {
			return errors.New("invalid multipolygon coordinates")
		}
		self.geometry = &MultiPolygon{
			coordinates: coords,
		}
	}
	self.properties = v.Properties
	return nil
}

type Geometry interface {
	Type() string
	Envelope() Envelope
	Contains(inner Geometry) bool
}

type Point struct {
	coordinates Coord
	envelope    Envelope
}

func NewPoint(coords Coord) Point {
	return Point{coordinates: coords}
}

func (self *Point) Type() string {
	return "Point"
}
func (self *Point) Coordinates() Coord {
	return self.coordinates
}
func (self *Point) SetCoordinates(coords Coord) {
	self.coordinates = coords
}
func (self *Point) Envelope() Envelope {
	if self.envelope == (Envelope{}) {
		self.envelope = CalcEnvelope(self)
	}
	return self.envelope
}
func (self *Point) Contains(inner Geometry) bool {
	return false
}

type MultiPoint struct {
	coordinates []Coord
	envelope    Envelope
}

func NewMultiPoint(coords []Coord) MultiPoint {
	return MultiPoint{coordinates: coords}
}

func (self *MultiPoint) Type() string {
	return "MultiPoint"
}
func (self *MultiPoint) Coordinates() []Coord {
	return self.coordinates
}
func (self *MultiPoint) SetCoordinates(coords []Coord) {
	self.coordinates = coords
}
func (self *MultiPoint) Envelope() Envelope {
	if self.envelope == (Envelope{}) {
		self.envelope = CalcEnvelope(self)
	}
	return self.envelope
}
func (self *MultiPoint) Contains(inner Geometry) bool {
	return false
}

type LineString struct {
	coordinates []Coord
	envelope    Envelope
}

func NewLineString(coords []Coord) LineString {
	return LineString{coordinates: coords}
}

func (self *LineString) Type() string {
	return "LineString"
}
func (self *LineString) Coordinates() []Coord {
	return self.coordinates
}
func (self *LineString) SetCoordinates(coords []Coord) {
	self.coordinates = coords
}
func (self *LineString) Envelope() Envelope {
	if self.envelope == (Envelope{}) {
		self.envelope = CalcEnvelope(self)
	}
	return self.envelope
}
func (self *LineString) Contains(inner Geometry) bool {
	return false
}

type MultiLineString struct {
	coordinates [][]Coord
	envelope    Envelope
}

func NewMultiLineString(coords [][]Coord) MultiLineString {
	return MultiLineString{coordinates: coords}
}

func (self *MultiLineString) Type() string {
	return "MultiLineString"
}
func (self *MultiLineString) Coordinates() [][]Coord {
	return self.coordinates
}
func (self *MultiLineString) SetCoordinates(coords [][]Coord) {
	self.coordinates = coords
}
func (self *MultiLineString) Envelope() Envelope {
	if self.envelope == (Envelope{}) {
		self.envelope = CalcEnvelope(self)
	}
	return self.envelope
}
func (self *MultiLineString) Contains(inner Geometry) bool {
	return false
}

type Polygon struct {
	coordinates [][]Coord
	envelope    Envelope
}

func NewPolygon(coords [][]Coord) Polygon {
	return Polygon{coordinates: coords}
}

func (self *Polygon) Type() string {
	return "Polygon"
}
func (self *Polygon) Coordinates() [][]Coord {
	return self.coordinates
}
func (self *Polygon) SetCoordinates(coords [][]Coord) {
	self.coordinates = coords
}
func (self *Polygon) Envelope() Envelope {
	if self.envelope == (Envelope{}) {
		self.envelope = CalcEnvelope(self)
	}
	return self.envelope
}
func (self *Polygon) Contains(inner Geometry) bool {
	if inner.Type() != "Point" {
		return false
	}
	point := inner.(*Point)
	envelope := self.Envelope()
	if envelope.ContainsCoord(point.Coordinates()) {
		return SimplePointInPolygon(point.Coordinates(), self.Coordinates())
	} else {
		return false
	}
}

type MultiPolygon struct {
	coordinates [][][]Coord
	envelope    Envelope
}

func NewMultiPolygon(coords [][][]Coord) MultiPolygon {
	return MultiPolygon{coordinates: coords}
}

func (self *MultiPolygon) Type() string {
	return "MultiPolygon"
}
func (self *MultiPolygon) Coordinates() [][][]Coord {
	return self.coordinates
}
func (self *MultiPolygon) SetCoordinates(coords [][][]Coord) {
	self.coordinates = coords
}
func (self *MultiPolygon) Envelope() Envelope {
	if self.envelope == (Envelope{}) {
		self.envelope = CalcEnvelope(self)
	}
	return self.envelope
}

func (self *MultiPolygon) Contains(inner Geometry) bool {
	if inner.Type() != "Point" {
		return false
	}
	point := inner.(*Point)
	envelope := self.Envelope()
	if envelope.ContainsCoord(point.Coordinates()) {
		for _, coords := range self.Coordinates() {
			is_within := SimplePointInPolygon(point.Coordinates(), coords)
			if is_within {
				return true
			}
		}
	}
	return false
}
