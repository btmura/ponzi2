package ui

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/view/rect"

	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/chart"
)

// Internal constants.
var (
	chartThumbSize      = image.Pt(155, 105)
	sidebarScrollAmount = image.Pt(0, chartThumbSize.Y+viewPadding)
)

// Sidebar is the sidebar that displays thumbnails of stocks.
type Sidebar struct {
	// Slots are the sidebar's slots that have thumbnails.
	Slots []*SidebarSlot
}

// SidebarSlot is a slot in the sidebar that has a thumbnail.
type SidebarSlot struct {
	// Thumbnail is the thumbnail shown in the sidebar slot.
	Thumbnail *chart.Thumb
}

type sidebar struct {
	// slots are slots which can have a Thumbnail or be a drop site.
	slots []*sidebarSlot

	// draggedSlot if not nil is the slot with the Thumbnail being dragged.
	draggedSlot *sidebarSlot

	// scrollOffset stores the Y offset accumulated from mouseScrollDirection event.
	scrollOffset int

	// bounds is the rectangle to draw within.
	bounds image.Rectangle

	// changeCallback is a callback fired when the sidebar changes.
	changeCallback func(sidebar *Sidebar)
}

func newSidebar() *sidebar {
	return new(sidebar)
}

func (s *sidebar) AddChartThumb(th *chart.Thumb) {
	s.slots = append(s.slots, newSidebarSlot(th))
}

