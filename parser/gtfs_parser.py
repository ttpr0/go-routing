from typing import Any
import json
import argparse
import pandas as pd
import numpy as np
from shapely import contains_xy, MultiPolygon, Polygon

class GTFSService:
    __slots__ = ["service_id", "days"]
    service_id: int
    days: list[int]

    def __init__(self, service_id, days: list[int]):
        self.service_id = service_id
        self.days = days

    def get_days(self):
        return self.days

def read_calendar(path: str) -> dict[int, GTFSService]:
    frame = pd.read_csv(f"{path}/calendar.txt")
    service_ids = frame["service_id"]
    monday = frame["monday"]
    tuesday = frame["tuesday"]
    wednesday = frame["wednesday"]
    thursday = frame["thursday"]
    friday = frame["friday"]
    saturday = frame["saturday"]
    sunday = frame["sunday"]
    services: dict[int, GTFSService] = {}
    for i in range(service_ids.size):
        service_id = int(service_ids[i])
        days = []
        if monday[i] == 1:
            days.append(1)
        if tuesday[i] == 1:
            days.append(2)
        if wednesday[i] == 1:
            days.append(3)
        if thursday[i] == 1:
            days.append(4)
        if friday[i] == 1:
            days.append(5)
        if saturday[i] == 1:
            days.append(6)
        if sunday[i] == 1:
            days.append(7)
        services[service_id] = GTFSService(service_id, days)
    return services

class GTFSStop:
    __slots__ = ["stop_id", "lon", "lat", "typ", "parent_id"]
    stop_id: int
    lon: float
    lat: float
    typ: int
    parent_id: int | None

    def __init__(self, id, lon, lat, typ, parent):
        self.stop_id = id
        self.lat = lat
        self.lon = lon
        self.typ = typ
        self.parent_id = parent

    def has_parent(self) -> bool:
        return self.parent_id is not None
    
    def get_parent(self) -> int:
        assert self.parent_id is not None
        return self.parent_id

    def get_lon_lat(self) -> tuple[float, float]:
        return self.lon, self.lat

def read_stop_locations(path: str, filter: MultiPolygon) -> dict[int, GTFSStop]:
    stops_frame = pd.read_csv(f"{path}/stops.txt")
    stop_ids = stops_frame["stop_id"]
    stop_lon = stops_frame["stop_lon"]
    stop_lat = stops_frame["stop_lat"]
    stop_parents = stops_frame["parent_station"]
    location_type = stops_frame["location_type"]
    stops: dict[int, GTFSStop] = {}
    for i in range(stop_ids.size):
        id = int(stop_ids[i])
        _lon = stop_lon[i]
        _lat = stop_lat[i]
        _parent = stop_parents[i]
        _typ = location_type[i]
        if np.isnan(_typ):
            typ = 0
        else:
            typ = int(_typ)
        if np.isnan(_lon) or np.isnan(_lat) or typ >= 2:
            if np.isnan(_parent):
                continue
            stops[id] = GTFSStop(id, 0, 0, typ, int(_parent))
        else:
            lon = float(_lon)
            lat = float(_lat)
            if not contains_xy(filter, lon, lat):
                continue
            if np.isnan(_parent):
                parent = None
            else:
                parent = int(_parent)
            stops[id] = GTFSStop(id, lon, lat, int(typ), parent)

    delete = []
    for id, stop in stops.items():
        if stop.has_parent():
            parent_id = stop.get_parent()
            if stop.parent_id not in stops:
                delete.append(id)
                continue
            if stop.typ == 4:
                parent = stops[parent_id]
                if parent.has_parent():
                    stop.parent_id = parent.get_parent()
                if stop.parent_id not in stops:
                    delete.append(id)
                    continue
    for d in delete:
        del stops[d]
    return stops


class GTFSTripStop:
    __slots__ = ["stop_id", "arrival", "departure", "sequence"]
    stop_id: int
    arrival: int
    departure: int
    sequence: int

    def __init__(self, stop_id, arival, departure, sequence):
        self.stop_id = stop_id
        self.arrival = arival
        self.departure = departure
        self.sequence = sequence


class GTFSTrip:
    __slots__ = ["trip_id", "route_id", "service_id", "stops"]
    trip_id: int
    route_id: int
    service_id: int
    stops: list[GTFSTripStop]

    def __init__(self, trip_id):
        self.trip_id = trip_id
        self.route_id = -1
        self.service_id = -1
        self.stops = []

    def set_service_id(self, service_id):
        self.service_id = service_id

    def set_route_id(self, route_id):
        self.route_id = route_id

    def add_stop(self, stop: GTFSTripStop):
        self.stops.append(stop)

    def order_stops(self):
        self.stops.sort(key=lambda x: x.sequence)

def parse_time(time_str: str) -> int:
    tokens = time_str.split(":")
    time = 0
    time += int(tokens[2])
    time += int(tokens[1]) * 60
    time += int(tokens[0]) * 3600
    return time

