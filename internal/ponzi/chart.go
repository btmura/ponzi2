package ponzi

import "image"

// Colors used by the chart.
var (
	black    = [3]float32{0, 0, 0}
	blue     = [3]float32{0, 0.25, 0.5}
	darkBlue = [3]float32{0, 0.05, 0.15}
	green    = [3]float32{0.25, 1, 0}
	red      = [3]float32{1, 0.3, 0}
	yellow   = [3]float32{1, 1, 0}
	purple   = [3]float32{0.5, 0, 1}
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

	rects := f.frame.render(stock, r)
	f.prices.render(rects[3])
	f.volume.render(rects[2])
	f.dailyStochastics.render(rects[1])
	f.weeklyStochastics.render(rects[0])
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
