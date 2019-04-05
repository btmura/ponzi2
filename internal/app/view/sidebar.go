package view

import "image"

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
}

type sidebarThumb struct {
	chartThumb *ChartThumb
	*viewAnimator
}

func (s *sidebar) AddChartThumb(th *ChartThumb) {
	s.thumbs = append(s.thumbs, &sidebarThumb{
		chartThumb:   th,
		viewAnimator: newViewAnimator(th),
	})
}

func (s *sidebar) RemoveChartThumb(th *ChartThumb) {
	for _, vth := range s.thumbs {
		if vth.chartThumb == th {
			vth.Exit()
			break
		}
	}
}

// ProcessInput processes input.
func (s *sidebar) ProcessInput(ic inputContext, m viewMetrics) {
	ic.Bounds = m.firstThumbBounds
	for _, th := range s.thumbs {
		th.ProcessInput(ic)
		ic.Bounds = ic.Bounds.Sub(chartThumbRenderOffset)
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

func (s *sidebar) Render(fudge float32) error {
	if len(s.thumbs) == 0 {
		return nil
	}

	for _, th := range s.thumbs {
		th.Render(fudge)
	}

	return nil
}
