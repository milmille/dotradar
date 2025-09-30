package internal

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
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
func FitToScreen(multiPolygon orb.MultiPolygon, bound orb.Bound, width int, height int) orb.MultiPolygon {

	xOffset := bound.Left()
	yOffset := bound.Top()
	xMax := bound.Right() + (xOffset * -1)
	yMax := math.Abs(bound.Bottom() + (yOffset * -1))

	translatedMulti := make([]orb.Polygon, len(multiPolygon))
	for i, singlePolygon := range multiPolygon {
		translated := make([]orb.Ring, len(singlePolygon))
		for j, ring := range singlePolygon {
			translatedRing := make([]orb.Point, len(ring))
			for k, point := range ring {
				translatedX := point[0] + (xOffset * -1)
				// need to take the abs because the x direction is flipped
				translatedY := math.Abs((point[1] + (yOffset * -1)))
				normalizedX := math.Round((translatedX / xMax) * float64(width))
				normalizedY := math.Round((translatedY / yMax) * float64(height))
				translatedRing[k] = orb.Point{normalizedX, normalizedY}
			}
			translated[j] = translatedRing
		}
		translatedMulti[i] = translated
	}

	return translatedMulti
}

func FindBound(multiPolygon orb.MultiPolygon, width, height, zoomFactor int) orb.Bound {
	centroid, _ := planar.CentroidArea(multiPolygon)
	left := centroid[0] - float64(width*zoomFactor/2)
	top := centroid[1] - float64(height*zoomFactor/2)
	right := centroid[0] + float64(width*zoomFactor/2)
	bottom := centroid[1] + float64(height*zoomFactor/2)

	return orb.MultiPoint{orb.Point{left, top}, orb.Point{right, bottom}}.Bound()
}

func GetFeature(name string, fc *geojson.FeatureCollection) *geojson.Feature {
	searchStr := strings.ToLower(name)
	for _, feature := range fc.Features {
		if strings.ToLower(feature.Properties.MustString("NAME")) == searchStr {
			return feature
		}
	}
	return nil
}
