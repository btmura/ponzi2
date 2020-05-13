package ui

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/log"
)

// Internal constants.
var (
	// thumbSize is the size of the thumbnails in the sidebar.
	thumbSize = image.Pt(155, 105)

	// wheelScrollAmount is how much to scroll when the mouse scroll wheel is used.
	wheelScrollAmount = image.Pt(0, thumbSize.Y+viewPadding)

	// bumperScrollAmount is how much to scroll when dragging over the bumpers which are the edges of the sidebar.
	bumperScrollAmount = image.Pt(0, (thumbSize.Y+viewPadding)/4)
)

// Sidebar is the sidebar that displays thumbnails of stocks.
type Sidebar struct {
	// Slots are the sidebar's slots that have thumbnails.
	Slots []*SidebarSlot
}

// SidebarSlot is a slot in the sidebar that has a thumbnail.
type SidebarSlot struct {
	// Thumb is the thumbnail shown in the sidebar slot.
	Thumb *chart.Thumb
}

// sidebar manages the sidebar to display and edit stock thumbnails.
type sidebar struct {
	// slots are slots which can have a Thumb or be a drop site.
	slots []*sidebarSlot

	// draggedSlot if not nil is the slot with the Thumb being dragged.
	draggedSlot *draggedSidebarSlot

	// scrollOffset stores the Y offset accumulated from mouseScrollDirection event.
	scrollOffset int

	// bounds is the rectangle to draw within.
	bounds image.Rectangle

	// changeCallback is a callback fired when the sidebar changes.
	changeCallback func(sidebar *Sidebar)
}

// sidebarSlot is a drag and drop container for single thumbnail.
type sidebarSlot struct {
	// bounds is the rectangle to draw the slot within.
	bounds image.Rectangle

	// thumbBounds is the bounds of the thumbnail.
	thumbBounds image.Rectangle

	// thumb is an optional stock thumbnail. Could have bounds outside the slot if dragged.
	thumb *chart.Thumb

	// fader fades out the slot.
	fader *view.Fader
}

// draggedSidebarSlot has additional info about the dragged slot.
type draggedSidebarSlot struct {
	// sidebarSlot is the slot being dragged.
	*sidebarSlot

	// mousePressOffset is an offset from the center of the slot.
	mousePressOffset image.Point
}

func newSidebar() *sidebar {
	return new(sidebar)
}

func (s *sidebar) AddChartThumb(thumb *chart.Thumb) {
	if thumb == nil {
		log.Error("thumb should not be nil")
		return
	}
	s.slots = append(s.slots, newSidebarSlot(thumb))
}

func (s *sidebar) RemoveChartThumb(thumb *chart.Thumb) {
	if thumb == nil {
		log.Error("thumb should not be nil")
		return
	}
	for _, slot := range s.slots {
		if slot.thumb == thumb {
			slot.FadeOut()
			break
		}
	}
}

// ContentSize returns the size of the sidebar's contents like thumbnails
// which could be less than the sidebar's bounds if there are not that many thumbnails.
func (s *sidebar) ContentSize() image.Point {
	num := len(s.slots)
	if num == 0 {
		return image.Pt(0, 0)
	}

	height := num * thumbSize.Y
	if num > 1 {
		// Add padding between thumbnails.
		height += (num - 1) * viewPadding

		// Add padding on top and bottom.
		height += 2 * viewPadding
	}
	return image.Pt(thumbSize.X, height)
}

func (s *sidebar) SetBounds(bounds image.Rectangle) {
	s.bounds = bounds
}

func (s *sidebar) ProcessInput(input *view.Input) {
	if input == nil {
		log.Error("input should not be nil")
		return
	}

	// Adjust the scroll offset due to mouse wheel scrolling or dragging over the bumpers.
	s.adjustScrollOffset(input)

	// Set each slot's bounds on the screen.
	s.setSlotBounds()

	// Find the slot being dragged and update its position.
	wasDragging := s.draggedSlot != nil
	s.setDraggedSlot(input)
	stillDragging := s.draggedSlot != nil

	// Absorb mouse event if dragging was released or still dragging.
	if wasDragging && !stillDragging || stillDragging {
		input.ClearMouseInput()
	}

	// Forward the input to the individual slots.
	for _, slot := range s.slots {
		slot.ProcessInput(input)
	}

	// Schedule a change event once the drag and drop is done.
	if wasDragging && !stillDragging {
		s.fireSidebarChangeCallback(input)
	}
}

