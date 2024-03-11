import struct
import json

class Edge:
    __slots__ = ['node_a', 'node_b', 'typ', 'coords']
    def __init__(self, node_a: int, node_b: int):
        self.node_a: int = node_a
        self.node_b: int = node_b
        self.typ: int = 0
        self.coords: list = None

class EdgeRef:
    __slots__ = ['edge_id']
    def __init__(self, edge_id: int):
        self.edge_id: int = edge_id

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
    tilefile = open(file + "-tiles", 'rb')

    edgecount = struct.unpack("i", edgefile.read(4))[0]
    for i in range(edgecount):
        node_a = struct.unpack("i", edgefile.read(4))[0]
        node_b = struct.unpack("i", edgefile.read(4))[0]
        w = struct.unpack("b", edgefile.read(1))[0]
        t = struct.unpack("b", edgefile.read(1))[0]
        l = struct.unpack("f", edgefile.read(4))[0]
        m = struct.unpack("b", edgefile.read(1))[0]
        edges.append(Edge(node_a, node_b))

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
        edgerefs.append(EdgeRef(edge_id))
        edges[edge_id].typ = typ
    
    for i in range(nodecount):
        tile = struct.unpack("h", tilefile.read(2))[0]
        nodes[i].tile = tile
    for i in range(nodecount):
        lon = struct.unpack("f", geomfile.read(4))[0]
        lat = struct.unpack("f", geomfile.read(4))[0]
        nodes[i].lon = lon
        nodes[i].lat = lat
    for i in range(edgecount):
        start = struct.unpack("i", geomfile.read(4))[0]
        count = struct.unpack("B", geomfile.read(1))[0]
        edges[i].coords = count
    for edge in edges:
        count = edge.coords
        coords = []
        for i in range(count):
            lon = struct.unpack("f", geomfile.read(4))[0]
            lat = struct.unpack("f", geomfile.read(4))[0]
            coords.append((lon, lat))
        edge.coords = coords

    nodefile.close()
    edgefile.close()
    geomfile.close()

    return Graph(nodes, edges, edgerefs)

class GeoJson():
    def __init__(self):
        self.type = "FeatureCollection"
        self.features = []

    def add_node(self, index, lon, lat, tile, edgecount, edges):
        feature = {}
        feature["type"] = "Feature"
        feature["geometry"] = {"type": "Point", "coordinates": [lon, lat]}
        feature["properties"] = {"index": index, "edgecount": edgecount, "edges": edges, "tile": tile}
        self.features.append(feature)

    def add_edge(self, index, nodeA, nodeB, typ, coords):
        feature = {}
        feature["type"] = "Feature"
        feature["geometry"] = {"type": "LineString", "coordinates": coords}
        feature["properties"] = {"index": index, "nodeA": nodeA, "nodeB": nodeB, "type": typ}
        self.features.append(feature)

if __name__ == '__main__':
    graph = load_graph("./data/default")

    # edges = GeoJson()
    # for i, edge in enumerate(graph.edges):
    #     if i%1000 == 0:
    #         print(f"edge {i}")
    #     edges.add_edge(i, edge.node_a, edge.node_b, edge.typ, edge.coords)
    # with open('edges.json', 'w') as f:
    #     json.dump(edges.__dict__, f)

    nodes = GeoJson()
    for i, node in enumerate(graph.nodes):
        if i%1000 == 0:
            print(f"node {i}")
        edges = []
        for j in range(node.count):
            edges.append(graph.edgerefs[node.start + j].edge_id)
        nodes.add_node(i, node.lon, node.lat, node.tile, node.count, edges)
    with open('nodes.json', 'w') as f:
        json.dump(nodes.__dict__, f)