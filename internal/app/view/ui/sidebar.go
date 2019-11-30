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
	// thumbs renders the stocks in the sidebar.
	thumbs []*sidebarThumb

	// sidebarScrollOffset stores the Y offset accumulated from scroll events
	// that should be used to calculate the sidebar's bounds.
	sidebarScrollOffset image.Point

	// bounds is the rectangle to draw within.
	bounds image.Rectangle
}

type sidebarThumb struct {
	*chart.Thumb
	*view.Fader
	dragging bool
}

func newSidebarThumb(th *chart.Thumb) *sidebarThumb {
	return &sidebarThumb{
		Thumb: th,
		Fader: view.NewFader(1 * fps),
	}
}

func (t *sidebarThumb) Update() (dirty bool) {
	if t.Thumb.Update() {
		dirty = true
	}
	if t.Fader.Update() {
		dirty = true
	}
	return dirty
}

func (t *sidebarThumb) Render(fudge float32) {
	t.Fader.Render(t.Thumb.Render, fudge)
}

func (s *sidebar) AddChartThumb(th *chart.Thumb) {
	s.thumbs = append(s.thumbs, newSidebarThumb(th))
}

func (s *sidebar) RemoveChartThumb(th *chart.Thumb) {
	for _, t := range s.thumbs {
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
	for _, t := range s.thumbs {
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
	for i := 0; i < len(s.thumbs); i++ {
		t := s.thumbs[i]
		if t.Update() {
			dirty = true
		}
		if t.DoneFadingOut() {
			s.thumbs = append(s.thumbs[:i], s.thumbs[i+1:]...)
			t.Close()
			i--
		}
	}
	return dirty
}

// Render renders a frame.
func (s *sidebar) Render(fudge float32) {
	// Draw fixed thumbnails before the dragged thumbnails.
	for _, t := range s.thumbs {
		if !t.dragging {
			t.Render(fudge)
		}
	}

	// Draw dragged thumbnails over fixed thumbnails.
	for _, t := range s.thumbs {
		if t.dragging {
			t.Render(fudge)
		}
	}
}