func (s *sidebar) adjustScrollOffset(input *view.Input) {
	if input == nil {
		log.Error("input should not be nil")
		return
	}

	// Adjust the mouseScrollDirection offset due to the mouse wheel or window size.
	h := s.ContentSize().Y
	if h == 0 || h < s.bounds.Dy() {
		// Reset offset if the sidebar is empty or the contents are less than the bounds.
		s.scrollOffset = 0
		return
	}

	// Scroll up or down if dragging and hovering over the bumpers.
	if input.MouseLeftButtonDragging != nil && s.draggedSlot != nil {
		pos := input.MouseLeftButtonDragging.CurrentPos
		top := image.Rect(
			s.bounds.Min.X, s.bounds.Max.Y-thumbSize.Y/2,
			s.bounds.Max.X, s.bounds.Max.Y)
		bottom := image.Rect(
			s.bounds.Min.X, 0,
			s.bounds.Max.X, thumbSize.Y/2)
		switch {
		case pos.In(top):
			s.scrollUp(bumperScrollAmount.Y)
			return
		case pos.In(bottom):
			s.scrollDown(bumperScrollAmount.Y)
			return
		}
	}

	// Scroll up or down if the mouseScrollDirection event occurred within the sidebar bounds.
	if input.MouseScrolled.In(s.bounds) {
		switch input.MouseScrolled.Direction {
		case view.ScrollUp:
			s.scrollUp(wheelScrollAmount.Y)
			return
		case view.ScrollDown:
			s.scrollDown(wheelScrollAmount.Y)
			return
		default:
			log.Errorf("unsupported scroll direction: %v", input.MouseScrolled.Direction)
			return
		}
	}
}

func (s *sidebar) scrollUp(scrollAmount int) {
	// Don't allow a gap at the top.
	s.scrollOffset += scrollAmount
	if s.scrollOffset > 0 {
		s.scrollOffset = 0
	}
}

func (s *sidebar) scrollDown(scrollAmount int) {
	// Don't scroll if there is no need.
	overflow := s.ContentSize().Y - s.bounds.Dy()
	if overflow <= 0 {
		return
	}

	// Don't allow a gap at the bottom.
	scroll := scrollAmount
	if available := s.scrollOffset + overflow; scroll > available {
		scroll = available
	}
	s.scrollOffset -= scroll
}

// setSlotBounds goes through the sidebar and assign bounds to each slot.
func (s *sidebar) setSlotBounds() {
	slotBounds := image.Rect(
		s.bounds.Min.X, s.bounds.Max.Y-viewPadding-thumbSize.Y,
		s.bounds.Max.X, s.bounds.Max.Y-viewPadding,
	)
	slotBounds = slotBounds.Sub(image.Pt(0, s.scrollOffset))

	for _, slot := range s.slots {
		slot.SetBounds(slotBounds)
		slot.SetThumbBounds(slotBounds)
		slotBounds = slotBounds.Sub(image.Pt(0, thumbSize.Y+viewPadding))
	}
}

// setDraggedSlot finds the slot being dragged and updates its position.
func (s *sidebar) setDraggedSlot(input *view.Input) {
	if input == nil {
		log.Error("input should not be nil")
		return
	}

	if input.MouseLeftButtonDragging == nil {
		s.draggedSlot = nil
	}

	for _, slot := range s.slots {
		if s.draggedSlot == nil && input.MouseLeftButtonDragging.PressedIn(slot.Bounds()) {
			s.draggedSlot = &draggedSidebarSlot{
				sidebarSlot:      slot,
				mousePressOffset: input.MouseLeftButtonDragging.CurrentPos.Sub(rect.CenterPoint(slot.Bounds())),
			}
		}
	}

	if s.draggedSlot != nil {
		// Determine the center from where the user pressed and held.
		center := input.MouseLeftButtonDragging.CurrentPos.Sub(s.draggedSlot.mousePressOffset)

		// Float the dragged slot's thumbnail to be under the mouse cursor.
		thumbBounds := rect.FromCenterPointAndSize(center, thumbSize)
		s.draggedSlot.SetThumbBounds(thumbBounds)

		s.moveDraggedSlot(input)
	}
}

