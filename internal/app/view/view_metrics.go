package view

import "image"

const viewOuterPadding = 10

var viewChartThumbSize = image.Pt(155, 105)

type viewMetrics struct {
	sidebarRegion image.Rectangle
}

func (v *View) metrics() *viewMetrics {
	if len(v.chartThumbs) == 0 {
		return &viewMetrics{
			sidebarRegion: image.ZR,
		}
	}

	return &viewMetrics{
		sidebarRegion: image.Rect(viewOuterPadding, 0, viewOuterPadding+viewChartThumbSize.X, v.winSize.Y),
	}
}
