# go-routing

*go-routing* can be used both as an engine to solve route-planning problems using state-of-the-art algorithms as well as an API aiming to power complex spatial accessibility problems.

## Overview

Though it can be extended to serve as a general purpose routing-engine the primary use case of *go-routing* is to be used in cases where network distances between multiple start and potentially many target locations are needed (e.g spatial-accessibility problems). Thus this service is built around a set of methods to compute batched-shortest-paths on both static as well as public-transit networks.

Features primarily include:

* Computation of travel-time matrices using state-of-the-art algorithms
* Building graphs from standard data-sources (OSM, GTFS)
* Static routing for different vehicles (car, bike, foot)
* Computing batched-shortest-paths within a time-window using public-transit
* Dynamic scenarios including avoid-road and avoid-area restrictions

## Installation

To run the API a *Dockerfile* as well as a *docker compose* config are provided.

Using *docker compose* installing and launching the latest version of *go-routing* can be done as follows:

```bash
$ git clone https://github.com/ttpr0/go-routing.git
$ cd go-routing
$ docker compose up --build -d
```

Configuration happens through the *config.yml* file. An example configuration looks as follows:

```yaml
build:
  source: # data-sources used to create routing networks from
    osm: "./data/saarland.pbf"
    gtfs: "./data/gtfs"
  profiles: # list of profile configurations to be build
    driving-car: # profile name
      type: "driving" # one of ["driving", "walking", "cycling", "transit"]; graphs are parsed depending on the type
      vehicle: "car" # the vehicle used to traverse the network
      metric: "fastest" # ["fastest", "shortest"]; together with vehicle this controls the weighting of the network
      preparation: # optional parameters defining additional preprocessing steps
        contraction: true # if this is set to true graph will be contracted which makes it possible to compute batched-shortest-paths more efficiently
    public-transit:
      type: "transit" # this type creates a public-transit network from the GTFS data-source
      vehicle: "foot" # transit network will be embedded into a graph with this vehicle
      preparation:
        filter-polygon: "./data/berlin.json" # optionally filters the GTFS-data to the provided polygon extent
        max-transfer-range: 900 # denotes the maximum range allowed between transit-stops (e.g. 900 -> maximum 15min walk between stations)
build-graphs: false # is set to true graphs will be built as specified in build value; build will always happen if none are found
```

## Usage

The main API computes a travel-time-matrix between a set of start- and target points (POST /v1/matrix). An example request looks as follows:

```js
{
  "sources": [[lon, lat], ...], // start points
  "destinations": [[lon, lat], ...], // target points
  "profile": "driving-car", // profile used during routing; made up of the profiles type and vehicle config: "type-vehicle"
  "metric": "fastest", // metric/weighting used during routing
  "max_range": 1800, // optionally a maximum range (in s) can be specified to make computation more efficient (most accessibility algorithms only require ranges up to a distance threshold)
  "time_window": [28800, 36000], // if public-transit is used this denotes the time-span during which routes are allowed to start (e.g. 28800s-36000s = 8h - 10h).
  "schedule_day": "monday", // weekday of travel for public-transit (transit graph is built with schedules for every day of the week)
  "avoid_roads": ["motorway", ...], // list of road-types to be avoided during search
  "avoid_area": {...} // geojson polygon/multi-polygon feature specifying an area to be avoided during search
}
```
