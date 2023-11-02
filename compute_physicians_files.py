import struct
import json
from shapely.predicates import contains_xy
from shapely import Polygon, MultiPolygon

class DSPoint:
    __slots__ = ['x', 'y', 'weight']
    def __init__(self, x: float, y: float, weight: int):
        self.x = x
        self.y = y
        self.weight = weight

class Landkreis:
    __slots__ = ['feature', 'extend', 'name', 'haus_points', 'augen_points', 'frauen_points', 'haut_points', 'hno_points', 'ortho_points']
    def __init__(self, feature: Polygon, name: str):
        self.feature = feature
        self.extend = feature.bounds
        self.name = name
        self.haus_points: list[DSPoint] = []
        self.augen_points: list[DSPoint] = []
        self.frauen_points: list[DSPoint] = []
        self.haut_points: list[DSPoint] = []
        self.hno_points: list[DSPoint] = []
        self.ortho_points: list[DSPoint] = []

    def in_extend(self, x: float, y: float) -> bool:
        return x > self.extend[0] and x < self.extend[2] and y > self.extend[1] and y < self.extend[3]
    
    def in_polygon(self, x, y):
        return contains_xy(self.feature, x, y)

    def add_point(self, point: DSPoint, group: str):
        match group:
            case "general physician":
                self.haus_points.append(point)
            case "ophthalmologist":
                self.augen_points.append(point)
            case "gynaecologist":
                self.frauen_points.append(point)
            case "dermatologist":
                self.haut_points.append(point)
            case "otolaryngologist":
                self.hno_points.append(point)
            case "surgeon":
                self.ortho_points.append(point)

    def write_files(self, folder: str):
        name = self.name.lower().replace(" ", "_")
        files = [
            ("hausarzt", self.haus_points),
            ("augenarzt", self.augen_points),
            ("frauenarzt", self.frauen_points),
            ("hautarzt", self.haut_points),
            ("hno_arzt", self.hno_points),
            ("orthopade", self.ortho_points),
        ]
        for physician_name, physician_data in files:
            with open(folder + "/physicians_" + name + "_" + physician_name + ".txt", "w") as file:
                lines = [f"{p.x} {p.y} {p.weight}\n" for p in physician_data]
                file.writelines(lines)


if __name__ == "__main__":
    polygons: list[Landkreis] = []
    polygons.append(Landkreis(Polygon([(0, 0), (100, 0), (100, 100), (0, 100)]), "niedersachsen"))
    print("start reading landkreise...")
    with open("./data/landkreise_puffer_clip.json", "r") as file:
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
            polygons.append(Landkreis(poly, feature["properties"]["Name"]))

    # read physician points
    print("start reading physicians...")
    c = 0
    with open("./data/physicians_niedersachsen.json", "r") as file:
        featurecollection = json.loads(file.read())
        for feature in featurecollection["features"]:
            c += 1
            if c%1000 == 0:
                print(f"iteration {c}")
            x, y = feature["geometry"]["coordinates"]
            group = feature["properties"]["NAME_EN"]
            weight = 1
            point = DSPoint(x, y, weight)
            for poly in polygons:
                if not poly.in_extend(x, y):
                    continue
                if poly.in_polygon(x, y):
                    poly.add_point(point, group)

    print("start writing results...")
    for poly in polygons:
        poly.write_files("./data/physicians_landkreise")
    print("finished successfully")