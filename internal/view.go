package internal

import (
	"log"
	"time"

	"github.com/bep/debounce"
	"github.com/gdamore/tcell/v2"
	"github.com/paulmach/orb"
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

	debounced := debounce.New(1000 * time.Millisecond)

	borders := NewBorderLayer(s, fc.Features, width, height)
	radar := NewRadarLayer(s, width, height)
	borders.Render(center, 5000, width, height)
	radar.Render(center, 5000, width, height)
	zoom := 5000
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
				zoom -= 1000
				stack(s, center, zoom, width, height, borders, radar, debounced)
			} else if ev.Rune() == 'u' || ev.Rune() == 'U' {
				// zoom out
				zoom += 1000
				stack(s, center, zoom, width, height, borders, radar, debounced)
				// s.Show()
			} else if ev.Rune() == 'l' || ev.Rune() == 'L' {
				// right
				center = orb.Point{center[0] + 100000, center[1]}
				stack(s, center, zoom, width, height, borders, radar, debounced)
			} else if ev.Rune() == 'h' || ev.Rune() == 'H' {
				// left
				center = orb.Point{center[0] - 100000, center[1]}
				stack(s, center, zoom, width, height, borders, radar, debounced)
			} else if ev.Rune() == 'j' || ev.Rune() == 'J' {
				// down
				center = orb.Point{center[0], center[1] - 100000}
				stack(s, center, zoom, width, height, borders, radar, debounced)
			} else if ev.Rune() == 'k' || ev.Rune() == 'K' {
				// up
				center = orb.Point{center[0], center[1] + 100000}
				stack(s, center, zoom, width, height, borders, radar, debounced)
			}

		}
	}
}

func stack(s tcell.Screen, center orb.Point, zoom, width, height int, border, radar Layer, debounced func(f func())) {
	s.Clear()
	border.Clear()
	radar.Clear()
	border.Render(center, zoom, width, height)
	f := func() {
		s.Clear()
		radar.Render(center, zoom, width, height)
		border.Render(center, zoom, width, height)
		s.Show()
	}
	debounced(f)
	border.Render(center, zoom, width, height)
	s.Show()
}
