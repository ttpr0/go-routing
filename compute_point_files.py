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
    __slots__ = ['feature', 'extend', 'name', 'sup_points', 'dem_points']
    def __init__(self, feature: Polygon, name: str):
        self.feature = feature
        self.extend = feature.bounds
        self.name = name
        self.sup_points: list[DSPoint] = []
        self.dem_points: list[DSPoint] = []

    def in_extend(self, x: float, y: float) -> bool:
        return x > self.extend[0] and x < self.extend[2] and y > self.extend[1] and y < self.extend[3]
    
    def in_polygon(self, x, y):
        return contains_xy(self.feature, x, y)

    def add_sup_point(self, point: DSPoint):
        self.sup_points.append(point)

    def add_dem_point(self, point: DSPoint):
        self.dem_points.append(point)

    def write_files(self, folder: str):
        name = self.name.lower().replace(" ", "_")
        with open(folder + "/population_" + name + ".txt", "w") as file:
            lines = [f"{p.x} {p.y} {p.weight}\n" for p in self.dem_points]
            file.writelines(lines)
        with open(folder + "/physicians_" + name + ".txt", "w") as file:
            lines = [f"{p.x} {p.y} {p.weight}\n" for p in self.sup_points]
            file.writelines(lines)

    def get_counts(self) -> tuple[int, int]:
        return len(self.dem_points), len(self.sup_points)


if __name__ == "__main__":
    polygons: list[Landkreis] = []
    polygons.append(Landkreis(Polygon([(0, 0), (100, 0), (100, 100), (0, 100)]), "niedersachsen"))
    print("start reading landkreise...")
    with open("./data/mittelbereiche_puffer_clip.json", "r") as file:
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

    # read population points
    print("start reading population...")
    c = 0
    with open("./data/population_niedersachsen.json", "r") as file:
        featurecollection = json.loads(file.read())
        for feature in featurecollection["features"]:
            c += 1
            if c%1000 == 0:
                print(f"iteration {c}")
            x, y = feature["geometry"]["coordinates"]
            weight = feature["properties"]["EW_GESAMT"]
            point = DSPoint(x, y, weight)
            for poly in polygons:
                if not poly.in_extend(x, y):
                    continue
                if poly.in_polygon(x, y):
                    poly.add_dem_point(point)

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
            weight = 1
            point = DSPoint(x, y, weight)
            for poly in polygons:
                if not poly.in_extend(x, y):
                    continue
                if poly.in_polygon(x, y):
                    poly.add_sup_point(point)

    print("start writing results...")
    for poly in polygons:
        poly.write_files("./data/points_mittelbereiche")
    print("finished successfully")

    counts = []
    for poly in polygons:
        d_c, s_c = poly.get_counts()
        counts.append((poly.name, d_c, s_c))
    counts_sorted = sorted(counts, key=lambda x: x[2])

    for count in counts_sorted:
        print(f"{count[0]}: {count[1]}  {count[2]}")