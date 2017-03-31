package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type miniChart struct {
	stock             *modelStock
	stockQuoteText    *dynamicText
	frameBorder       *vao
	frameDivider      *vao
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createMiniChart(stock *modelStock, stockQuoteText *dynamicText) *miniChart {
	return &miniChart{
		stock:             stock,
		stockQuoteText:    stockQuoteText,
		frameBorder:       createStrokedRectVAO(white, white, white, white),
		frameDivider:      createLineVAO(white, white),
		dailyStochastics:  createChartStochastics(stock, stockQuoteText, daily),
		weeklyStochastics: createChartStochastics(stock, stockQuoteText, weekly),
	}
}

func (mc *miniChart) update() {
	mc.dailyStochastics.update()
	mc.weeklyStochastics.update()
}

func (mc *miniChart) render(r image.Rectangle) {
	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	mc.frameBorder.render()

	//
	// Render the symbol and its quote.
	//

	const pad = 5
	s := mc.stockQuoteText.measure(mc.stock.symbol)
	pt.Y -= pad + s.Y
	{
		pt := pt
		pt.X += pad
		pt = pt.Add(mc.stockQuoteText.render(mc.stock.symbol, pt, white))
		pt = pt.Add(mc.stockQuoteText.render(shortFormatQuote(mc.stock.quote), pt, quoteColor(mc.stock.quote)))
	}

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = pt.Y
	gl.Uniform1f(colorMixAmountLocation, 1)

	rects := sliceRectangle(r, 0.5, 0.5)
	for _, r := range rects {
		setModelMatrixRectangle(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		mc.frameDivider.render()
	}

	//
	// Render the graphs.
	//

	mc.dailyStochastics.render(rects[1].Inset(pad))
	mc.weeklyStochastics.render(rects[0].Inset(pad))
}

func (mc *miniChart) close() {
	mc.frameBorder.close()
	mc.frameBorder = nil
	mc.frameDivider.close()
	mc.frameDivider = nil
	mc.dailyStochastics.close()
	mc.dailyStochastics = nil
	mc.weeklyStochastics.close()
	mc.weeklyStochastics = nil
}
