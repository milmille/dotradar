package internal

import "github.com/paulmach/orb"

type Layer interface {
	// Render the layer on the canvas given a view defined by the center point, zoom level, and size of the screen
	Render(bound orb.Bound, width, height int)
	Clear()
}
