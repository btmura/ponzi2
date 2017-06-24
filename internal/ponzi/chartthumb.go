package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartThumbnail struct {
	stock             *modelStock
	stockQuoteText    *dynamicText
	lines             *chartLines
	frameBorder       *gfx.VAO
	frameDivider      *gfx.VAO
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createchartThumbnail(stock *modelStock, stockQuoteText *dynamicText) *chartThumbnail {
	return &chartThumbnail{
		stock:             stock,
		stockQuoteText:    stockQuoteText,
		lines:             createChartLines(stock),
		frameBorder:       gfx.CreateStrokedRectVAO(white, white, white, white),
		frameDivider:      gfx.CreateLineVAO(white, white),
		dailyStochastics:  createChartStochastics(stock, stockQuoteText, daily),
		weeklyStochastics: createChartStochastics(stock, stockQuoteText, weekly),
	}
}

func (ct *chartThumbnail) update() {
	ct.lines.update()
	ct.dailyStochastics.update()
	ct.weeklyStochastics.update()
}

func (ct *chartThumbnail) render(r image.Rectangle) {
	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ct.frameBorder.Render()

	//
	// Render the symbol and its quote.
	//

	const pad = 5
	s := ct.stockQuoteText.measure(ct.stock.symbol)
	pt.Y -= pad + s.Y
	{
		pt := pt
		pt.X += pad
		pt = pt.Add(ct.stockQuoteText.render(ct.stock.symbol, pt, white))
		pt = pt.Add(ct.stockQuoteText.render(shortFormatQuote(ct.stock.quote), pt, quoteColor(ct.stock.quote)))
	}

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = pt.Y
	gl.Uniform1f(colorMixAmountLocation, 1)

	rects := sliceRectangle(r, 0.5, 0.5)
	for _, r := range rects {
		setModelMatrixRectangle(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		ct.frameDivider.Render()
	}

	//
	// Render the graphs.
	//

	ct.dailyStochastics.render(rects[1].Inset(pad))
	ct.weeklyStochastics.render(rects[0].Inset(pad))
	ct.lines.render(rects[1].Inset(pad))
	ct.lines.render(rects[0].Inset(pad))
}

func (ct *chartThumbnail) close() {
	ct.lines.close()
	ct.lines = nil
	ct.frameBorder.Close()
	ct.frameBorder = nil
	ct.frameDivider.Close()
	ct.frameDivider = nil
	ct.dailyStochastics.close()
	ct.dailyStochastics = nil
	ct.weeklyStochastics.close()
	ct.weeklyStochastics = nil
}
