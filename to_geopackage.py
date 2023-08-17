import json
import fiona

driver = "GPKG"

nodes_schema = {
    'geometry': 'Point',
    'properties': dict([
        ('index', 'int'),
        ('edgecount', 'int'),
        ('tile', 'int')
    ])
}

with open("nodes.json", 'r') as file:
    data = json.loads(file.read())

with fiona.open("graph.gpkg", "w", driver=driver, schema=nodes_schema, layer="nodes") as file:
    features = data["features"]
    for feature in features:
        del feature["properties"]["edges"]
    file.writerecords(features)

edges_schema = {
    'geometry': 'LineString',
    'properties': dict([
        ('index', 'int'),
        ('nodeA', 'int'),
        ('nodeB', 'int'),
        ('type', 'int')
    ])
}

with open("edges.json", 'r') as file:
    data = json.loads(file.read())

with fiona.open("graph.gpkg", "w", driver=driver, schema=edges_schema, layer="edges") as file:
    features = data["features"]
    file.writerecords(features)
