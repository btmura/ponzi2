package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartHeader shows a header for charts and thumbnails with a clickable button.
type ChartHeader struct {
	symbol                  string
	quoteText               string
	quoteColor              [3]float32
	symbolQuoteTextRenderer *gfx.TextRenderer
	quoteFormatter          func(*ModelStock) string
	button                  *Button
	roundAmount             int
	padding                 int
	buttonClickCallbacks    []func()
}

// NewChartHeader creates a new chart header.
func NewChartHeader(symbolQuoteTextRenderer *gfx.TextRenderer, quoteFormatter func(*ModelStock) string, button *Button, roundAmount, padding int) *ChartHeader {
	return &ChartHeader{
		symbolQuoteTextRenderer: symbolQuoteTextRenderer,
		quoteFormatter:          quoteFormatter,
		button:                  button,
		roundAmount:             roundAmount,
		padding:                 padding,
	}
}

// Update updates the ChartHeader with the given stock.
func (ch *ChartHeader) Update(st *ModelStock) {
	ch.symbol = st.Symbol
	ch.quoteText = ch.quoteFormatter(st)
	switch {
	case st.PercentChange() > 0:
		ch.quoteColor = green

	case st.PercentChange() < 0:
		ch.quoteColor = red

	default:
		ch.quoteColor = white
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
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.symbol, pt, white)
		pt.X += ch.padding
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.quoteText, pt, ch.quoteColor)
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

// Close frees the resources backing the ChartHeader.
func (ch *ChartHeader) Close() {}
