package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

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
	prices            *chartPrices
	volume            *chartVolume
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics

	symbolQuoteText *dynamicText
	frameBorder     *vao
	frameDivider    *vao
}

func createChart(stock *modelStock, titleText, labelText *dynamicText) *chart {
	return &chart{
		stock:             stock,
		prices:            createChartPrices(stock, labelText),
		volume:            createChartVolume(stock, labelText),
		dailyStochastics:  createChartStochastics(stock, dailyInterval, labelText),
		weeklyStochastics: createChartStochastics(stock, weeklyInterval, labelText),

		symbolQuoteText: titleText,
		frameBorder:     createStrokedRectVAO(white, white, white, white),
		frameDivider:    createLineVAO(white, white),
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
	subRects := ch.renderFrame(r)
	ch.prices.render(subRects[3].Inset(pad))
	ch.volume.render(subRects[2].Inset(pad))
	ch.dailyStochastics.render(subRects[1].Inset(pad))
	ch.weeklyStochastics.render(subRects[0].Inset(pad))
}

func (ch *chart) renderFrame(r image.Rectangle) []image.Rectangle {
	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.frameBorder.render()

	//
	// Render the symbol and its quote.
	//

	const pad = 10
	s := ch.symbolQuoteText.measure(ch.stock.symbol)
	pt.Y -= pad + s.Y
	{
		c := pt
		c.X += pad
		c = c.Add(ch.symbolQuoteText.render(ch.stock.symbol, c))
		c = c.Add(ch.symbolQuoteText.render(formatQuote(ch.stock.quote), c))
	}
	pt.Y -= pad

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = pt.Y
	gl.Uniform1f(colorMixAmountLocation, 1)

	rects := sliceRectangle(r, 0.13, 0.13, 0.13, 0.6)
	for _, r := range rects {
		setModelMatrixRectangle(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		ch.frameDivider.render()
	}
	return rects
}

func (ch *chart) close() {
	if ch == nil {
		return
	}
	ch.prices.close()
	ch.volume.close()
	ch.dailyStochastics.close()
	ch.weeklyStochastics.close()

	ch.frameDivider.close()
	ch.frameBorder.close()
}
