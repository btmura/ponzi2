package ponzi

import (
	"image"
	"math"
)

// Colors used by the chart.
var (
	green  = [3]float32{0.25, 1, 0}
	red    = [3]float32{1, 0.3, 0}
	yellow = [3]float32{1, 1, 0}
	purple = [3]float32{0.5, 0, 1}
	white  = [3]float32{1, 1, 1}
	gray   = [3]float32{0.1, 0.1, 0.1}
)

const chartLabelPadding = 2

type chart struct {
	stock *modelStock

	symbolQuoteText *dynamicText
	labelText       *dynamicText

	minPrice float32
	maxPrice float32

	frameBorder       *vao
	frameDivider      *vao
	stickLines        *vao
	stickRects        *vao
	volume            *chartVolume
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createChart(stock *modelStock, symbolQuoteText, labelText *dynamicText) *chart {
	return &chart{
		stock:             stock,
		symbolQuoteText:   symbolQuoteText,
		labelText:         labelText,
		frameBorder:       createStrokedRectVAO(white, white, white, white),
		frameDivider:      createLineVAO(white, white),
		volume:            createChartVolume(stock, labelText),
		dailyStochastics:  createChartStochastics(stock, labelText, daily),
		weeklyStochastics: createChartStochastics(stock, labelText, weekly),
	}
}

func (ch *chart) update() {
	if ch == nil || ch.stock.dailySessions == nil {
		return
	}

	ch.minPrice, ch.maxPrice = math.MaxFloat32, 0
	for _, s := range ch.stock.dailySessions {
		if ch.minPrice > s.low {
			ch.minPrice = s.low
		}
		if ch.maxPrice < s.high {
			ch.maxPrice = s.high
		}
	}

	ch.stickLines, ch.stickRects = ch.createPriceVAOs()
	ch.volume.update()
	ch.dailyStochastics.update()
	ch.weeklyStochastics.update()
}

func (ch *chart) render(r image.Rectangle) {
	if ch == nil {
		return
	}

	subRects := ch.renderFrame(r)
	pr, vr, dr, wr := subRects[3], subRects[2], subRects[1], subRects[0]

	const pad = 3
	pr = pr.Inset(pad)
	vr = vr.Inset(pad)
	dr = dr.Inset(pad)
	wr = wr.Inset(pad)

	maxWidth := ch.renderPriceLabels(pr)
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

	ch.renderPrices(pr)
	ch.volume.renderGraph(vr)
	ch.dailyStochastics.renderGraph(dr)
	ch.weeklyStochastics.renderGraph(wr)
}

func (ch *chart) close() {
	if ch == nil {
		return
	}
	ch.frameDivider.close()
	ch.frameBorder.close()
	ch.stickLines.close()
	ch.stickRects.close()
	ch.volume.close()
	ch.dailyStochastics.close()
	ch.weeklyStochastics.close()
}
