package view

import "image"

const viewPadding = 10

var (
	chartThumbSize         = image.Pt(155, 105)
	chartThumbRenderOffset = image.Pt(0, viewPadding+chartThumbSize.Y)
)

type viewMetrics struct {
	// chartBounds is where to draw the main chart.
	chartBounds image.Rectangle

	// sidebarBounds is where to draw the sidebar with thumbnails.
	sidebarBounds image.Rectangle

	// firstThumbBounds is where to draw the first thumbnail in the sidebar.
	firstThumbBounds image.Rectangle

	// sidebarScrollBounds is where to detect scroll events for the sidebar.
	sidebarScrollBounds image.Rectangle
}

func (v *View) metrics() viewMetrics {
	// +---+---------+---+
	// |   | padding |   |
	// |   +---------+   |
	// |   |         |   |
	// |   |         |   |
	// | p | chart   | p |
	// |   |         |   |
	// |   |         |   |
	// |   +---------+   |
	// |   | padding |   |
	// +---+---------+---+

	if len(v.chartThumbs) == 0 {
		cb := image.Rect(0, 0, v.winSize.X, v.winSize.Y)
		cb = cb.Inset(viewPadding)
		return viewMetrics{chartBounds: cb}
	}

	cb := image.Rect(viewPadding+chartThumbSize.X, 0, v.winSize.X, v.winSize.Y)
	cb = cb.Inset(viewPadding)

	// +---+---------+---+---------+---+
	// |   | padding |   | padding |   |
	// |   +---------+   +---------+   |
	// |   | thumb   |   |         |   |
	// |   +---------+   |         |   |
	// | p | padding | p | chart   | p |
	// |   +---------+   |         |   |
	// |   | thumb   |   |         |   |
	// |   +---------+   +---------+   |
	// |   | padding |   | padding |   |
	// +---+---------+---+---------+---+

	sh := (viewPadding+chartThumbSize.Y)*len(v.chartThumbs) + viewPadding

	sb := image.Rect(
		viewPadding, v.winSize.Y-sh,
		viewPadding+chartThumbSize.X, v.winSize.Y,
	)
	sb = sb.Add(v.sidebarScrollOffset)

	fb := image.Rect(
		sb.Min.X, sb.Max.Y-viewPadding-chartThumbSize.Y,
		sb.Max.X, sb.Max.Y-viewPadding,
	)

	ssb := image.Rect(
		viewPadding, 0,
		viewPadding+chartThumbSize.X, v.winSize.Y,
	)

	return viewMetrics{
		chartBounds:         cb,
		sidebarBounds:       sb,
		firstThumbBounds:    fb,
		sidebarScrollBounds: ssb,
	}
}
