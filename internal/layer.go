package internal

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/paulmach/orb"
)

type Pixel struct {
	character uint8
	style     tcell.Style
}

type Layer struct {
	Pixels      *[][]Pixel
	XMultiplier int
	YMultiplier int
	screen      tcell.Screen
}

const LOWER_HALF_BLOCK = 0b11110000

func (l *Layer) PaintPixel(x int, y int, style tcell.Style) {
	xCell := x / l.XMultiplier
	yCell := y / l.YMultiplier
	xMod := x % l.XMultiplier
	yMod := y % l.YMultiplier
	if style != tcell.StyleDefault {
		mask := uint8(1) << uint8((yMod*l.XMultiplier)+xMod)
		(*l.Pixels)[xCell][yCell].character = (*l.Pixels)[xCell][yCell].character | mask
		(*l.Pixels)[xCell][yCell].style = style
	}
}

func (l *Layer) Draw() {
	for x := range *l.Pixels {
		for y := range (*l.Pixels)[x] {
			pixel := (*l.Pixels)[x][y]
			rune, _, _, _ := l.screen.GetContent(x, y)
			if rune == ' ' {
				l.screen.SetContent(x, y, OctantRunes[pixel.character], nil, pixel.style)
			}
		}
	}
}

func (l *Layer) PaintPixelFromImage(x int, y int, color tcell.Color) {
	xCell := x / l.XMultiplier
	yCell := y / l.YMultiplier
	yMod := y % l.YMultiplier
	pixel := (*l.Pixels)[xCell][yCell]
	pixel.character = LOWER_HALF_BLOCK
	if yMod == 0 {
		pixel.style = pixel.style.Background(color)
	} else {
		pixel.style = pixel.style.Foreground(color)
	}
	(*l.Pixels)[xCell][yCell] = pixel
}

func (l *Layer) PaintLine(x1, y1, x2, y2 int, style tcell.Style) {
	if x1 == x2 {
		// vertical line case
		if y1 > y2 {
			//swap ys
			yTemp := y2
			y2 = y1
			y1 = yTemp
		}
		for y := y1; y < y2; y++ {
			l.PaintPixel(x1-1, y, style)
		}
	} else {
		if x1 > x2 {
			//swap points
			xTemp, yTemp := x2, y2
			x2, y2 = x1, y1
			x1, y1 = xTemp, yTemp
		}
		slope := float64(y2-y1) / float64(x2-x1)
		for x := x1; x < x2; x++ {
			y := math.Round(slope*float64(x-x1)+float64(y1)) - 1
			l.PaintPixel(x, int(y), style)
		}
	}
}

// draw the polygon on the screen assuming in screen coordinates
func (l *Layer) DrawPolygon(multiPolygon orb.MultiPolygon, style tcell.Style) {
	for _, polygon := range multiPolygon {
		for _, ring := range polygon {
			for j := 0; j < len(ring)-1; j++ {
				point1 := ring[j]
				point2 := ring[j+1]
				l.PaintLine(int(point1[0]), int(point1[1]), int(point2[0]), int(point2[1]), style)
			}
		}
	}
}

// Generate a 2d slice of uint8, each representing a cell
// given the size of the tcell screen
func NewPixelSlice(width, height int) *[][]Pixel {
	pixels := make([][]Pixel, width)
	for i := range pixels {
		pixels[i] = make([]Pixel, height)
	}
	return &pixels
}
