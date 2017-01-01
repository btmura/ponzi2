package ponzi

import "image"

type chart struct {
	symbol            string
	propText          *dynamicText
	frame             *chartFrame
	prices            *chartPrices
	volume            *chartVolume
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createChart(symbol string, propText *dynamicText) *chart {
	return &chart{
		symbol:   symbol,
		propText: propText,
		frame:    createChartFrame(propText),
	}
}

func (f *chart) render(stock *modelStock, r image.Rectangle) {
	if f.prices == nil && stock.dailySessions != nil {
		f.prices = createChartPrices(stock.dailySessions)
		f.volume = createChartVolume(stock.dailySessions)
		f.dailyStochastics = createChartStochastics(stock.dailySessions, [3]float32{1, 1, 0})
		f.weeklyStochastics = createChartStochastics(stock.weeklySessions, [3]float32{1, 0, 1})
	}

	rects := f.frame.render(stock, r)

	setModelMatrixRectangle(rects[3])
	f.prices.render()

	setModelMatrixRectangle(rects[2])
	f.volume.render()

	setModelMatrixRectangle(rects[1])
	f.dailyStochastics.render()

	setModelMatrixRectangle(rects[0])
	f.weeklyStochastics.render()
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
