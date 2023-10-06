import json

if __name__ == "__main__":
    features = []
    with open("./data/point_files/population_aurich.txt", "r") as file:
        for line in file.readlines():
            tokens = line.split(" ")
            features.append({
                "type": "Feature",
                "geometry": {
                    "type": "Point",
                    "coordinates": [float(tokens[0]), float(tokens[1])] 
                },
                "properties": {
                    "weight": int(tokens[2])
                }
            })
    with open("./data/test/population_aurich.json", "w") as file:
        file.write(json.dumps({
            "type": "FeatureCollection",
            "features": features
        }))