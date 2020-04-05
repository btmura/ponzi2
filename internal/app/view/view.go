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
	// MouseMoved is non-nil when the mouse is just moving without being clicked or dragged.
	MouseMoved *MouseMoveEvent

	// MouseLeftButtonClicked is non-nil when the left mouse button was pressed and released.
	MouseLeftButtonClicked *MouseClickEvent

	// MouseLeftButtonDragging is non-nil when the mouse is being dragged from press to release.
	MouseLeftButtonDragging *MouseDraggingEvent

	// Scroll is the up or down scroll direction or unspecified if the user did not scroll.
	Scroll ScrollDirection

	// ScheduledCallbacks are callbacks to be called at the end of Render.
	ScheduledCallbacks []func()
}

// MouseMoveEvent tracks the mouse when it is just moving without being clicked or dragged.
type MouseMoveEvent struct {
	// Pos is where the mouse is on the screen in global coordinates.
	Pos image.Point
}

// In returns true if the mouse was moved within the bounds.
func (m *MouseMoveEvent) In(bounds image.Rectangle) bool {
	if m == nil {
		return false
	}
	return m != nil && m.Pos.In(bounds)
}

// MouseClickEvent tracks a mouse click which is a press followed by a release.
type MouseClickEvent struct {
	// PressedPos is where the mouse was pressed.
	PressedPos image.Point

	// ReleasedPos is where the mouse was released.
	ReleasedPos image.Point
}

// In returns true if the left mouse button was clicked within
// the bounds. Doesn't take into account overlapping view parts.
func (m *MouseClickEvent) In(bounds image.Rectangle) bool {
	return m != nil && m.PressedPos.In(bounds) && m.ReleasedPos.In(bounds)
}

// MouseDraggingEvent tracks a drag and drop mouse motion from drag to release.
type MouseDraggingEvent struct {
	// PressedPos is always a non-empty Point where the mouse was originally pressed.
	PressedPos image.Point

	// MovedPos is always a non-empty Point where the mouse is currently moving.
	MovedPos image.Point

	// ReleasedPos is a non-empty Point when the mouse button is finally released.
	ReleasedPos image.Point
}

// PressedIn return true if the left mouse button was pressed within the bounds.
func (m *MouseDraggingEvent) PressedIn(bounds image.Rectangle) bool {
	return m != nil && m.PressedPos.In(bounds) && m.PressedPos.In(bounds)
}
