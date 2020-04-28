// Package view
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

// Input contains input events to be passed down the view hierarchy.
type Input struct {
	// MousePos is the current mouse position. Nil if a view does not want to report mouse input.
	MousePos *MousePosition

	// MouseLeftButtonClicked reports a left mouse button click. Nil if no click happened.
	MouseLeftButtonClicked *MouseClickEvent

	// MouseLeftButtonDragging reports the mouse being dragged with the left button held.
	// Nil if no click happened.
	MouseLeftButtonDragging *MouseDraggingEvent

	// MouseScrolled reports a mouse scroll wheel event. Nil if no scroll happened.
	MouseScrolled *MouseScrollEvent

	// firedCallbacks are callbacks that were fired while processing input.
	firedCallbacks []func()
}

// ClearMouseInput removes mouse input from the Input.
// This is used if a view handles the mouse input and does not want another view to handle them.
func (i *Input) ClearMouseInput() {
	i.MousePos = nil
	i.MouseLeftButtonClicked = nil
	i.MouseLeftButtonDragging = nil
	i.MouseScrolled = nil
}

// AddFiredCallback records a callback that was fired while processing input.
func (i *Input) AddFiredCallback(cb func()) {
	i.firedCallbacks = append(i.firedCallbacks, cb)
}

// FiredCallbacks returns the callbacks that were filed while processing input.
func (i *Input) FiredCallbacks() []func() {
	return i.firedCallbacks
}

// MousePosition is the position of the mouse.
type MousePosition struct {
	image.Point
}

// In returns true if the mouse position is within the bounds.
func (m *MousePosition) In(bounds image.Rectangle) bool {
	return m != nil && m.Point.In(bounds)
}

// MouseClickEvent reports a mouse click that just happened.
type MouseClickEvent struct {
	// PressedPos is where the mouse was pressed.
	PressedPos MousePosition

	// ReleasedPos is where the mouse was released.
	ReleasedPos MousePosition
}

// In returns true if the left mouse button was clicked within
// the bounds. Doesn't take into account overlapping view parts.
func (m *MouseClickEvent) In(bounds image.Rectangle) bool {
	return m != nil && m.PressedPos.In(bounds) && m.ReleasedPos.In(bounds)
}

// MouseDraggingEvent reports a drag and drop mouse motion from drag to release.
type MouseDraggingEvent struct {
	// CurrentPos is the current mouse position.
	CurrentPos MousePosition

	// PreviousPos is the previous mouse position.
	PreviousPos MousePosition

	// PressedPos is where the mouse button was originally pressed.
	PressedPos MousePosition

	// ReleasedPos is where the mouse button was finally released if it has been.
	// Nil if the mouse has not been released yet.
	ReleasedPos *MousePosition
}

// PressedIn return true if the left mouse button was pressed within the bounds.
func (m *MouseDraggingEvent) PressedIn(bounds image.Rectangle) bool {
	return m != nil && m.PressedPos.In(bounds) && m.PressedPos.In(bounds)
}

// MouseScrollEvent reports a mouse scroll wheel event.
type MouseScrollEvent struct {
	// CurrentPos is the current mouse position.
	CurrentPos MousePosition

	// Direction is the direction the scroll wheel went. Always a non-unspecified value.
	Direction ScrollDirection
}

// In returns true if the mouse scroll happened within the bounds.
func (m *MouseScrollEvent) In(bounds image.Rectangle) bool {
	return m != nil && m.CurrentPos.In(bounds)
}
