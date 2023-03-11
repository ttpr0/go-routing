import struct
import json
from shapely.predicates import contains_xy
from shapely import Polygon, MultiPolygon

class Edge:
    __slots__ = ['node_a', 'node_b']
    def __init__(self, node_a: int, node_b: int):
        self.node_a: int = node_a
        self.node_b: int = node_b

class EdgeRef:
    __slots__ = ['edge_id', 'typ']
    def __init__(self, edge_id: int, typ: int):
        self.edge_id: int = edge_id
        self.typ: int = typ

class Node:
    __slots__ = ['start', 'count', 'lon', 'lat', 'tile']
    def __init__(self, start: int, count: int, lon: float, lat: float):
        self.start: int = start
        self.count: int = count
        self.lon: float = lon
        self.lat: float = lat
        self.tile: int = -1

class Graph:
    def __init__(self, nodes, edges, edgerefs):
        self.nodes: list[Node] = nodes
        self.edges: list[Edge] = edges
        self.edgerefs: list[EdgeRef] = edgerefs

def load_graph(file: str) -> Graph:
    nodes: list[Node] = []
    edges: list[Edge] = []
    edgerefs: list[EdgeRef] = []

    nodefile = open(file + "-nodes", 'rb')
    edgefile = open(file + "-edges", 'rb')
    geomfile = open(file + "-geom", 'rb')

    nodecount = struct.unpack("i", nodefile.read(4))[0]
    edgerefcount = struct.unpack("i", nodefile.read(4))[0]
    for i in range(nodecount):
        typ = struct.unpack("b", nodefile.read(1))[0]
        start = struct.unpack("i", nodefile.read(4))[0]
        count = struct.unpack("h", nodefile.read(2))[0]
        nodes.append(Node(start, count, 0, 0))
    for i in range(edgerefcount):
        edge_id = struct.unpack("i", nodefile.read(4))[0]
        typ = struct.unpack("b", nodefile.read(1))[0]
        edgerefs.append(EdgeRef(edge_id, typ))

    edgecount = struct.unpack("i", edgefile.read(4))[0]
    for i in range(edgecount):
        node_a = struct.unpack("i", edgefile.read(4))[0]
        node_b = struct.unpack("i", edgefile.read(4))[0]
        w = struct.unpack("b", edgefile.read(1))[0]
        t = struct.unpack("b", edgefile.read(1))[0]
        l = struct.unpack("f", edgefile.read(4))[0]
        m = struct.unpack("b", edgefile.read(1))[0]
        edges.append(Edge(node_a, node_b))
    
    for i in range(nodecount):
        lon = struct.unpack("f", geomfile.read(4))[0]
        lat = struct.unpack("f", geomfile.read(4))[0]
        nodes[i].lon = lon
        nodes[i].lat = lat

    nodefile.close()
    edgefile.close()
    geomfile.close()

    return Graph(nodes, edges, edgerefs)

def store_graph(graph: Graph, file: str):
    nodefile = open(file + "-nodes", 'wb')
    tilefile = open(file + "-tiles", 'wb')

    nodecount = len(graph.nodes)
    edgerefcount = len(graph.edgerefs)
    nodefile.write(struct.pack("i", nodecount))
    nodefile.write(struct.pack("i", edgerefcount))
    for node in graph.nodes:
        nodefile.write(struct.pack("b", 0))
        nodefile.write(struct.pack("i", node.start))
        nodefile.write(struct.pack("h", node.count))
        tilefile.write(struct.pack("h", node.tile))
    for edgref in graph.edgerefs:
        nodefile.write(struct.pack("i", edgref.edge_id))
        nodefile.write(struct.pack("b", edgref.typ))

    nodefile.close()
    tilefile.close()

class Landkreis:
    __slots__ = ['feature', 'extend', 'id']
    def __init__(self, feature: Polygon, id: int):
        self.feature = feature
        self.extend = feature.bounds
        self.id = id

    def in_extend(self, x: float, y: float) -> bool:
        return x > self.extend[0] and x < self.extend[2] and y > self.extend[1] and y < self.extend[3]
    
    def in_polygon(self, x, y):
        return contains_xy(self.feature, x, y)
    

if __name__ == "__main__":
    polygons: list[Landkreis] = []

    with open("./data/landkreise.json", "r") as file:
        featurecollection = json.loads(file.read())
        for feature in featurecollection["features"]:
            if feature["geometry"]["type"] == "Polygon":
                coords = feature["geometry"]["coordinates"]
                poly = Polygon(coords[0], coords[1:])
            else:
                coords = feature["geometry"]["coordinates"]
                rings = []
                for ring in coords:
                    rings.append((ring[0], ring[1:]))
                poly = MultiPolygon(rings)
            polygons.append(Landkreis(poly, int(feature["properties"]["TileID"])))

    graph: Graph = load_graph("data/niedersachsen")

    print("finished loading graph and polygons")

    c = 0
    for node in graph.nodes:
        if c % 1000 == 0:
            print(f"finished node {c}")
        x: float = node.lon
        y: float = node.lat
        for poly in polygons:
            if not poly.in_extend(x, y):
                continue
            if poly.in_polygon(x, y):
                node.tile = poly.id
        c += 1

    for edgeref in graph.edgerefs:
        if edgeref.typ != 0 and edgeref.typ != 1:
            continue
        edge = graph.edges[edgeref.edge_id]
        node_a = graph.nodes[edge.node_a]
        node_b = graph.nodes[edge.node_b]
        if node_a.tile != node_b.tile:
            edgeref.typ += 10

    store_graph(graph, "data/niedersachsen")
