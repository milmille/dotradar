package internal

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/project"
)

func View(stateStr string) {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	drawStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)

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

	bound := FindBound(centerStateMerc.Clone(), width*2, height*4, 5000)

	image := GetMap(centerStateMerc.Clone(), width, height)

	for _, feature := range fc.Features {
		var state orb.MultiPolygon
		if multiPolygon, ok := feature.Geometry.(orb.MultiPolygon); ok {
			state = multiPolygon
		} else if polygon, ok := feature.Geometry.(orb.Polygon); ok {
			state = orb.MultiPolygon{polygon}
		}
		stateMerc := project.MultiPolygon(state, project.WGS84.ToMercator)

		stateClipped := clip.MultiPolygon(bound, stateMerc)
		if !stateClipped.Bound().IsEmpty() {
			stateFit := FitToScreen(stateClipped, bound, width*2, height*4)
			layer.DrawPolygon(stateFit, drawStyle)
		}
	}

	imagePixels := NewPixelSlice(width, height)
	imageLayer := Layer{Pixels: imagePixels, XMultiplier: 1, YMultiplier: 2, screen: s}

	drawImage(imageLayer, image)

	// Here's how to get the screen size when you need it.
	// xmax, ymax := s.Size()

	// Here's an example of how to inject a keystroke where it will
	// be picked up by the next PollEvent call.  Note that the
	// queue is LIFO, it has a limited length, and PostEvent() can
	// return an error.
	// s.PostEvent(tcell.NewEventKey(tcell.KeyRune, rune('a'), 0))
	imageLayer.Draw()
	layer.Draw()
	s.Show()
	// Update screen

	// Event loop
	for {
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
