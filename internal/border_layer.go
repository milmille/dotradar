package internal

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/project"
)

type borderLayerImpl struct {
	canvas   *Canvas
	features []*geojson.Feature
}

func NewBorderLayer(screen tcell.Screen, features []*geojson.Feature, width, height int) Layer {
	return &borderLayerImpl{
		canvas:   NewCanvas(screen, width, height, 2, 4),
		features: features,
	}
}

func (bl *borderLayerImpl) Render(centerPoint orb.Point, zoom, width, height int) {
	drawStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)

	//TODO: need to abstract away the x/y mults in findBound
	bound := FindBound(centerPoint, width*bl.canvas.XMultiplier, height*bl.canvas.YMultiplier, zoom)
	for _, feature := range bl.features {
		var state orb.MultiPolygon
		if multiPolygon, ok := feature.Geometry.(orb.MultiPolygon); ok {
			state = multiPolygon
		} else if polygon, ok := feature.Geometry.(orb.Polygon); ok {
			state = orb.MultiPolygon{polygon}
		}
		stateMerc := project.MultiPolygon(state.Clone(), project.WGS84.ToMercator)

		stateClipped := clip.MultiPolygon(bound, stateMerc)
		if !stateClipped.Bound().IsEmpty() {
			stateFit := FitToScreen(stateClipped, bound, width*2, height*4)
			bl.drawPolygon(stateFit, drawStyle)
		}
	}
	bl.canvas.Draw()
}

func (bl *borderLayerImpl) Clear() {
	bl.canvas.Clear()
}

func (bl *borderLayerImpl) drawPolygon(multiPolygon orb.MultiPolygon, style tcell.Style) {

	for _, polygon := range multiPolygon {
		for _, ring := range polygon {
			for j := 0; j < len(ring)-1; j++ {
				point1 := ring[j]
				point2 := ring[j+1]
				bl.paintLine(int(point1[0]), int(point1[1]), int(point2[0]), int(point2[1]), style)
			}
		}
	}
}

func (bl *borderLayerImpl) paintLine(x1, y1, x2, y2 int, style tcell.Style) {
	if x1 == x2 {
		// vertical line case
		if y1 > y2 {
			//swap ys
			yTemp := y2
			y2 = y1
			y1 = yTemp
		}
		for y := y1; y < y2; y++ {
			bl.canvas.PaintPixel(x1-1, y, style)
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
			bl.canvas.PaintPixel(x, int(y), style)
		}
	}
}
