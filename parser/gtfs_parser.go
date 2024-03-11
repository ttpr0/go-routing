package parser

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// gtfs parser
//*******************************************

func ParseGtfs(gtfs_path string, filter_polygon string) (Array[structs.Node], Array[structs.Connection], Dict[string, List[[]comps.ConnectionWeight]]) {
	file_str, err := os.ReadFile(filter_polygon)
	if err != nil {
		panic(err.Error())
	}
	collection := geo.FeatureCollection{}
	err = json.Unmarshal(file_str, &collection)
	if err != nil {
		panic(err.Error())
	}
	filter := collection.Features()[0].Geometry()

	_stops := _ReadStopLocations(gtfs_path, filter)
	_services := _ReadCalendar(gtfs_path)
	_trips := _ReadTrips(gtfs_path, _stops, _services)

	stops, conns, schedules := _BuildTransitGraph(_trips, _stops, _services)
	return Array[structs.Node](stops), Array[structs.Connection](conns), schedules
}

//*******************************************
// parser utility
//*******************************************

func ReadCSV[T any](filename string) []T {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, _ := reader.Read()
	name_row_mapping := NewDict[string, int](10)
	for i, name := range header {
		name_row_mapping[name] = i
	}

	var val T
	typ := reflect.TypeOf(val)
	num_field := typ.NumField()
	fields := NewList[Triple[int, int, reflect.Kind]](num_field)
	for i := 0; i < num_field; i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("csv")
		if tag == "" {
			continue
		}
		if !name_row_mapping.ContainsKey(tag) {
			continue
		}
		row := name_row_mapping[tag]
		switch field.Type.Kind() {
		case reflect.Bool:
			fields.Add(MakeTriple(i, row, reflect.Bool))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fields.Add(MakeTriple(i, row, reflect.Int))
		case reflect.Float32, reflect.Float64:
			fields.Add(MakeTriple(i, row, reflect.Float64))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fields.Add(MakeTriple(i, row, reflect.Uint))
		case reflect.String:
			fields.Add(MakeTriple(i, row, reflect.String))
		}
	}

	records := NewList[T](100)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err == csv.ErrFieldCount {
			continue
		}

		t := reflect.New(typ).Elem()
		for _, field := range fields {
			index := field.A
			row := field.B
			typ := field.C
			value := record[row]
			if value == "" {
				continue
			}
			f := t.Field(index)
			switch typ {
			case reflect.Bool:
				num, _ := strconv.ParseBool(value)
				f.SetBool(num)
			case reflect.Int:
				num, _ := strconv.ParseInt(value, 10, 64)
				f.SetInt(num)
			case reflect.Uint:
				num, _ := strconv.ParseUint(value, 10, 64)
				f.SetUint(num)
			case reflect.Float64:
				num, _ := strconv.ParseFloat(value, 64)
				f.SetFloat(num)
			case reflect.String:
				f.SetString(value)
			}
		}
		value := t.Interface().(T)
		records.Add(value)
	}

	return records
}

type GTFSCalendarEntry struct {
	ServiceID int `csv:"service_id"`
	Monday    int `csv:"monday"`
	Tuesday   int `csv:"tuesday"`
	Wednesday int `csv:"wednesday"`
	Thursday  int `csv:"thursday"`
	Friday    int `csv:"friday"`
	Saturday  int `csv:"saturday"`
	Sunday    int `csv:"sunday"`
}

type GTFSStopEntry struct {
	StopID int     `csv:"stop_id"`
	Lon    float32 `csv:"stop_lon"`
	Lat    float32 `csv:"stop_lat"`
	Parent int     `csv:"parent_station"`
	Type   int     `csv:"location_type"`
}

type GTFSStopTimesEntry struct {
	TripID    int    `csv:"trip_id"`
	Arival    string `csv:"arrival_time"`
	Departure string `csv:"departure_time"`
	StopID    int    `csv:"stop_id"`
	StopSeq   int    `csv:"stop_sequence"`
}

type GTFSTripsEntry struct {
	TripID    int `csv:"trip_id"`
	RouteID   int `csv:"route_id"`
	ServiceID int `csv:"service_id"`
}

//*******************************************
// parse services
//*******************************************

type GTFSService struct {
	service_id int
	days       []int
}

func (self *GTFSService) GetDays() []int {
	return self.days
}

