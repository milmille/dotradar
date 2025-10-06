package internal

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"maps"
	"net/http"
	"net/url"

	"github.com/gdamore/tcell/v2"

	"github.com/paulmach/orb"
)

const WIDTH_MULT = 1
const HEIGHT_MULT = 2

const NOAA_URL = "https://opengeo.ncep.noaa.gov/geoserver/conus/conus_bref_qcd/ows?"

var baseUrlParams = url.Values{
	"service":     []string{"wms"},
	"request":     []string{"GetMap"},
	"layers":      []string{"conus_bref_qcd"},
	"format":      []string{"image/png"},
	"CRS":         []string{"EPSG:3857"}, // web mercator
	"transparent": []string{"true"},
}

func GetMap(bound orb.Bound, screenWidth, screenHeight int) image.Image {

	urlParams := maps.Clone(baseUrlParams)

	widthStr := fmt.Sprintf("%d", screenWidth*WIDTH_MULT)
	heightStr := fmt.Sprintf("%d", screenHeight*HEIGHT_MULT)
	urlParams.Add("width", widthStr)
	urlParams.Add("height", heightStr)

	bottom := fmt.Sprintf("%f", bound.Bottom())
	left := fmt.Sprintf("%f", bound.Left())
	top := fmt.Sprintf("%f", bound.Top())
	right := fmt.Sprintf("%f", bound.Right())

	urlParams.Add("bbox", left+","+bottom+","+right+","+top)

	url := NOAA_URL + urlParams.Encode()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error sending request to noaa %v", err.Error())
	}
	if resp.Header.Get("content-type") != "image/png" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("error reading response: %s", err.Error())
		}
		fmt.Println(string(body))
		return nil
	}

	img, err := png.Decode(resp.Body)
	if err != nil {
		fmt.Println("error decoding image: %s", err.Error())
	}
	return img
}

func drawImage(layer *Layer, image image.Image) {
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			newColor := color.NRGBAModel.Convert(image.At(x, y)).(color.NRGBA)
			if newColor.A != 0 {
				layer.PaintPixelFromImage(x, y, tcell.NewRGBColor(int32(newColor.R), int32(newColor.B), int32(newColor.G)))
			}
		}
	}
}
