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
		dailyStochastics:  createChartStochastics(stock, dailyInterval),
		weeklyStochastics: createChartStochastics(stock, weeklyInterval),
	}
}

func (f *chart) update() {
	if f == nil {
		return
	}
	f.prices.update()
	f.volume.update()
	f.dailyStochastics.update()
	f.weeklyStochastics.update()
}

func (f *chart) render(r image.Rectangle) {
	if f == nil {
		return
	}
	const pad = 3
	subRects := f.frame.render(r)
	f.prices.render(subRects[3].Inset(pad))
	f.volume.render(subRects[2].Inset(pad))
	f.dailyStochastics.render(subRects[1].Inset(pad))
	f.weeklyStochastics.render(subRects[0].Inset(pad))
}

func (f *chart) close() {
	if f == nil {
		return
	}
	f.frame.close()
	f.prices.close()
	f.volume.close()
	f.dailyStochastics.close()
	f.weeklyStochastics.close()
}
