package internal

import (
	"encoding/json"
	"log"
	"math"
	"os"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

// read in json file and return feature collection
func ReadGeoJSON(path string) *geojson.FeatureCollection {
	filePath := "./gz_2010_us_040_00_20m.json" // Replace with your file path

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	fc := geojson.NewFeatureCollection()
	err = json.Unmarshal(content, &fc)
	if err != nil {
		log.Fatalf("Error unmarshaling json: %v", err)
	}
	return fc
}

// assuming mercator projection, fit to the polygon to the screen
func FitToScreen(polygon orb.Polygon, bound orb.Bound, width int, height int) orb.Polygon {

	xOffset := bound.Left()
	yOffset := bound.Top()
	xMax := bound.Right() + (xOffset * -1)
	yMax := math.Abs(bound.Bottom() + (yOffset * -1))

	translated := make([]orb.Ring, len(polygon))
	for i, ring := range polygon {
		translatedRing := make([]orb.Point, len(ring))
		for j, point := range ring {
			translatedX := point[0] + (xOffset * -1)
			// need to take the abs because the x direction is flipped
			translatedY := math.Abs((point[1] + (yOffset * -1)))
			normalizedX := math.Round((translatedX / xMax) * float64(width))
			normalizedY := math.Round((translatedY / yMax) * float64(height))
			translatedRing[j] = orb.Point{normalizedX, normalizedY}
		}
		translated[i] = translatedRing
	}

	return translated
}

func GetFeature(name string, fc *geojson.FeatureCollection) *geojson.Feature {

	for _, feature := range fc.Features {
		if feature.Properties.MustString("NAME") == name {
			return feature
		}
	}
	return nil
}
