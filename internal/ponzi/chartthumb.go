package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartThumbnail struct {
	stock             *modelStock
	lines             *chartLines
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createChartThumbnail(stock *modelStock) *chartThumbnail {
	return &chartThumbnail{
		stock:             stock,
		lines:             createChartLines(stock),
		dailyStochastics:  createChartStochastics(stock, daily),
		weeklyStochastics: createChartStochastics(stock, weekly),
	}
}

func (ct *chartThumbnail) update() {
	ct.lines.update()
	ct.dailyStochastics.update()
	ct.weeklyStochastics.update()
}

func (ct *chartThumbnail) render(r image.Rectangle) {
	r = renderChartFrame(r, ct.stock, thumbSymbolQuoteTextRenderer, thumbChartRounding, thumbChartPadding)

	gfx.SetColorMixAmount(1)
	rects := sliceRect(r, 0.5)
	for _, r := range rects {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		horizLine.Render()
	}

	//
	// Render the graphs.
	//

	ct.dailyStochastics.render(rects[1].Inset(thumbChartPadding))
	ct.weeklyStochastics.render(rects[0].Inset(thumbChartPadding))
	ct.lines.render(rects[1].Inset(thumbChartPadding))
	ct.lines.render(rects[0].Inset(thumbChartPadding))
}

func (ct *chartThumbnail) close() {
	ct.lines.close()
	ct.lines = nil
	ct.dailyStochastics.close()
	ct.dailyStochastics = nil
	ct.weeklyStochastics.close()
	ct.weeklyStochastics = nil
}
