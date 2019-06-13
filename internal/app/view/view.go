package view

import "image"

// ZoomChange specifies whether the user has zoomed in or not.
type ZoomChange int

// ZoomChange values.
//go:generate stringer -type=ZoomChange
const (
	ZoomChangeUnspecified ZoomChange = iota
	ZoomIn
	ZoomOut
)

// Input contains input events to be passed down the view hierarchy.
type Input struct {
	// MousePos is the current global mouse position.
	MousePos image.Point

	// MouseLeftButtonDragging is whether the left mouse button is dragging.
	MouseLeftButtonDragging bool

	// MouseLeftButtonReleased is whether the left mouse button was released.
	MouseLeftButtonReleased bool

	// ScheduledCallbacks are callbacks to be called at the end of Render.
	ScheduledCallbacks []func()
}

// LeftClickInBounds returns true if the left mouse button was clicked within
// the bounds. Doesn't take into account overlapping view parts.
func (i *Input) LeftClickInBounds(bounds image.Rectangle) bool {
	return i.MouseLeftButtonReleased && i.MousePos.In(bounds)
}
