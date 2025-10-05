/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"image"
	"image/color"
	"net/url"

	"github.com/milmille/dotradar/internal"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/project"
	"github.com/spf13/cobra"
)

const SCREEN_WIDTH = 320
const SCREEN_HEIGHT = 82
const WIDTH_MULT = 1
const HEIGHT_MULT = 2

const NOAA_URL = "https://opengeo.ncep.noaa.gov/geoserver/conus/conus_bref_qcd/ows?"

var urlParams = url.Values{
	"service":     []string{"wms"},
	"request":     []string{"GetMap"},
	"layers":      []string{"conus_bref_qcd"},
	"format":      []string{"image/png"},
	"CRS":         []string{"EPSG:3857"}, // web mercator
	"transparent": []string{"true"},
}

// radarCmd represents the radar command
var radarCmd = &cobra.Command{
	Use:   "radar",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("radar called")
		state, err := cmd.Flags().GetString("state")
		if err != nil {
			fmt.Println("error getting state flat %v", err.Error())
		}
		image := getMap(state)
		drawImage(image)

	},
}

func init() {
	rootCmd.AddCommand(radarCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// radarCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// radarCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getMap(state string) image.Image {

	fc := internal.ReadGeoJSON("./gz_2010_us_040_00_20m.json")
	geometry := internal.GetFeature(state, fc).Geometry
	var centerState orb.MultiPolygon
	if multiPolygon, ok := geometry.(orb.MultiPolygon); ok {
		centerState = multiPolygon
	} else if polygon, ok := geometry.(orb.Polygon); ok {
		centerState = orb.MultiPolygon{polygon}
	}
	centerStateMerc := project.MultiPolygon(centerState.Clone(), project.WGS84.ToMercator)

	return internal.GetMap(centerStateMerc.Clone(), SCREEN_WIDTH, SCREEN_HEIGHT)
}

func drawImage(image image.Image) {
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			oldColor := image.At(x, y)
			newColor := color.NRGBAModel.Convert(oldColor).(color.NRGBA)
			_, _, _, alpha := newColor.RGBA()
			if alpha != 0 {
				fmt.Printf("drawing color (%d, %d, %d) at %d, %d\n", newColor.R, newColor.G, newColor.B, x, y)
			} else {
				fmt.Println("transparent")
			}
		}
	}
}
