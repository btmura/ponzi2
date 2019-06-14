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
	chartThumb *chart.Thumb
	*viewAnimator
}

func (s *sidebar) AddChartThumb(th *chart.Thumb) {
	s.thumbs = append(s.thumbs, &sidebarThumb{
		chartThumb:   th,
		viewAnimator: newViewAnimator(th),
	})
}

func (s *sidebar) RemoveChartThumb(th *chart.Thumb) {
	for _, vth := range s.thumbs {
		if vth.chartThumb == th {
			vth.Exit()
			break
		}
	}
}

// SetBounds sets the bounds to draw within.
func (s *sidebar) SetBounds(bounds image.Rectangle) {
	s.bounds = bounds
}

// ProcessInput processes input.
func (s *sidebar) ProcessInput(input *view.Input) {
	thumbBounds := image.Rect(
		s.bounds.Min.X, s.bounds.Max.Y-viewPadding-chartThumbSize.Y,
		s.bounds.Max.X, s.bounds.Max.Y-viewPadding,
	)
	for _, th := range s.thumbs {
		th.chartThumb.SetBounds(thumbBounds)
		th.chartThumb.ProcessInput(input)
		thumbBounds = thumbBounds.Sub(chartThumbRenderOffset)
	}
}

func (s *sidebar) Update() (dirty bool) {
	for i := 0; i < len(s.thumbs); i++ {
		th := s.thumbs[i]
		if th.Update() {
			dirty = true
		}
		if th.DoneExiting() {
			s.thumbs = append(s.thumbs[:i], s.thumbs[i+1:]...)
			th.Close()
			i--
		}
	}
	return dirty
}

func (s *sidebar) Render(fudge float32) {
	if len(s.thumbs) == 0 {
		return
	}

	for _, th := range s.thumbs {
		th.Render(fudge)
	}
}
