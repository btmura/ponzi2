package ponzi

import "image"

// Colors used by the chart.
var (
	black  = [3]float32{0, 0, 0}
	green  = [3]float32{0.25, 1, 0}
	red    = [3]float32{1, 0.3, 0}
	yellow = [3]float32{1, 1, 0}
	purple = [3]float32{0.5, 0, 1}
	white  = [3]float32{1, 1, 1}
)

type chart struct {
	symbol            string
	frame             *chartFrame
	prices            *chartPrices
	volume            *chartVolume
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createChart(symbol string, propText *dynamicText) *chart {
	return &chart{
		symbol: symbol,
		frame:  createChartFrame(propText),
	}
}

func (f *chart) render(stock *modelStock, r image.Rectangle) {
	if f.prices == nil && stock.dailySessions != nil {
		f.prices = createChartPrices(stock.dailySessions)
		f.volume = createChartVolume(stock.dailySessions)
		f.dailyStochastics = createChartStochastics(stock.dailySessions, yellow)
		f.weeklyStochastics = createChartStochastics(stock.weeklySessions, purple)
	}

	const p = 3
	rects := f.frame.render(stock, r)
	f.prices.render(rects[3].Inset(p))
	f.volume.render(rects[2].Inset(p))
	f.dailyStochastics.render(rects[1].Inset(p))
	f.weeklyStochastics.render(rects[0].Inset(p))
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