func _ReadCalendar(path string) Dict[int, GTFSService] {
	data := ReadCSV[GTFSCalendarEntry](path + "/calendar.txt")
	services := NewDict[int, GTFSService](100)
	for _, service := range data {
		service_id := service.ServiceID
		days := NewList[int](3)
		if service.Monday == 1 {
			days.Add(1)
		}
		if service.Tuesday == 1 {
			days.Add(2)
		}
		if service.Wednesday == 1 {
			days.Add(3)
		}
		if service.Thursday == 1 {
			days.Add(4)
		}
		if service.Friday == 1 {
			days.Add(5)
		}
		if service.Saturday == 1 {
			days.Add(6)
		}
		if service.Sunday == 1 {
			days.Add(7)
		}
		services[service_id] = GTFSService{
			service_id: service_id,
			days:       days,
		}
	}
	return services
}

//*******************************************
// parse stops
//*******************************************

type GTFSStop struct {
	stop_id   int
	lon       float32
	lat       float32
	typ       int
	parent_id int
}

func (self *GTFSStop) HasParent() bool {
	return self.parent_id >= 0
}
func (self *GTFSStop) GetParent() int {
	return self.parent_id
}
func (self *GTFSStop) GetLonLat() (float32, float32) {
	return self.lon, self.lat
}

func _ReadStopLocations(path string, filter geo.Geometry) Dict[int, GTFSStop] {
	frame := ReadCSV[GTFSStopEntry](path + "/stops.txt")
	stops := NewDict[int, GTFSStop](100)
	point := geo.Point{}
	for _, entry := range frame {
		id := entry.StopID
		lon := entry.Lon
		lat := entry.Lat
		parent := entry.Parent
		typ := entry.Type
		if lon == 0 || lat == 0 || typ >= 2 {
			if parent == 0 {
				continue
			}
			stops[id] = GTFSStop{id, 0, 0, typ, parent}
		} else {
			point.SetCoordinates(geo.Coord{lon, lat})
			if !filter.Contains(&point) {
				continue
			}
			if parent == 0 {
				parent = -1
			}
			stops[id] = GTFSStop{id, lon, lat, typ, parent}
		}
	}

	delete := NewList[int](10)
	for id, stop := range stops {
		if stop.HasParent() {
			parent_id := stop.GetParent()
			if !stops.ContainsKey(stop.parent_id) {
				delete.Add(id)
				continue
			}
			if stop.typ == 4 {
				parent := stops[parent_id]
				if parent.HasParent() {
					stop.parent_id = parent.GetParent()
					stops[parent_id] = stop
				}
				if !stops.ContainsKey(stop.parent_id) {
					delete.Add(id)
					continue

				}
			}
		}
	}
	for _, d := range delete {
		stops.Delete(d)
	}
	return stops

}

//*******************************************
// parse trips
//*******************************************

type GTFSTripStop struct {
	stop_id   int
	arrival   int
	departure int
	sequence  int
}

type GTFSTrip struct {
	trip_id    int
	route_id   int
	service_id int
	stops      List[GTFSTripStop]
}

func (self *GTFSTrip) SetServiceID(service_id int) {
	self.service_id = service_id
}
func (self *GTFSTrip) SetRouteID(route_id int) {
	self.route_id = route_id
}
func (self *GTFSTrip) AddStop(stop GTFSTripStop) {
	self.stops.Add(stop)
}
func (self *GTFSTrip) OrderStops() {
	sort.Slice(self.stops, func(i, j int) bool {
		return self.stops[i].sequence < self.stops[j].sequence
	})
}

func _ParseTime(time_str string) int {
	tokens := strings.Split(time_str, ":")
	time := 0
	dt, _ := strconv.Atoi(tokens[2])
	time += dt
	dt, _ = strconv.Atoi(tokens[1])
	time += dt * 60
	dt, _ = strconv.Atoi(tokens[0])
	time += dt * 3600
	return time
}

