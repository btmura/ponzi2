package ponzi

import (
	"image"
)

type chartThumbnail struct {
	stock             *modelStock
	header            *chartHeader
	lines             *chartLines
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createChartThumbnail(stock *modelStock) *chartThumbnail {
	return &chartThumbnail{
		stock:             stock,
		header:            newChartHeader(stock, thumbSymbolQuoteTextRenderer, thumbChartRounding, thumbChartPadding),
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
	r = ct.header.render(r)

	rects := renderHorizDividers(r, 0.5)

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