// moveDraggedSlot moves the dragged slot up or down depending on the mouse position.
func (s *sidebar) moveDraggedSlot(input *view.Input) {
	if input == nil {
		log.Error("input should not be nil")
		return
	}

	if s.draggedSlot == nil {
		log.Error("draggedSlot should not be nil")
		return
	}

	if input.MouseLeftButtonDragging == nil {
		log.Error("should be dragging if dragged slot exists")
		return
	}

	// Determine whether to move the dragged slot up or down
	// by checking whether the mouse is moving up or down.

	currentPos := input.MouseLeftButtonDragging.CurrentPos
	previousPos := input.MouseLeftButtonDragging.PreviousPos

	var dy int
	switch {
	case currentPos.Y < previousPos.Y:
		dy = +1
	case currentPos.Y > previousPos.Y:
		dy = -1
	}

	// If the mouse is not moving, then we don't need to do anything.
	if dy == 0 {
		return
	}

	// Find the index of the dragged slot.
	draggedSlotIndex := -1
	for i := range s.slots {
		if s.draggedSlot.sidebarSlot == s.slots[i] {
			draggedSlotIndex = i
			break
		}
	}

	if draggedSlotIndex < 0 {
		log.Error("should find dragged slot if draggedSlot exists")
		return
	}

	// Find the slot to swap with and swap it.
	for i := draggedSlotIndex + dy; i >= 0 && i < len(s.slots); i += dy {
		b := s.slots[i].Bounds()
		xOverlaps := b.Overlaps(s.draggedSlot.ThumbBounds())
		yOverlaps := b.Min.Y <= currentPos.Y && currentPos.Y < b.Max.Y
		if xOverlaps && yOverlaps {
			j, k := draggedSlotIndex, i
			s.slots[j], s.slots[k] = s.slots[k], s.slots[j]
			break
		}
	}
}

// fireSidebarChangeCallback schedules the sidebar change callback.
func (s *sidebar) fireSidebarChangeCallback(input *view.Input) {
	if input == nil {
		log.Error("input should not be nil")
		return
	}

	if s.changeCallback == nil {
		return
	}

	sidebar := new(Sidebar)
	for _, slot := range s.slots {
		if slot.thumb != nil {
			sidebar.Slots = append(sidebar.Slots, &SidebarSlot{Thumb: slot.thumb})
		}
	}

	input.AddFiredCallback(func() {
		s.changeCallback(sidebar)
	})
}

// Update moves the animation one step forward.
func (s *sidebar) Update() (dirty bool) {
	for i := 0; i < len(s.slots); i++ {
		slot := s.slots[i]
		if slot.Update() {
			dirty = true
		}
		if slot.DoneFadingOut() {
			s.slots = append(s.slots[:i], s.slots[i+1:]...)
			slot.Close()
			i--
		}
	}
	return dirty
}

// Render renders a frame.
func (s *sidebar) Render(fudge float32) {
	// Draw the non-dragged thumbnails first, so they appear under the dragged thumbnail.
	for _, slot := range s.slots {
		if s.draggedSlot != nil && s.draggedSlot.sidebarSlot == slot {
			continue
		}
		slot.Render(fudge)
	}

	if s.draggedSlot != nil {
		s.draggedSlot.Render(fudge)
	}
}

func (s *sidebar) SetChangeCallback(cb func(sidebar *Sidebar)) {
	s.changeCallback = cb
}

func (s *sidebar) Close() {
	s.changeCallback = nil
}

func newSidebarSlot(thumb *chart.Thumb) *sidebarSlot {
	return &sidebarSlot{
		thumb: thumb,
		fader: view.NewFader(1 * fps),
	}
}

func (s *sidebarSlot) FadeOut() {
	s.fader.FadeOut()
}

func (s *sidebarSlot) DoneFadingOut() bool {
	return s.fader.DoneFadingOut()
}

func (s *sidebarSlot) SetBounds(bounds image.Rectangle) {
	s.bounds = bounds
}

func (s *sidebarSlot) Bounds() image.Rectangle {
	return s.bounds
}

func (s *sidebarSlot) SetThumbBounds(bounds image.Rectangle) {
	s.thumbBounds = bounds

	if s.thumb == nil {
		log.Error("thumbnail should not be nil")
		return
	}
	s.thumb.SetBounds(bounds)
}

func (s *sidebarSlot) ThumbBounds() image.Rectangle {
	return s.thumbBounds
}

func (s *sidebarSlot) ProcessInput(input *view.Input) {
	if input == nil {
		log.Error("input should not be nil")
		return
	}

	if s.thumb != nil {
		s.thumb.ProcessInput(input)
	}
}

func (s *sidebarSlot) Update() (dirty bool) {
	if s.thumb != nil && s.thumb.Update() {
		dirty = true
	}
	if s.fader.Update() {
		dirty = true
	}
	return dirty
}

func (s *sidebarSlot) Render(fudge float32) {
	s.fader.Render(func(fudge float32) {
		if s.thumb != nil {
			s.thumb.Render(fudge)
		}
	}, fudge)
}

func (s *sidebarSlot) Close() {
	if s.thumb != nil {
		s.thumb.Close()
		s.thumb = nil
	}
}