func _ReadTrips(path string, stops Dict[int, GTFSStop], services Dict[int, GTFSService]) Dict[int, GTFSTrip] {
	trips := NewDict[int, GTFSTrip](10)
	frame := ReadCSV[GTFSStopTimesEntry](path + "/stop_times.txt")
	for _, entry := range frame {
		trip_id := entry.TripID
		if !trips.ContainsKey(trip_id) {
			trips[trip_id] = GTFSTrip{
				trip_id: trip_id,
			}
		}
		trip := trips[trip_id]
		s_id := entry.StopID
		if !stops.ContainsKey(s_id) {
			continue
		}
		a_time := _ParseTime(entry.Arival)
		d_time := _ParseTime(entry.Departure)
		s_seq := entry.StopSeq
		trip.AddStop(GTFSTripStop{s_id, a_time, d_time, s_seq})
		trips[trip_id] = trip
	}
	for _, trip := range trips {
		trip.OrderStops()
	}

	frame2 := ReadCSV[GTFSTripsEntry](path + "/trips.txt")
	for _, entry := range frame2 {
		trip_id := entry.TripID
		if !trips.ContainsKey(trip_id) {
			continue
		}
		trip := trips[trip_id]
		route_id := entry.RouteID
		trip.SetRouteID(route_id)
		service_id := entry.ServiceID
		if !services.ContainsKey(service_id) {
			continue
		}
		trip.SetServiceID(service_id)
		trips[trip_id] = trip
	}

	return trips
}

//*******************************************
// parse to graph
//*******************************************

func _BuildTransitGraph(trips Dict[int, GTFSTrip], stops Dict[int, GTFSStop], services Dict[int, GTFSService]) (List[structs.Node], List[structs.Connection], Dict[string, List[[]comps.ConnectionWeight]]) {
	stops_vec := NewList[structs.Node](10)
	stop_mapping := NewDict[int, int](10)
	skiped := NewList[int](10)
	for stop_id, stop := range stops {
		if stop.HasParent() {
			skiped.Add(stop_id)
		} else {
			new_id := len(stops_vec)
			stop_mapping[stop_id] = new_id
			stops_vec.Add(structs.Node{
				Loc: geo.Coord{stop.lon, stop.lat},
			})
		}
	}
	for _, stop_id := range skiped {
		stop := stops[stop_id]
		parent_id := stop.GetParent()
		stop_mapping[stop_id] = stop_mapping[parent_id]
	}

	conns_vec := NewList[structs.Connection](10)
	conn_mapping := NewDict[Triple[int, int, int], int](10)
	schedules := Dict[string, List[[]comps.ConnectionWeight]]{
		"monday":    NewList[[]comps.ConnectionWeight](10),
		"tuesday":   NewList[[]comps.ConnectionWeight](10),
		"wednesday": NewList[[]comps.ConnectionWeight](10),
		"thursday":  NewList[[]comps.ConnectionWeight](10),
		"friday":    NewList[[]comps.ConnectionWeight](10),
		"saturday":  NewList[[]comps.ConnectionWeight](10),
		"sunday":    NewList[[]comps.ConnectionWeight](10),
	}
	for _, trip := range trips {
		if trip.service_id == -1 || trip.route_id == -1 {
			continue
		}
		route_id := trip.route_id
		service := services[trip.service_id]
		days := service.GetDays()
		trip_stops := trip.stops
		for i := 0; i < len(trip_stops)-1; i++ {
			curr_t_stop := trip_stops[i]
			next_t_stop := trip_stops[i+1]
			stop_a := stop_mapping[curr_t_stop.stop_id]
			stop_b := stop_mapping[next_t_stop.stop_id]
			dep := curr_t_stop.departure
			arr := next_t_stop.arrival
			var conn_id int
			if !conn_mapping.ContainsKey(MakeTriple(stop_a, stop_b, route_id)) {
				conn := structs.Connection{
					StopA:   int32(stop_a),
					StopB:   int32(stop_b),
					RouteID: int32(route_id),
				}
				conns_vec.Add(conn)
				conn_id = len(conns_vec) - 1
				conn_mapping[MakeTriple(stop_a, stop_b, route_id)] = conn_id
				for _, day := range []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"} {
					schedule := schedules[day]
					schedule.Add(NewList[comps.ConnectionWeight](2))
					schedules[day] = schedule
				}
			} else {
				conn_id = conn_mapping[MakeTriple(stop_a, stop_b, route_id)]
			}
			for _, day := range days {
				var sc List[[]comps.ConnectionWeight]
				switch day {
				case 1:
					sc = schedules["monday"]
				case 2:
					sc = schedules["tuesday"]
				case 3:
					sc = schedules["wednesday"]
				case 4:
					sc = schedules["thursday"]
				case 5:
					sc = schedules["friday"]
				case 6:
					sc = schedules["saturday"]
				case 7:
					sc = schedules["sunday"]
				default:
					panic("Invalid day: " + string(day))
				}
				sc[conn_id] = append(sc[conn_id], comps.ConnectionWeight{
					Departure: int32(dep),
					Arrival:   int32(arr),
				})
			}
		}
	}
	return stops_vec, conns_vec, schedules
}
