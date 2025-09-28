/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/planar"
	"github.com/spf13/cobra"
)

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

	michigan := getFeature("Michigan", fc).Geometry.(orb.MultiPolygon)
	centroid, _ := planar.CentroidArea(michigan[0])
	tile := maptile.At(centroid, 50)
	fmt.Println(tile.Bound().Intersects(michigan[0].Bound()))

}

func getFeature(name string, fc *geojson.FeatureCollection) *geojson.Feature {

	for _, feature := range fc.Features {
		if feature.Properties.MustString("NAME") == name {
			return feature
		}
	}
	return nil
}
