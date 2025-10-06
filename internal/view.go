package internal

import (
	"log"
	"time"

	"github.com/bep/debounce"
	"github.com/gdamore/tcell/v2"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/paulmach/orb/project"
)

func View(stateStr string) {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	quit := func() {
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	width, height := s.Size()
	pixels := NewPixelSlice(width, height)
	layer := Layer{Pixels: pixels, XMultiplier: 2, YMultiplier: 4, screen: s}

	fc := ReadGeoJSON("./gz_2010_us_040_00_20m.json")
	geometry := GetFeature(stateStr, fc).Geometry
	var centerState orb.MultiPolygon
	if multiPolygon, ok := geometry.(orb.MultiPolygon); ok {
		centerState = multiPolygon
	} else if polygon, ok := geometry.(orb.Polygon); ok {
		centerState = orb.MultiPolygon{polygon}
	}
	centerStateMerc := project.MultiPolygon(centerState.Clone(), project.WGS84.ToMercator)
	startingCenter, _ := planar.CentroidArea(centerStateMerc)

	imagePixels := NewPixelSlice(width, height)
	imageLayer := Layer{Pixels: imagePixels, XMultiplier: 1, YMultiplier: 2, screen: s}

	debounced := debounce.New(1000 * time.Millisecond)
	xOffset := 0.0
	yOffset := 0.0
	zoomOffset := 0

	renderRadar(startingCenter, xOffset, yOffset, zoomOffset, width, height, &imageLayer)
	renderBorders(startingCenter, xOffset, yOffset, zoomOffset, width, height, fc.Features, &layer)
	imageLayer.Draw()
	layer.Draw()
	s.Show()
	for {
		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				// quit
				return
			} else if ev.Rune() == 'c' || ev.Rune() == 'C' {
				// clear
				s.Clear()
			} else if ev.Rune() == 'd' || ev.Rune() == 'D' {
				// zoom in
				s.Clear()
				layer.Clear()
				imageLayer.Clear()
				zoomOffset -= 1000
				renderBorders(startingCenter, xOffset, yOffset, zoomOffset, width, height, fc.Features, &layer)
				f := func() {
					s.Clear()
					renderRadar(startingCenter, xOffset, yOffset, zoomOffset, width, height, &imageLayer)
					imageLayer.Draw()
					layer.Draw()
					s.Show()
				}
				debounced(f)
				imageLayer.Draw()
				layer.Draw()
				s.Show()
			} else if ev.Rune() == 'u' || ev.Rune() == 'U' {
				// zoom out
				s.Clear()
				layer.Clear()
				imageLayer.Clear()
				zoomOffset += 1000
				renderBorders(startingCenter, xOffset, yOffset, zoomOffset, width, height, fc.Features, &layer)
				f := func() {
					s.Clear()
					renderRadar(startingCenter, xOffset, yOffset, zoomOffset, width, height, &imageLayer)
					imageLayer.Draw()
					layer.Draw()
					s.Show()
				}
				debounced(f)
				imageLayer.Draw()
				layer.Draw()
				s.Show()
			} else if ev.Rune() == 'l' || ev.Rune() == 'L' {
				// right
				s.Clear()
				layer.Clear()
				imageLayer.Clear()
				xOffset += 100000
				renderBorders(startingCenter, xOffset, yOffset, zoomOffset, width, height, fc.Features, &layer)
				f := func() {
					s.Clear()
					renderRadar(startingCenter, xOffset, yOffset, zoomOffset, width, height, &imageLayer)
					imageLayer.Draw()
					layer.Draw()
					s.Show()
				}
				debounced(f)
				imageLayer.Draw()
				layer.Draw()
				s.Show()
			} else if ev.Rune() == 'h' || ev.Rune() == 'H' {
				// left
				s.Clear()
				layer.Clear()
				imageLayer.Clear()
				xOffset -= 100000
				renderBorders(startingCenter, xOffset, yOffset, zoomOffset, width, height, fc.Features, &layer)
				f := func() {
					s.Clear()
					renderRadar(startingCenter, xOffset, yOffset, zoomOffset, width, height, &imageLayer)
					imageLayer.Draw()
					layer.Draw()
					s.Show()
				}
				debounced(f)
				imageLayer.Draw()
				layer.Draw()
				s.Show()
			} else if ev.Rune() == 'j' || ev.Rune() == 'J' {
				// down
				s.Clear()
				layer.Clear()
				imageLayer.Clear()
				yOffset -= 100000
				renderBorders(startingCenter, xOffset, yOffset, zoomOffset, width, height, fc.Features, &layer)
				f := func() {
					s.Clear()
					renderRadar(startingCenter, xOffset, yOffset, zoomOffset, width, height, &imageLayer)
					imageLayer.Draw()
					layer.Draw()
					s.Show()
				}
				debounced(f)
				imageLayer.Draw()
				layer.Draw()
				s.Show()
			} else if ev.Rune() == 'k' || ev.Rune() == 'K' {
				// up
				s.Clear()
				layer.Clear()
				imageLayer.Clear()
				yOffset += 100000
				renderBorders(startingCenter, xOffset, yOffset, zoomOffset, width, height, fc.Features, &layer)
				f := func() {
					s.Clear()
					renderRadar(startingCenter, xOffset, yOffset, zoomOffset, width, height, &imageLayer)
					imageLayer.Draw()
					layer.Draw()
					s.Show()
				}
				debounced(f)
				imageLayer.Draw()
				layer.Draw()
				s.Show()
			}

		}
	}
}

func renderRadar(startingCenter orb.Point, xOffset, yOffset float64, zoomOffset, width, height int, layer *Layer) {
	newCenter := orb.Point{startingCenter[0] + xOffset, startingCenter[1] + yOffset}
	bound := FindBound(newCenter, width*2, height*4, 5000+zoomOffset)
	image := GetMap(bound, width, height)
	drawImage(layer, image)
}

func renderBorders(startingCenter orb.Point, xOffset, yOffset float64, zoomOffset, width, height int, features []*geojson.Feature, layer *Layer) {
	drawStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)
	newCenter := orb.Point{startingCenter[0] + xOffset, startingCenter[1] + yOffset}
	bound := FindBound(newCenter, width*2, height*4, 5000+zoomOffset)
	for _, feature := range features {
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
			layer.DrawPolygon(stateFit, drawStyle)
		}
	}
}