func (s *sidebar) RemoveChartThumb(th *chart.Thumb) {
	for _, slot := range s.slots {
		if slot.thumbnail == th {
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

	height := num * chartThumbSize.Y
	if num > 1 {
		// Add padding between thumbnails.
		height += (num - 1) * viewPadding

		// Add padding on top and bottom.
		height += 2 * viewPadding
	}
	return image.Pt(chartThumbSize.X, height)
}

func (s *sidebar) SetBounds(bounds image.Rectangle) {
	s.bounds = bounds
}

func (s *sidebar) ProcessInput(input *view.Input) {
	// Adjust the mouseScrollDirection offset due to the mouse wheel or window size.
	h := s.ContentSize().Y
	if h == 0 || h < s.bounds.Dy() {
		// Reset offset if the sidebar is empty or the contents are less than the bounds.
		s.scrollOffset = 0
	} else {
		// Scroll up or down if the mouseScrollDirection event occurred within the sidebar bounds.
		if input.MouseScrolled.In(s.bounds) {
			switch input.MouseScrolled.Direction {
			case view.ScrollUp:
				s.scrollUp()
			case view.ScrollDown:
				s.scrollDown()
			}
		}

		// Scroll up or down if dragging and hovering over the bumpers.
		if input.MouseLeftButtonDragging != nil && s.draggedSlot != nil {
			pos := input.MouseLeftButtonDragging.CurrentPos
			top := image.Rect(
				s.bounds.Min.X, s.bounds.Max.Y-chartThumbSize.Y/2,
				s.bounds.Max.X, s.bounds.Max.Y)
			bottom := image.Rect(
				s.bounds.Min.X, 0,
				s.bounds.Max.X, chartThumbSize.Y/2)
			switch {
			case pos.In(top):
				s.scrollUp()
			case pos.In(bottom):
				s.scrollDown()
			}
		}
	}

	// Go down the sidebar and assign bounds to each slot and identify the dragged slot.
	slotBounds := image.Rect(
		s.bounds.Min.X, s.bounds.Max.Y-viewPadding-chartThumbSize.Y,
		s.bounds.Max.X, s.bounds.Max.Y-viewPadding,
	)
	slotBounds = slotBounds.Sub(image.Pt(0, s.scrollOffset))

	wasDragging := s.draggedSlot != nil

	if input.MouseLeftButtonDragging == nil {
		s.draggedSlot = nil
	}

	draggedSlotIndex := -1
	for i, slot := range s.slots {
		slot.SetBounds(slotBounds)
		slot.SetThumbnailBounds(slotBounds)

		if s.draggedSlot == nil {
			if input.MouseLeftButtonDragging.PressedIn(slotBounds) {
				s.draggedSlot = slot
			}
		}

		if s.draggedSlot == slot {
			draggedSlotIndex = i
		}

		slotBounds = slotBounds.Sub(image.Pt(0, chartThumbSize.Y+viewPadding))
	}

	stillDragging := s.draggedSlot != nil

	if s.draggedSlot != nil {
		pos := input.MouseLeftButtonDragging.CurrentPos

		// Float the dragged slot's thumbnail to be under the mouse cursor.
		s.draggedSlot.SetThumbnailBounds(rect.FromCenterPointAndSize(pos.Point, chartThumbSize))

		// Determine whether to move the dragged slot up or down.
		dy := -1
		if c := rect.CenterPoint(s.draggedSlot.Bounds()); pos.Y < c.Y {
			dy = +1
		}

		// Find the slot to swap with and swap it.
		for i := draggedSlotIndex + dy; i >= 0 && i < len(s.slots); i += dy {
			if pos.In(s.slots[i].Bounds()) {
				j, k := draggedSlotIndex, i
				s.slots[j], s.slots[k] = s.slots[k], s.slots[j]
				break
			}
		}
	}

	// Forward the input such as clicks to the adjusted sidebar now.
	for _, slot := range s.slots {
		slot.ProcessInput(input)
	}

	if wasDragging && !stillDragging && s.changeCallback != nil {
		sidebar := new(Sidebar)
		for _, slot := range s.slots {
			if slot.thumbnail != nil {
				sidebar.Slots = append(sidebar.Slots, &SidebarSlot{Thumbnail: slot.thumbnail})
			}
		}
		input.ScheduledCallbacks = append(input.ScheduledCallbacks, func() {
			s.changeCallback(sidebar)
		})
	}
}

func (s *sidebar) scrollUp() {
	// Don't allow a gap at the top.
	s.scrollOffset += sidebarScrollAmount.Y
	if s.scrollOffset > 0 {
		s.scrollOffset = 0
	}
}

func (s *sidebar) scrollDown() {
	// Don't mouseScrollDirection if there is no need.
	overflow := s.ContentSize().Y - s.bounds.Dy()
	if overflow <= 0 {
		return
	}

	// Don't allow a gap at the bottom.
	scroll := sidebarScrollAmount.Y
	if available := s.scrollOffset + overflow; scroll > available {
		scroll = available
	}
	s.scrollOffset -= scroll
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
		if slot != s.draggedSlot {
			slot.Render(fudge)
		}
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

type sidebarSlot struct {
	// bounds is the rectangle to draw the slot within. Can be empty for collapsed slots.
	bounds image.Rectangle

	// Thumbnail is an optional stock Thumbnail. Could have bounds outside the slot if dragged.
	thumbnail *chart.Thumb

	// fader fades out the slot.
	fader *view.Fader
}

func newSidebarSlot(th *chart.Thumb) *sidebarSlot {
	return &sidebarSlot{
		thumbnail: th,
		fader:     view.NewFader(1 * fps),
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

func (s *sidebarSlot) SetThumbnailBounds(bounds image.Rectangle) {
	if s.thumbnail != nil {
		s.thumbnail.SetBounds(bounds)
	}
}

func (s *sidebarSlot) ProcessInput(input *view.Input) {
	if s.thumbnail != nil {
		s.thumbnail.ProcessInput(input)
	}
}

func (s *sidebarSlot) Update() (dirty bool) {
	if s.thumbnail != nil && s.thumbnail.Update() {
		dirty = true
	}
	if s.fader.Update() {
		dirty = true
	}
	return dirty
}

func (s *sidebarSlot) Render(fudge float32) {
	s.fader.Render(func(fudge float32) {
		if s.thumbnail != nil {
			s.thumbnail.Render(fudge)
		}
	}, fudge)
}

func (s *sidebarSlot) Close() {
	if s.thumbnail != nil {
		s.thumbnail.Close()
		s.thumbnail = nil
	}
}
