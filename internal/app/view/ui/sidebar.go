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

type sidebarSlot struct {
	*chart.Thumb
	*view.Fader
	dragging bool
}

func newSidebarSlot(th *chart.Thumb) *sidebarSlot {
	return &sidebarSlot{
		Thumb: th,
		Fader: view.NewFader(1 * fps),
	}
}

func (s *sidebarSlot) Update() (dirty bool) {
	if s.Thumb.Update() {
		dirty = true
	}
	if s.Fader.Update() {
		dirty = true
	}
	return dirty
}

func (s *sidebarSlot) Render(fudge float32) {
	s.Fader.Render(s.Thumb.Render, fudge)
}

func (s *sidebar) AddChartThumb(th *chart.Thumb) {
	s.slots = append(s.slots, newSidebarSlot(th))
}

func (s *sidebar) RemoveChartThumb(th *chart.Thumb) {
	for _, t := range s.slots {
		if t.Thumb == th {
			t.FadeOut()
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
	for _, t := range s.slots {
		t.dragging = input.MouseLeftButtonDragging && input.MouseLeftButtonDraggingStartedPos.In(slotBounds)
		if t.dragging {
			bounds := image.Rect(
				input.MousePos.X-slotBounds.Dx()/2, input.MousePos.Y-slotBounds.Dy()/2,
				input.MousePos.X+slotBounds.Dx()/2, input.MousePos.Y+slotBounds.Dy()/2,
			)
			t.SetBounds(bounds)
		} else {
			t.SetBounds(slotBounds)
		}
		t.ProcessInput(input)
		slotBounds = slotBounds.Sub(chartThumbRenderOffset)
	}
}

// Update moves the animation one step forward.
func (s *sidebar) Update() (dirty bool) {
	for i := 0; i < len(s.slots); i++ {
		t := s.slots[i]
		if t.Update() {
			dirty = true
		}
		if t.DoneFadingOut() {
			s.slots = append(s.slots[:i], s.slots[i+1:]...)
			t.Close()
			i--
		}
	}
	return dirty
}

// Render renders a frame.
func (s *sidebar) Render(fudge float32) {
	// Draw fixed thumbnails before the dragged thumbnails.
	for _, t := range s.slots {
		if !t.dragging {
			t.Render(fudge)
		}
	}

	// Draw dragged thumbnails over fixed thumbnails.
	for _, t := range s.slots {
		if t.dragging {
			t.Render(fudge)
		}
	}
}
