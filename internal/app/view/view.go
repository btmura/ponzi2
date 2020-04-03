package view

import (
	"image"
)

// ScrollDirection specifies the mouse scroll wheel direction.
type ScrollDirection int

// ScrollDirection values.
//go:generate stringer -type ScrollDirection
const (
	ScrollDirectionUnspecified ScrollDirection = iota
	ScrollUp
	ScrollDown
)

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

	// MouseLeftButtonDraggingStartedPos is the global mouse position where the dragging started.
	// Only set when MouseLeftButtonDragging is true.
	MouseLeftButtonDraggingStartedPos image.Point

	// MouseLeftButtonClicked is non-nil when the left mouse button was pressed and released.
	MouseLeftButtonClicked *MouseClickEvent

	// Scroll is the up or down scroll direction or unspecified if the user did not scroll.
	Scroll ScrollDirection

	// ScheduledCallbacks are callbacks to be called at the end of Render.
	ScheduledCallbacks []func()
}

// MouseClickEvent tracks a mouse click which is a press followed by a release.
type MouseClickEvent struct {
	// MousePressedPos is where the mouse was pressed.
	MousePressedPos image.Point

	// MouseReleasedPos is where the mouse was released.
	MouseReleasedPos image.Point
}

// In returns true if the left mouse button was clicked within
// the bounds. Doesn't take into account overlapping view parts.
func (m *MouseClickEvent) In(bounds image.Rectangle) bool {
	return m != nil && m.MousePressedPos.In(bounds) && m.MouseReleasedPos.In(bounds)
}
