package ponzi

import "image"

// Colors used by the chart.
var (
	green  = [3]float32{0.25, 1, 0}
	red    = [3]float32{1, 0.3, 0}
	yellow = [3]float32{1, 1, 0}
	purple = [3]float32{0.5, 0, 1}
	white  = [3]float32{1, 1, 1}
)

type chart struct {
	stock             *modelStock
	frame             *chartFrame
	prices            *chartPrices
	volume            *chartVolume
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createChart(stock *modelStock, titleText, labelText *dynamicText) *chart {
	return &chart{
		stock:             stock,
		frame:             createChartFrame(stock, titleText),
		prices:            createChartPrices(stock, labelText),
		volume:            createChartVolume(stock, labelText),
		dailyStochastics:  createChartStochastics(stock, dailyInterval, labelText),
		weeklyStochastics: createChartStochastics(stock, weeklyInterval, labelText),
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
	const pad = 3
	subRects := ch.frame.render(r)
	ch.prices.render(subRects[3].Inset(pad))
	ch.volume.render(subRects[2].Inset(pad))
	ch.dailyStochastics.render(subRects[1].Inset(pad))
	ch.weeklyStochastics.render(subRects[0].Inset(pad))
}

func (ch *chart) close() {
	if ch == nil {
		return
	}
	ch.frame.close()
	ch.prices.close()
	ch.volume.close()
	ch.dailyStochastics.close()
	ch.weeklyStochastics.close()
}