def read_trips(path: str, stops: dict[int, GTFSStop], services: dict[int, GTFSService]) -> dict[int, GTFSTrip]:
    trips: dict[int, GTFSTrip] = {}
    
    times_frame = pd.read_csv(f"{path}/stop_times.txt")
    trip_ids = times_frame["trip_id"]
    arrival_times = times_frame["arrival_time"]
    departure_times = times_frame["departure_time"]
    stop_ids = times_frame["stop_id"]
    stop_sequences = times_frame["stop_sequence"]
    for i in range(trip_ids.size):
        trip_id = int(trip_ids[i])
        if trip_id not in trips:
            trips[trip_id] = GTFSTrip(trip_id)
        trip = trips[trip_id]
        s_id = int(stop_ids[i])
        if s_id not in stops:
            continue
        a_time = parse_time(arrival_times[i])
        d_time = parse_time(departure_times[i])
        s_seq = int(stop_sequences[i])
        trip.add_stop(GTFSTripStop(s_id, a_time, d_time, s_seq))
    for trip in trips.values():
        trip.order_stops()

    frame = pd.read_csv(f"{path}/trips.txt")
    trip_ids = frame["trip_id"]
    route_ids = frame["route_id"]
    service_ids = frame["service_id"]
    for i in range(trip_ids.size):
        trip_id = int(trip_ids[i])
        if trip_id not in trips:
            continue
        trip = trips[trip_id]
        route_id = int(route_ids[i])
        trip.set_route_id(route_id)
        service_id = int(service_ids[i])
        if service_id not in services:
            continue
        trip.set_service_id(service_id)

    return trips

class Node:
    __slots__ = ["lon", "lat"]
    lon: float
    lat: float

    def __init__(self, lon: float, lat: float):
        self.lon = lon
        self.lat = lat

class Connection:
    __slots__ = ["stop_a", "stop_b", "route_id"]
    stop_a: int
    stop_b: int
    route_id: int

    def __init__(self, stop_a: int, stop_b: int, route_id: int):
        self.stop_a = stop_a
        self.stop_b = stop_b
        self.route_id = route_id

def build_transit_graph(trips: dict[int, GTFSTrip], stops: dict[int, GTFSStop], services: dict[int, GTFSService]) -> tuple[list[Node], list[Connection], dict[str, list[list[tuple[int, int]]]]]:
    stops_vec: list[Node] = []
    stop_mapping = {}
    skiped = []
    for stop_id, stop in stops.items():
        if stop.has_parent():
            skiped.append(stop_id)
        else:
            new_id = len(stops_vec)
            stop_mapping[stop_id] = new_id
            stops_vec.append(Node(stop.lon, stop.lat))
    for stop_id in skiped:
        stop = stops[stop_id]
        parent_id = stop.get_parent()
        stop_mapping[stop_id] = stop_mapping[parent_id]

    conns_vec: list[Connection] = []
    conn_mapping = {}
    schedules: dict[str, list[list[tuple[int, int]]]] = {
        "monday": [],
        "tuesday": [],
        "wednesday": [],
        "thursday": [],
        "friday": [],
        "saturday": [],
        "sunday": [],
    }
    for trip_id, trip in trips.items():
        if trip.service_id == -1 or trip.route_id == -1:
            continue
        route_id = trip.route_id
        service = services[trip.service_id]
        days = service.get_days()
        trip_stops = trip.stops
        for i in range(len(trip_stops)-1):
            curr_t_stop = trip_stops[i]
            next_t_stop = trip_stops[i+1]
            stop_a = stop_mapping[curr_t_stop.stop_id]
            stop_b = stop_mapping[next_t_stop.stop_id]
            dep = curr_t_stop.departure
            arr = next_t_stop.arrival
            if (stop_a, stop_b, route_id) not in conn_mapping:
                conn = Connection(stop_a, stop_b, route_id)
                conns_vec.append(conn)
                conn_id = len(conns_vec) - 1
                conn_mapping[(stop_a, stop_b, route_id)] = conn_id
                schedules["monday"].append([])
                schedules["tuesday"].append([])
                schedules["wednesday"].append([])
                schedules["thursday"].append([])
                schedules["friday"].append([])
                schedules["saturday"].append([])
                schedules["sunday"].append([])
            else:
                conn_id = conn_mapping[(stop_a, stop_b, route_id)]
            for day in days:
                sc: list[list[tuple[int, int]]]
                match day:
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
                    case _:
                        raise Exception(f"Invalid day: {day}")
                sc[conn_id].append((dep, arr))
    
    return stops_vec, conns_vec, schedules

def parse_gtfs(gtfs_path: str, filter_polygon: str) -> tuple[list[Node], list[Connection], dict[str, list[list[tuple[int, int]]]]]:
    filter = None
    with open(filter_polygon, "r") as file:
        data = json.loads(file.read())
        features = data["features"]
        coords = features[0]["geometry"]["coordinates"]
        filter = Polygon(coords[0], coords[1:])

    stops = read_stop_locations(gtfs_path, filter)
    services = read_calendar(gtfs_path)
    trips = read_trips(gtfs_path, stops, services)

    return build_transit_graph(trips, stops, services)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        'gtfs_path',
        action='store',
        type=str,
        help="specify path to gtfs-data",
    )
    parser.add_argument(
        '-f',
        '--filter',
        action='store',
        type=str,
        help="specify path to filter-geojson-file",
    )
    parser.add_argument(
        '-o',
        '--output',
        action='store',
        type=str,
        help="specify name of output-file",
    )
    args = parser.parse_args()

    stops, conns, schedules = parse_gtfs(args.input, args.filter)

    with open(args.output, "w") as file:
        file.write(json.dumps({
            "stops": stops,
            "conns": conns,
            "schedules": schedules
        }))
