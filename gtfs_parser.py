from typing import Any
import pandas as pd
import numpy as np
from shapely import contains_xy, MultiPolygon, Polygon
import struct
import json


class GTFSTripStop:
    __slots__ = ["stop_id", "arival", "departure", "sequence"]
    stop_id: int
    arival: int
    departure: int
    sequence: int

    def __init__(self, stop_id, arival, departure, sequence):
        self.stop_id = stop_id
        self.arival = arival
        self.departure = departure
        self.sequence = sequence


class GTFSTrip:
    __slots__ = ["trip_id", "days", "stops"]
    trip_id: int
    days: list[int]
    stops: list[GTFSTripStop]

    def __init__(self, trip_id):
        self.trip_id = trip_id
        self.days = []
        self.stops = []

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


def read_trips(trips: dict[int, GTFSTrip], stop_locs: dict[int, Any]):
    times_frame = pd.read_csv("./data/gtfs/stop_times.txt")
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
        if s_id not in stop_locs:
            continue
        a_time = parse_time(arrival_times[i])
        d_time = parse_time(departure_times[i])
        s_seq = int(stop_sequences[i])

        trip.add_stop(GTFSTripStop(s_id, a_time, d_time, s_seq))

    for trip in trips.values():
        trip.order_stops()

    return trips


def read_trip_days(trips: dict[int, GTFSTrip]):
    frame = pd.read_csv("./data/gtfs/calendar.txt")
    service_ids = frame["service_id"]
    monday = frame["monday"]
    tuesday = frame["tuesday"]
    wednesday = frame["wednesday"]
    thursday = frame["thursday"]
    friday = frame["friday"]
    saturday = frame["saturday"]
    sunday = frame["sunday"]
    services = {}
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

        services[service_id] = days

    frame = pd.read_csv("./data/gtfs/trips.txt")
    trip_ids = frame["trip_id"]
    service_ids = frame["service_id"]
    for i in range(trip_ids.size):
        trip_id = int(trip_ids[i])
        if trip_id not in trips:
            continue
        trip = trips[trip_id]
        service_id = int(service_ids[i])
        if service_id not in services:
            continue
        days = services[service_id]
        trip.days = days


class StopLoc:
    __slots__ = ["stop_id", "lon", "lat", "typ", "parent_id"]
    stop_id: int
    lon: float
    lat: float
    typ: int
    parent_id: int

    def __init__(self, id, lon, lat, typ, parent):
        self.stop_id = id
        self.lat = lat
        self.lon = lon
        self.typ = typ
        self.parent_id = parent

    def has_parant(self) -> bool:
        return self.typ >= 3

    def get_lon_lat(self) -> tuple[float, float]:
        return self.lon, self.lat


def read_stop_locations(filter: MultiPolygon):
    stops_frame = pd.read_csv("./data/gtfs/stops.txt")
    stop_ids = stops_frame["stop_id"]
    stop_lon = stops_frame["stop_lon"]
    stop_lat = stops_frame["stop_lat"]
    stop_parents = stops_frame["parent_station"]
    location_type = stops_frame["location_type"]
    stops: dict[int, StopLoc] = {}
    for i in range(stop_ids.size):
        id = int(stop_ids[i])
        lon = stop_lon[i]
        lat = stop_lat[i]
        parent = stop_parents[i]
        typ = location_type[i]
        if np.isnan(lon) or np.isnan(lat) or typ >= 3:
            if np.isnan(parent):
                continue
            stops[id] = StopLoc(id, 0, 0, int(typ), int(parent))
        else:
            lon = float(lon)
            lat = float(lat)
            if not contains_xy(filter, lon, lat):
                continue
            stops[id] = StopLoc(id, lon, lat, 0, 0)

    delete = []
    for id, stop in stops.items():
        if stop.lat == 0:
            if stop.parent_id not in stops:
                delete.append(id)
                continue
            parent_stop = stops[stop.parent_id]
            stop.lon = parent_stop.lon
            stop.lat = parent_stop.lat
    for d in delete:
        del stops[d]
    return stops


class TransitStop:
    __slots__ = ["stop_id", "lon", "lat", "neighbours"]
    stop_id: int
    lon: float
    lat: float
    # neighbour -> [(day, departure, arival), ...]
    neighbours: dict[int, list[tuple[int, int, int]]]

    def __init__(self, stop_id, lon, lat):
        self.stop_id = stop_id
        self.lon = lon
        self.lat = lat
        self.neighbours = {}

    def add_neighbour(self, other_id: int, days: list[int], dep: int, ar: int):
        for day in days:
            d_neigh = self.neighbours
            if other_id not in d_neigh:
                d_neigh[other_id] = []
            trips = d_neigh[other_id]
            trips.append((day, dep, ar))


def build_transit_graph(trips: dict[int, GTFSTrip], stop_locs: dict[int, StopLoc]) -> dict[int, TransitStop]:
    stops: dict[int, TransitStop] = {}
    for trip in trips.values():
        t_days = trip.days
        t_stops = trip.stops
        if len(t_stops) <= 1:
            continue
        for i in range(len(t_stops)-1):
            curr_t_stop = t_stops[i]
            next_t_stop = t_stops[i+1]

            # check if stops are already registered
            if i == 0:
                if curr_t_stop.stop_id not in stops:
                    s_id = curr_t_stop.stop_id
                    s_loc = stop_locs[s_id]
                    lon, lat = s_loc.get_lon_lat()
                    stops[s_id] = TransitStop(s_id, lon, lat)
            if next_t_stop.stop_id not in stops:
                s_id = next_t_stop.stop_id
                s_loc = stop_locs[s_id]
                lon, lat = s_loc.get_lon_lat()
                stops[s_id] = TransitStop(s_id, lon, lat)

            curr_stop = stops[curr_t_stop.stop_id]
            curr_stop.add_neighbour(
                next_t_stop.stop_id, t_days, curr_t_stop.departure, next_t_stop.arival)
    return stops


def store_transit_graph(stops: dict[int, TransitStop]):
    stop_count = 0
    id_mapping = {}
    store_order = []
    for i, id in enumerate(stops.keys()):
        id_mapping[id] = i
        store_order.append(id)
        stop_count += 1

    with open("./graphs/test/transit_graph", "wb") as file:
        file.write(struct.pack("i", stop_count))
        for s_id in store_order:
            stop = stops[s_id]
            file.write(struct.pack("d", stop.lon))
            file.write(struct.pack("d", stop.lat))
            file.write(struct.pack("i", len(stop.neighbours)))
            for neigh, trips in stop.neighbours.items():
                file.write(struct.pack("i", id_mapping[neigh]))
                file.write(struct.pack("i", len(trips)))
                for trip in trips:
                    file.write(struct.pack("i", trip[0]))
                    file.write(struct.pack("i", trip[1]))
                    file.write(struct.pack("i", trip[2]))


if __name__ == "__main__":
    # filter = MultiPolygon([([(9, 50), (10, 50), (10, 51), (9, 51)], [])])
    filter = None
    with open("./data/niedersachsen.json", "r") as file:
        data = json.loads(file.read())
        features = data["features"]
        coords = features[0]["geometry"]["coordinates"]
        filter = Polygon(coords[0], coords[1:])

    stop_locs = read_stop_locations(filter)

    trips = {}
    read_trips(trips, stop_locs)
    read_trip_days(trips)

    stops = build_transit_graph(trips, stop_locs)

    store_transit_graph(stops)
