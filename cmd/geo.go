/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/project"
	"github.com/spf13/cobra"
)

const width = 320 * 2
const height = 82 * 4

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "geo",
	Short: "testing out geography things",
	Run: func(cmd *cobra.Command, args []string) {
		unmarshal()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func unmarshal() {
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

	minnesota := getFeature("Minnesota", fc).Geometry.(orb.Polygon)
	topLeft := orb.Point{-101.54149625952375, 49.31848896184871}
	bottomRight := orb.Point{-88.52181788184106, 42.54514923415394}
	bound := orb.MultiPoint{topLeft, bottomRight}.Bound()
	mercMinn := project.Polygon(minnesota, project.WGS84.ToMercator)
	mercBound := project.Bound(bound, project.WGS84.ToMercator)
	xOffset := mercBound.Left()
	yOffset := mercBound.Top()
	xMax := mercBound.Right() + (xOffset * -1)
	yMax := math.Abs(mercBound.Bottom() + (yOffset * -1))

	translated := make([]orb.Ring, len(mercMinn))
	for i, ring := range mercMinn {
		translatedRing := make([]orb.Point, len(ring))
		for j, point := range ring {
			translatedX := point[0] + (xOffset * -1)
			translatedY := math.Abs((point[1] + (yOffset * -1)))
			normalizedX := math.Round((translatedX / xMax) * width)
			normalizedY := math.Round((translatedY / yMax) * height)
			translatedRing[j] = orb.Point{normalizedX, normalizedY}
		}
		translated[i] = translatedRing
	}

	for _, ring := range translated {
		for _, point := range ring {
			fmt.Printf("[%f, %f]\n", point[0], point[1])
		}
	}

}

func getFeature(name string, fc *geojson.FeatureCollection) *geojson.Feature {

	for _, feature := range fc.Features {
		if feature.Properties.MustString("NAME") == name {
			return feature
		}
	}
	return nil
}
