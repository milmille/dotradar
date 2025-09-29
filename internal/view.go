package internal

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/project"
)

func View() {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	boxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple)

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

	width, height := s.Size()
	pixels := NewPixelSlice(width, height)
	layer := Layer{Pixels: pixels, screen: s}

	fc := ReadGeoJSON("./gz_2010_us_040_00_20m.json")
	minnesotaWGS84 := GetFeature("Minnesota", fc).Geometry.(orb.Polygon)
	//TODO: support multi-polygons
	minnsotaMerc := project.Polygon(minnesotaWGS84, project.WGS84.ToMercator)

	topLeft := orb.Point{-101.54149625952375, 49.31848896184871}
	bottomRight := orb.Point{-88.52181788184106, 42.54514923415394}
	boundWGS84 := orb.MultiPoint{topLeft, bottomRight}.Bound()
	boundMerc := project.Bound(boundWGS84, project.WGS84.ToMercator)

	//TODO: decide the bound based on the aspect ratio of the screen
	minnesotaFit := FitToScreen(minnsotaMerc, boundMerc, width*2, height*4)
	layer.DrawPolygon(minnesotaFit)

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
		fmt.Println(s.Size())
	}
	defer quit()

	// Here's how to get the screen size when you need it.
	// xmax, ymax := s.Size()

	// Here's an example of how to inject a keystroke where it will
	// be picked up by the next PollEvent call.  Note that the
	// queue is LIFO, it has a limited length, and PostEvent() can
	// return an error.
	// s.PostEvent(tcell.NewEventKey(tcell.KeyRune, rune('a'), 0))

	// Event loop
	for {
		layer.Draw(boxStyle)
		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			} else if ev.Key() == tcell.KeyCtrlL {
				s.Sync()
			} else if ev.Rune() == 'C' || ev.Rune() == 'c' {
				s.Clear()
			}
		}
	}
}
