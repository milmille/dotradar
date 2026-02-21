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

const RADAR_X_MULT = 1
const RADAR_Y_MULT = 2

const LOWER_HALF_BLOCK = 0b11110000

const NOAA_URL = "https://opengeo.ncep.noaa.gov/geoserver/conus/conus_bref_qcd/ows?"

var baseUrlParams = url.Values{
	"service":     []string{"wms"},
	"request":     []string{"GetMap"},
	"layers":      []string{"conus_bref_qcd"},
	"format":      []string{"image/png"},
	"CRS":         []string{"EPSG:3857"}, // web mercator
	"transparent": []string{"true"},
}

type radarLayerImpl struct {
	canvas *Canvas
}

func NewRadarLayer(screen tcell.Screen, container *Container) Layer {
	return &radarLayerImpl{
		canvas: NewCanvas(screen, container, RADAR_X_MULT, RADAR_Y_MULT),
	}
}

func (rl *radarLayerImpl) Render(bound orb.Bound, container *Container) {
	var width, height int
	if container == nil {
		width, height = rl.canvas.screen.Size()
	} else {
		width, height = container.Width, container.Height
	}
	image := rl.getMap(bound, width, height)
	rl.drawRadarImage(image)
	rl.canvas.Draw()
}

func (rl *radarLayerImpl) Clear() {
	rl.canvas.Clear()
}

func (rl *radarLayerImpl) drawRadarImage(image image.Image) {
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			newColor := color.NRGBAModel.Convert(image.At(x, y)).(color.NRGBA)
			if newColor.A != 0 {
				rl.paintPixelFromImage(x, y, tcell.NewRGBColor(int32(newColor.R), int32(newColor.B), int32(newColor.G)))
			}
		}
	}
}

func (rl *radarLayerImpl) paintPixelFromImage(x int, y int, color tcell.Color) {
	xCell := x / rl.canvas.XMultiplier
	yCell := y / rl.canvas.YMultiplier
	yMod := y % rl.canvas.YMultiplier
	pixel := (*rl.canvas.Cells)[xCell][yCell]
	pixel.character = LOWER_HALF_BLOCK
	if yMod == 0 {
		pixel.style = pixel.style.Background(color)
	} else {
		pixel.style = pixel.style.Foreground(color)
	}
	(*rl.canvas.Cells)[xCell][yCell] = pixel
}

func (rl *radarLayerImpl) getMap(bound orb.Bound, screenWidth, screenHeight int) image.Image {

	urlParams := maps.Clone(baseUrlParams)

	widthStr := fmt.Sprintf("%d", screenWidth*rl.canvas.XMultiplier)
	heightStr := fmt.Sprintf("%d", screenHeight*rl.canvas.YMultiplier)
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
