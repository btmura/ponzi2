package ui

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/chart"
)

var (
	chartThumbSize         = image.Pt(155, 105)
	chartThumbRenderOffset = image.Pt(0, viewPadding+chartThumbSize.Y)
	sidebarScrollAmount    = chartThumbRenderOffset
)

type sidebar struct {
	// slots are slots which can have a thumbnail or be a drop site.
	slots []*sidebarSlot

	// sidebarScrollOffset stores the Y offset accumulated from scroll events
	// that should be used to calculate the sidebar's bounds.
	sidebarScrollOffset image.Point

	// bounds is the rectangle to draw within.
	bounds image.Rectangle
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

func (s *sidebar) SetBounds(bounds image.Rectangle) {
	s.bounds = bounds
}

func (s *sidebar) ProcessInput(input *view.Input) {
	slotBounds := image.Rect(
		s.bounds.Min.X, s.bounds.Max.Y-viewPadding-chartThumbSize.Y,
		s.bounds.Max.X, s.bounds.Max.Y-viewPadding,
	)

	draggedSlotIndex := -1

	// Set the bounds for each slot and identify the slot being dragged.
	for i, slot := range s.slots {
		slot.SetBounds(slotBounds)

		// TODO(btmura): If there is a slot being dragged already, do not recalculate this,
		//               because the slots may have already shifted order.
		slot.dragging = input.MouseLeftButtonDragging &&
			input.MouseLeftButtonDraggingStartedPos.In(slotBounds)

		thumbBounds := slotBounds
		if slot.dragging {
			draggedSlotIndex = i
			// TODO(btmura): Add a helper method in rect.go for this.
			thumbBounds = image.Rect(
				input.MousePos.X-slotBounds.Dx()/2, input.MousePos.Y-slotBounds.Dy()/2,
				input.MousePos.X+slotBounds.Dx()/2, input.MousePos.Y+slotBounds.Dy()/2,
			)
		}
		slot.SetThumbnailBounds(thumbBounds)

		slotBounds = slotBounds.Sub(chartThumbRenderOffset)
	}

	// Move the dragged slot to its proper place in the sidebar.
	if draggedSlotIndex != -1 {
		draggedSlot := s.slots[draggedSlotIndex]
		bounds := draggedSlot.Bounds().Add(image.Pt(0, -chartThumbSize.Y))
		if input.MousePos.In(bounds) && draggedSlotIndex+1 < len(s.slots) {
			i, j := draggedSlotIndex, draggedSlotIndex+1
			s.slots[i], s.slots[j] = s.slots[j], s.slots[i]
		}
	}

	for _, slot := range s.slots {
		slot.ProcessInput(input)
	}
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
	// Draw fixed thumbnails before the dragged thumbnails.
	for _, slot := range s.slots {
		if !slot.dragging {
			slot.Render(fudge)
		}
	}

	// Draw dragged thumbnails over fixed thumbnails.
	for _, slot := range s.slots {
		if slot.dragging {
			slot.Render(fudge)
		}
	}
}

type sidebarSlot struct {
	// bounds is the rectangle to draw the slot within. Can be empty for collapsed slots.
	bounds image.Rectangle

	// thumbnail is an optional stock thumbnail. Could have bounds outside the slot if dragged.
	thumbnail *chart.Thumb

	// fader fades out the slot.
	fader *view.Fader

	dragging bool
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
