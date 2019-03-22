package view

import "image"

var (
	chartThumbSize         = image.Pt(155, 105)
	chartThumbRenderOffset = image.Pt(0, viewPadding+chartThumbSize.Y)
	sidebarScrollAmount    = chartThumbRenderOffset
)

type sidebar struct {
	// chartThumbs renders the stocks in the sidebar.
	chartThumbs []*viewChartThumb

	// sidebarScrollOffset stores the Y offset accumulated from scroll events
	// that should be used to calculate the sidebar's bounds.
	sidebarScrollOffset image.Point
}

type viewChartThumb struct {
	chartThumb *ChartThumb
	*viewAnimator
}

func (s *sidebar) AddChartThumb(th *ChartThumb) {
	s.chartThumbs = append(s.chartThumbs, &viewChartThumb{
		chartThumb:   th,
		viewAnimator: newViewAnimator(th),
	})
}

func (s *sidebar) RemoveChartThumb(th *ChartThumb) {
	for _, vth := range s.chartThumbs {
		if vth.chartThumb == th {
			vth.Exit()
			break
		}
	}
}

func (s *sidebar) Update() (dirty bool) {
	for i := 0; i < len(s.chartThumbs); i++ {
		th := s.chartThumbs[i]
		if th.Update() {
			dirty = true
		}
		if th.Remove() {
			s.chartThumbs = append(s.chartThumbs[:i], s.chartThumbs[i+1:]...)
			th.Close()
			i--
		}
	}
	return dirty
}

func (s *sidebar) Render(vc viewContext, m viewMetrics) error {
	if len(s.chartThumbs) == 0 {
		return nil
	}

	vc.Bounds = m.firstThumbBounds
	for _, th := range s.chartThumbs {
		th.Render(vc)
		vc.Bounds = vc.Bounds.Sub(chartThumbRenderOffset)
	}

	return nil
}
