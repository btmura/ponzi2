package ponzi

import (
	"image"
)

type chart struct {
	stock             *modelStock
	frame             *chartFrame
	lines             *chartLines
	prices            *chartPrices
	volume            *chartVolume
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

// TODO(btmura): create chart factory that creates charts
// TODO(btmura): create chart components and pass them in
func createChart(stock *modelStock) *chart {
	return &chart{
		stock:             stock,
		frame:             createChartFrame(stock),
		lines:             createChartLines(stock),
		prices:            createChartPrices(stock),
		volume:            createChartVolume(stock),
		dailyStochastics:  createChartStochastics(stock, daily),
		weeklyStochastics: createChartStochastics(stock, weekly),
	}
}

func (ch *chart) update() {
	ch.lines.update()
	ch.prices.update()
	ch.volume.update()
	ch.dailyStochastics.update()
	ch.weeklyStochastics.update()
}

func (ch *chart) render(r image.Rectangle) {
	subRects := ch.frame.render(r)
	pr, vr, dr, wr := subRects[3], subRects[2], subRects[1], subRects[0]

	const pad = 5
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

	ch.prices.render(pr)
	ch.volume.render(vr)
	ch.dailyStochastics.render(dr)
	ch.weeklyStochastics.render(wr)

	ch.lines.render(pr)
	ch.lines.render(vr)
	ch.lines.render(dr)
	ch.lines.render(wr)
}

func (ch *chart) close() {
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
