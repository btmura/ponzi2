package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartHeader shows a header for charts and thumbnails with a clickable button.
type ChartHeader struct {
	stock                   *ModelStock
	symbolQuoteTextRenderer *gfx.TextRenderer
	quoteFormatter          func(*ModelQuote) string
	button                  *Button
	roundAmount             int
	padding                 int
	buttonClickCallbacks    []func()
}

// NewChartHeader creates a new chart header.
func NewChartHeader(stock *ModelStock, symbolQuoteTextRenderer *gfx.TextRenderer, quoteFormatter func(*ModelQuote) string, button *Button, roundAmount, padding int) *ChartHeader {
	return &ChartHeader{
		stock: stock,
		symbolQuoteTextRenderer: symbolQuoteTextRenderer,
		quoteFormatter:          quoteFormatter,
		button:                  button,
		roundAmount:             roundAmount,
		padding:                 padding,
	}
}

// Render renders the chart header.
func (ch *ChartHeader) Render(vc ViewContext) (body image.Rectangle, buttonClicked bool) {
	// Render the border around the chart.
	r := vc.Bounds
	renderRoundedRect(r, ch.roundAmount)

	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)
	pt.Y -= ch.padding + ch.symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += ch.roundAmount
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.stock.symbol, pt, white)
		pt.X += ch.padding
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.quoteFormatter(ch.stock.quote), pt, quoteColor(ch.stock.quote))
	}
	pt.Y -= ch.padding

	// Render button in the upper right corner.
	buttonSize := image.Pt(r.Max.Y-pt.Y, r.Max.Y-pt.Y)
	vc.Bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}
	buttonClicked = ch.button.Render(vc)

	r.Max.Y = pt.Y
	return r, buttonClicked
}

// AddButtonClickCallback adds a callback for when the button is clicked.
func (ch *ChartHeader) AddButtonClickCallback(cb func()) {
	ch.button.AddClickCallback(cb)
}

func quoteColor(q *ModelQuote) [3]float32 {
	switch {
	case q.percentChange > 0:
		return green

	case q.percentChange < 0:
		return red
	}
	return white
}
