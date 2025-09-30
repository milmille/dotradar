package internal

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/paulmach/orb"
)

const X_MULT = 2
const Y_MULT = 4

type Layer struct {
	Pixels *[][]uint8
	screen tcell.Screen
}

func (l *Layer) PaintPixel(x int, y int) {
	xCell := x / X_MULT
	yCell := y / Y_MULT
	xMod := x % X_MULT
	yMod := y % Y_MULT
	mask := uint8(1) << uint8((yMod*X_MULT)+xMod)
	(*l.Pixels)[xCell][yCell] = (*l.Pixels)[xCell][yCell] | mask
}

func (l *Layer) Draw(style tcell.Style) {
	for x := range *l.Pixels {
		for y := range (*l.Pixels)[x] {
			l.screen.SetContent(x, y, OctantRunes[(*l.Pixels)[x][y]], nil, style)
		}
	}
}

func (l *Layer) PaintLine(x1, y1, x2, y2 int) {
	if x1 == x2 {
		// vertical line case
		if y1 > y2 {
			//swap ys
			yTemp := y2
			y2 = y1
			y1 = yTemp
		}
		for y := y1; y < y2; y++ {
			l.PaintPixel(x1-1, y)
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
			l.PaintPixel(x, int(y))
		}
	}
}

// draw the polygon on the screen assuming in screen coordinates
func (l *Layer) DrawPolygon(multiPolygon orb.MultiPolygon) {
	for _, polygon := range multiPolygon {
		for _, ring := range polygon {
			for j := 0; j < len(ring)-1; j++ {
				point1 := ring[j]
				point2 := ring[j+1]
				l.PaintLine(int(point1[0]), int(point1[1]), int(point2[0]), int(point2[1]))
			}
		}
	}
}

// Generate a 2d slice of uint8, each representing a cell
// given the size of the tcell screen
func NewPixelSlice(width, height int) *[][]uint8 {
	pixels := make([][]uint8, width)
	for i := range pixels {
		pixels[i] = make([]uint8, height)
	}
	return &pixels
}
