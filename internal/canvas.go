package internal

import (
	"github.com/gdamore/tcell/v2"
)

const (
	BOX_TOP_LEFT     = 0x256d
	BOX_TOP_RIGHT    = 0x256e
	BOX_BOTTOM_LEFT  = 0x2570
	BOX_BOTTOM_RIGHT = 0x256f
	BOX_VERTICAL     = 0x2502
	BOX_HORIZONTAL   = 0x2500
)

type Canvas struct {
	Cells       *[][]cell
	XMultiplier int
	YMultiplier int
	screen      tcell.Screen
	container   *Container
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
	width, height := c.screen.Size()
	var xMin, xMax, yMin, yMax int
	if c.container == nil {
		xMin, yMin = 0, 0
		xMax = width
		yMax = height
	} else {
		// pin the container to the top, center horizontally
		xMin = (width - c.container.Width) / 2
		xMax = len(*c.Cells) + xMin
		yMin = 1
		yMax = len((*c.Cells)[0])
	}
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)
	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			if c.container != nil {
				if x == xMin && y == yMin {
					// top left
					c.screen.SetContent(x, y, BOX_TOP_LEFT, nil, borderStyle)
				} else if x == xMax-1 && y == yMin {
					// top right
					c.screen.SetContent(x, y, BOX_TOP_RIGHT, nil, borderStyle)
				} else if x == xMin && y == yMax-1 {
					// bottom left
					c.screen.SetContent(x, y, BOX_BOTTOM_LEFT, nil, borderStyle)
				} else if x == xMax-1 && y == yMax-1 {
					// bottom right
					c.screen.SetContent(x, y, BOX_BOTTOM_RIGHT, nil, borderStyle)
				} else if x == xMin || x == xMax-1 {
					// sides
					c.screen.SetContent(x, y, BOX_VERTICAL, nil, borderStyle)
				} else if y == yMin || y == yMax-1 {
					// top bottom
					c.screen.SetContent(x, y, BOX_HORIZONTAL, nil, borderStyle)
				} else {
					pixel := (*c.Cells)[x-xMin][y-yMin]
					rune, _, _, _ := c.screen.GetContent(x, y)
					if rune == ' ' {
						c.screen.SetContent(x, y, OctantRunes[pixel.character], nil, pixel.style)
					}
				}
			} else {
				pixel := (*c.Cells)[x-xMin][y-yMin]
				rune, _, _, _ := c.screen.GetContent(x, y)
				if rune == ' ' {
					c.screen.SetContent(x, y, OctantRunes[pixel.character], nil, pixel.style)
				}
			}
		}
	}
	c.DrawToolbar()
}

func (c *Canvas) DrawToolbar() {
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)
	width, height := c.screen.Size()
	if c.container == nil {
		return
	}
	xMin := (width - c.container.Width) / 2
	xMax := len(*c.Cells) + xMin
	yMin := len((*c.Cells)[0])
	yMax := height

	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			if x == xMin && y == yMin {
				// top left
				c.screen.SetContent(x, y, BOX_TOP_LEFT, nil, borderStyle)
			} else if x == xMax-1 && y == yMin {
				// top right
				c.screen.SetContent(x, y, BOX_TOP_RIGHT, nil, borderStyle)
			} else if x == xMin && y == yMax-1 {
				// bottom left
				c.screen.SetContent(x, y, BOX_BOTTOM_LEFT, nil, borderStyle)
			} else if x == xMax-1 && y == yMax-1 {
				// bottom right
				c.screen.SetContent(x, y, BOX_BOTTOM_RIGHT, nil, borderStyle)
			} else if x == xMin || x == xMax-1 {
				// sides
				c.screen.SetContent(x, y, BOX_VERTICAL, nil, borderStyle)
			} else if y == yMin || y == yMax-1 {
				// top bottom
				c.screen.SetContent(x, y, BOX_HORIZONTAL, nil, borderStyle)
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
	clearStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)
	width, height = c.screen.Size()
	var xMin, xMax, yMin, yMax int
	if c.container == nil {
		xMin, yMin = 0, 0
		xMax = width
		yMax = height
	} else {
		// pin the container to the top, center horizontally
		xMin = (width - c.container.Width) / 2
		xMax = len(*c.Cells) + xMin
		yMin = 1
		yMax = len((*c.Cells)[0])
	}
	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			c.screen.SetContent(x, y, ' ', nil, clearStyle)
		}
	}
}

// Generate a 2d slice of uint8, each representing a cell
func NewCanvas(screen tcell.Screen, container *Container, xMult, yMult int) *Canvas {
	var width, height int
	if container == nil {
		width, height = screen.Size()
	} else {
		width, height = container.Width, container.Height
	}
	canvas := make([][]cell, width)
	for i := range canvas {
		canvas[i] = make([]cell, height)
	}
	return &Canvas{
		screen:      screen,
		Cells:       &canvas,
		XMultiplier: xMult,
		YMultiplier: yMult,
		container:   container,
	}
}
