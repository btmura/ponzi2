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

func (s *sidebar) Render(vc viewContext, m viewMetrics) error {
	if len(s.thumbs) == 0 {
		return nil
	}

	vc.Bounds = m.firstThumbBounds
	for _, th := range s.thumbs {
		th.Render(vc)
		vc.Bounds = vc.Bounds.Sub(chartThumbRenderOffset)
	}

	return nil
}
