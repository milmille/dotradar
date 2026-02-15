package internal

import "github.com/gdamore/tcell/v2"

type Canvas struct {
	Cells       *[][]cell
	XMultiplier int
	YMultiplier int
	screen      tcell.Screen
}

type cell struct {
	character uint8
	style     tcell.Style
}

func (c *Canvas) PaintPixel(x int, y int, style tcell.Style) {
	xCell := x / c.XMultiplier
	yCell := y / c.YMultiplier
	xMod := x % c.XMultiplier
	yMod := y % c.YMultiplier
	if style != tcell.StyleDefault {
		mask := uint8(1) << uint8((yMod*c.XMultiplier)+xMod)
		(*c.Cells)[xCell][yCell].character = (*c.Cells)[xCell][yCell].character | mask
		(*c.Cells)[xCell][yCell].style = style
	}
}

func (c *Canvas) Draw() {
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)
	for x := 0; x < len(*c.Cells); x++ {
		for y := 0; y < len((*c.Cells)[x]); y++ {
			if x == 0 && y == 0 {
				c.screen.SetContent(x, y, 0x250c, nil, borderStyle)
			} else if x == len(*c.Cells)-1 && y == 0 {
				c.screen.SetContent(x, y, 0x2510, nil, borderStyle)
			} else if x == 0 && y == len((*c.Cells)[x])-1 {
				c.screen.SetContent(x, y, 0x2514, nil, borderStyle)
			} else if x == len(*c.Cells)-1 && y == len((*c.Cells)[x])-1 {
				c.screen.SetContent(x, y, 0x2518, nil, borderStyle)
			} else if x == 0 || x == len(*c.Cells)-1 {
				c.screen.SetContent(x, y, 0x2502, nil, borderStyle)
			} else if y == 0 || y == len((*c.Cells)[x])-1 {
				c.screen.SetContent(x, y, 0x2500, nil, borderStyle)
			} else {
				pixel := (*c.Cells)[x][y]
				rune, _, _, _ := c.screen.GetContent(x, y)
				if rune == ' ' {
					c.screen.SetContent(x, y, OctantRunes[pixel.character], nil, pixel.style)
				}
			}
		}
	}
}

func (c *Canvas) Clear() {
	width := len(*c.Cells)
	height := len((*c.Cells)[0])
	cells := make([][]cell, width)
	for i := range cells {
		cells[i] = make([]cell, height)
	}
	c.Cells = &cells
}

// Generate a 2d slice of uint8, each representing a cell
// given the size of the tcell screen
func NewCanvas(screen tcell.Screen, width, height, xMult, yMult int) *Canvas {
	canvas := make([][]cell, width)
	for i := range canvas {
		canvas[i] = make([]cell, height)
	}
	return &Canvas{
		screen:      screen,
		Cells:       &canvas,
		XMultiplier: xMult,
		YMultiplier: yMult,
	}
}
