package geo

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/ttpr0/go-routing/util"
)

func TestMarshalGeometry(t *testing.T) {
	features := NewList[Feature](3)
	poly := NewPolygon([][]Coord{{{0.1, 0.2}, {0.3, 10.4}, {10.5, 10.6}, {10.7, 0.8}}, {{3.1, 3.2}, {7.3, 7.4}, {5.5, 7.6}}})
	features.Add(NewFeature(&poly, NewDict[string, any](3)))
	line := NewLineString([]Coord{{0.9, 0.8}, {5.7, 5.6}, {10.5, 10.4}, {15.3, 15.2}})
	features.Add(NewFeature(&line, NewDict[string, any](3)))
	point := NewPoint(Coord{11.9, 9.11})
	features.Add(NewFeature(&point, NewDict[string, any](3)))
	collection := NewFeatureCollection(features)

	json_str, err := json.Marshal(&collection)
	if err != nil {
		t.Errorf("failed to marshal FeatureCollection")
	}
	file_str, err := os.ReadFile("./test_data/geom.txt")
	if err != nil {
		t.Errorf("missing test data file: geom.txt")
	}
	if string(json_str) != string(file_str) {
		t.Errorf("expected true, got false")
	}
}

func TestUnMarshalGeometry(t *testing.T) {
	file_str, err := os.ReadFile("./test_data/geom.json")
	if err != nil {
		t.Errorf("missing test data file: geom.json")
	}

	collection := FeatureCollection{}
	err = json.Unmarshal(file_str, &collection)
	if err != nil {
		t.Errorf("failed to unmarshal to FeatureCollection")
	}
	features := collection.Features()
	point, ok := (features[0].Geometry()).(*Point)
	if !ok {
		t.Errorf("invalid geometry type: expected Point")
	}
	val := point.Coordinates()[1]
	if val != 10.3 {
		t.Errorf("expected 10.3, got %v", val)
	}

	multiline, ok := (features[1].Geometry()).(*MultiLineString)
	if !ok {
		t.Errorf("invalid geometry type: expected MultiLineString")
	}
	val = multiline.Coordinates()[0][3][1]
	if val != 1.9 {
		t.Errorf("expected 1.9, got %v", val)
	}

	multipoint, ok := (features[2].Geometry()).(*MultiPoint)
	if !ok {
		t.Errorf("invalid geometry type: expected MultiPoint")
	}
	val = multipoint.Coordinates()[1][0]
	if val != -3.8 {
		t.Errorf("expected -3.8, got %v", val)
	}
}
