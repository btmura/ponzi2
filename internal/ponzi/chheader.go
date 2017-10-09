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
	button1                 *Button
	button2                 *Button
	rounding                int
	padding                 int
	loading                 bool
}

// ChartHeaderArgs are passed to NewChartHeader.
type ChartHeaderArgs struct {
	SymbolQuoteTextRenderer *gfx.TextRenderer
	QuoteFormatter          func(*ModelStock) string
	Button1                 *Button
	Button2                 *Button
	Rounding                int
	Padding                 int
}

// NewChartHeader creates a new chart header.
func NewChartHeader(args *ChartHeaderArgs) *ChartHeader {
	return &ChartHeader{
		symbolQuoteTextRenderer: args.SymbolQuoteTextRenderer,
		quoteFormatter:          args.QuoteFormatter,
		button1:                 args.Button1,
		button2:                 args.Button2,
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
func (ch *ChartHeader) Render(vc ViewContext) (body image.Rectangle, button1Clicked, button2Clicked bool) {
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
		return r, false /* no left button click */, false /* no right button click */
	}

	// Render buttons in the upper right corner.
	buttonSize := image.Pt(r.Max.Y-pt.Y, r.Max.Y-pt.Y)
	vc.Bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}

	if ch.button2 != nil {
		button2Clicked = ch.button2.Render(vc)
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	if ch.button1 != nil {
		button1Clicked = ch.button1.Render(vc)
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	r.Max.Y = pt.Y
	return r, button1Clicked, button2Clicked
}

// SetButton1ClickCallback sets the callback for button 1 clicks.
func (ch *ChartHeader) SetButton1ClickCallback(cb func()) {
	ch.button1.SetClickCallback(cb)
}

// SetButton2ClickCallback sets the callback for button 2 clicks.
func (ch *ChartHeader) SetButton2ClickCallback(cb func()) {
	ch.button2.SetClickCallback(cb)
}

// Close frees the resources backing the ChartHeader.
func (ch *ChartHeader) Close() {}
