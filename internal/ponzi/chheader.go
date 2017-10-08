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
	rounding                int
	padding                 int
	buttonClickCallbacks    []func()
	loading                 bool
}

// ChartHeaderArgs are passed to NewChartHeader.
type ChartHeaderArgs struct {
	SymbolQuoteTextRenderer *gfx.TextRenderer
	QuoteFormatter          func(*ModelStock) string
	Button                  *Button
	Rounding                int
	Padding                 int
}

// NewChartHeader creates a new chart header.
func NewChartHeader(args *ChartHeaderArgs) *ChartHeader {
	return &ChartHeader{
		symbolQuoteTextRenderer: args.SymbolQuoteTextRenderer,
		quoteFormatter:          args.QuoteFormatter,
		button:                  args.Button,
		rounding:                args.Rounding,
		padding:                 args.Padding,
	}
}

// Update updates the ChartHeader.
func (ch *ChartHeader) Update(u *ChartUpdate) {
	ch.loading = u.Loading
	ch.symbol = u.Stock.Symbol
	ch.quoteText = ch.quoteFormatter(u.Stock)
	switch {
	case u.Stock.PercentChange() > 0:
		ch.quoteColor = green

	case u.Stock.PercentChange() < 0:
		ch.quoteColor = red

	default:
		ch.quoteColor = white
	}
}

// Render renders the chart header.
func (ch *ChartHeader) Render(vc ViewContext) (body image.Rectangle, buttonClicked bool) {
	// Start rendering from the top left. Track position with point.
	r := vc.Bounds
	pt := image.Pt(r.Min.X, r.Max.Y)
	pt.Y -= ch.padding + ch.symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += ch.rounding
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.symbol, pt, white)
		pt.X += ch.padding
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.quoteText, pt, ch.quoteColor)
	}
	pt.Y -= ch.padding

	if ch.loading {
		r.Max.Y = pt.Y
		return r, false /* no button click */
	}

	// Render button in the upper right corner.
	buttonSize := image.Pt(r.Max.Y-pt.Y, r.Max.Y-pt.Y)
	vc.Bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}
	buttonClicked = ch.button.Render(vc)

	r.Max.Y = pt.Y
	return r, buttonClicked
}

// SetButtonClickCallback sets the callback for when the button is clicked.
func (ch *ChartHeader) SetButtonClickCallback(cb func()) {
	ch.button.SetClickCallback(cb)
}

// Close frees the resources backing the ChartHeader.
func (ch *ChartHeader) Close() {}
