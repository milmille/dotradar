package internal

import (
	"log"
	"math"
	"time"

	"github.com/bep/debounce"
	"github.com/gdamore/tcell/v2"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"github.com/paulmach/orb/project"
)

type Container struct {
	Width  int
	Height int
}

const HEIGHT_RATIO = 0.85

func View(stateStr string, logger *log.Logger, showBox bool) {
	logger.Println("starting!")
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

	fc := ReadGeoJSON("./gz_2010_us_040_00_20m.json")
	geometry := GetFeature(stateStr, fc).Geometry
	var centerState orb.MultiPolygon
	if multiPolygon, ok := geometry.(orb.MultiPolygon); ok {
		centerState = multiPolygon
	} else if polygon, ok := geometry.(orb.Polygon); ok {
		centerState = orb.MultiPolygon{polygon}
	}
	centerStateMerc := project.MultiPolygon(centerState.Clone(), project.WGS84.ToMercator)
	center, _ := planar.CentroidArea(centerStateMerc)

	debounced := debounce.New(500 * time.Millisecond)

	var mapContainer *Container
	if showBox {
		mapContainer = &Container{
			Width:  width - 5,
			Height: int(math.Round(float64(height) * HEIGHT_RATIO)),
		}
	}

	borders := NewBorderLayer(s, fc.Features, mapContainer)
	radar := NewRadarLayer(s, mapContainer)

	var boundWidth, boundHeight int
	if mapContainer == nil {
		boundWidth, boundHeight = width*2, height*4
	} else {
		boundWidth, boundHeight = mapContainer.Width*2, mapContainer.Height*4
	}
	zoom := 5000
	bound := FindBound(center, boundWidth, boundHeight, zoom)

	radar.Render(bound, mapContainer)
	s.Show()
	borders.Render(bound, mapContainer)
	s.Show()

	ticker := time.NewTicker(1 * time.Second)
	tickerDone := make(chan bool)
	mainDone := make(chan bool)
	go func() {
		index := 0
		forward := true
		for {
			select {
			case <-tickerDone:
				return
			case <-ticker.C:
				s.SetContent(15+index, mapContainer.Height+5, '|', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset))
				if forward {
					s.SetContent(15+(index-1), mapContainer.Height+5, ' ', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset))
					index++
				} else {
					s.SetContent(15+(index+1), mapContainer.Height+5, ' ', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset))
					index--
				}
				if index == 6 {
					forward = false
					index -= 2
				} else if index == -1 {
					forward = true
					index += 2
				}
				s.Show()
			}
		}
	}()

	go func(mainDone chan bool) {
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
					zoom -= 1000
					bound := FindBound(center, boundWidth, boundHeight, zoom)
					stack(s, bound, mapContainer, borders, radar, debounced)
				} else if ev.Rune() == 'u' || ev.Rune() == 'U' {
					// zoom out
					zoom += 1000
					bound := FindBound(center, boundWidth, boundHeight, zoom)
					stack(s, bound, mapContainer, borders, radar, debounced)
					// s.Show()
				} else if ev.Rune() == 'l' || ev.Rune() == 'L' {
					// right
					center = orb.Point{center[0] + 100000, center[1]}
					bound := FindBound(center, boundWidth, boundHeight, zoom)
					stack(s, bound, mapContainer, borders, radar, debounced)
				} else if ev.Rune() == 'h' || ev.Rune() == 'H' {
					// left
					center = orb.Point{center[0] - 100000, center[1]}
					bound := FindBound(center, boundWidth, boundHeight, zoom)
					stack(s, bound, mapContainer, borders, radar, debounced)
				} else if ev.Rune() == 'j' || ev.Rune() == 'J' {
					// down
					center = orb.Point{center[0], center[1] - 100000}
					bound := FindBound(center, boundWidth, boundHeight, zoom)
					stack(s, bound, mapContainer, borders, radar, debounced)
				} else if ev.Rune() == 'k' || ev.Rune() == 'K' {
					// up
					center = orb.Point{center[0], center[1] + 100000}
					bound := FindBound(center, boundWidth, boundHeight, zoom)
					stack(s, bound, mapContainer, borders, radar, debounced)
				} else if ev.Rune() == 'q' || ev.Rune() == 'Q' {
					mainDone <- true
				}

			}
		}
	}(mainDone)

	// TODO: handle signals correctly
	<-mainDone
	tickerDone <- true
	ticker.Stop()
}

func stack(s tcell.Screen, bound orb.Bound, container *Container, border, radar Layer, debounced func(f func())) {
	// TODO: s.Clear is clearing the whole screen, should write an implementation to only clear the map container
	// to prevent clearing other parts of the UI
	border.Clear()
	radar.Clear()
	f := func() {
		border.Clear()
		radar.Render(bound, container)
		border.Render(bound, container)
		s.Show()
	}
	debounced(f)
	border.Render(bound, container)
	s.Show()
}
