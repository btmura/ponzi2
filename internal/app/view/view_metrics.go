package view

import "image"

const viewOuterPadding = 10

var viewChartThumbSize = image.Pt(155, 105)

var sidebarScrollAmount = image.Pt(0, viewOuterPadding*viewChartThumbSize.Y)

type viewMetrics struct {
	// sidebarRegion is the left-hand side of the window used to show the sidebar.
	sidebarRegion image.Rectangle

	// sidebarRect is the rectangle to draw the sidebar in.
	sidebarRect image.Rectangle
}

func (v *View) metrics() viewMetrics {
	if len(v.chartThumbs) == 0 {
		return viewMetrics{}
	}

	sbHeight := (viewOuterPadding+viewChartThumbSize.Y)*len(v.chartThumbs) + viewOuterPadding

	sbRect := image.Rect(
		viewOuterPadding, v.winSize.Y-sbHeight,
		viewOuterPadding+viewChartThumbSize.X, v.winSize.Y,
	)
	sbRect = sbRect.Add(v.sidebarOffset)

	sbReg := image.Rect(
		viewOuterPadding, 0,
		viewOuterPadding+viewChartThumbSize.X, v.winSize.Y,
	)

	return viewMetrics{
		sidebarRegion: sbReg,
		sidebarRect:   sbRect,
	}
}
