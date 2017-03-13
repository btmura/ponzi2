package ponzi

import (
	"image"
)

// Colors used by the chart.
var (
	green  = [3]float32{0.25, 1, 0}
	red    = [3]float32{1, 0.3, 0}
	yellow = [3]float32{1, 1, 0}
	purple = [3]float32{0.5, 0, 1}
	white  = [3]float32{1, 1, 1}
	gray   = [3]float32{0.15, 0.15, 0.15}
)

const chartLabelPadding = 2

type chart struct {
	stock             *modelStock
	frame             *chartFrame
	prices            *chartPrices
	volume            *chartVolume
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createChart(stock *modelStock, symbolQuoteText, labelText *dynamicText) *chart {
	return &chart{
		stock:             stock,
		frame:             createChartFrame(stock, symbolQuoteText),
		prices:            createChartPrices(stock, labelText),
		volume:            createChartVolume(stock, labelText),
		dailyStochastics:  createChartStochastics(stock, labelText, daily),
		weeklyStochastics: createChartStochastics(stock, labelText, weekly),
	}
}

func (ch *chart) update() {
	if ch == nil {
		return
	}

	ch.prices.update()
	ch.volume.update()
	ch.dailyStochastics.update()
	ch.weeklyStochastics.update()
}

func (ch *chart) render(r image.Rectangle) {
	if ch == nil {
		return
	}

	subRects := ch.frame.render(r)
	pr, vr, dr, wr := subRects[3], subRects[2], subRects[1], subRects[0]

	const pad = 3
	pr = pr.Inset(pad)
	vr = vr.Inset(pad)
	dr = dr.Inset(pad)
	wr = wr.Inset(pad)

	maxWidth := ch.prices.renderLabels(pr)
	if w := ch.volume.renderLabels(vr); w > maxWidth {
		maxWidth = w
	}
	if w := ch.dailyStochastics.renderLabels(dr); w > maxWidth {
		maxWidth = w
	}
	if w := ch.weeklyStochastics.renderLabels(wr); w > maxWidth {
		maxWidth = w
	}

	pr.Max.X -= maxWidth + pad
	vr.Max.X -= maxWidth + pad
	dr.Max.X -= maxWidth + pad
	wr.Max.X -= maxWidth + pad

	ch.prices.renderGraph(pr)
	ch.volume.renderGraph(vr)
	ch.dailyStochastics.renderGraph(dr)
	ch.weeklyStochastics.renderGraph(wr)
}

func (ch *chart) close() {
	if ch == nil {
		return
	}
	ch.frame.close()
	ch.frame = nil
	ch.prices.close()
	ch.prices = nil
	ch.volume.close()
	ch.volume = nil
	ch.dailyStochastics.close()
	ch.dailyStochastics = nil
	ch.weeklyStochastics.close()
	ch.weeklyStochastics = nil
}
