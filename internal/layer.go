package internal

import (
	"github.com/gdamore/tcell/v2"
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

func (l *Layer) PainLine(x1, y1, x2, y2 int) {
	if x1 == x2 {
		// vertical line case
		if y1 > y2 {
			//swap ys
			yTemp := y2
			y2 = y1
			y1 = yTemp
		}
		for y := y1; y <= y2; y++ {
			l.PaintPixel(x1, y)
		}
	} else {
		if x1 > x2 {
			//swap points
			xTemp, yTemp := x2, y2
			x2, y2 = x1, y1
			x1, y1 = xTemp, yTemp
		}
		slope := (y2 - y1) / (x2 - x1)
		for x := x1; x <= x2; x++ {
			y := slope*(x-x1) + y1
			l.PaintPixel(x, y)
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
