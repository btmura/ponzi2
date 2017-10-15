package ponzi

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

var (
	chartAddButtonVAO     = gfx.SquareImageVAO(bytes.NewReader(MustAsset("addButton.png")))
	chartRefreshButtonVAO = gfx.SquareImageVAO(bytes.NewReader(MustAsset("refreshButton.png")))
	chartRemoveButtonVAO  = gfx.SquareImageVAO(bytes.NewReader(MustAsset("removeButton.png")))
)

// ChartHeader shows a header for charts and thumbnails with a clickable button.
type ChartHeader struct {
	// symbol is the symbol to render.
	symbol string

	// quoteText is the text with the price information.
	quoteText string

	// quoteColor is the color to render the quote text.
	quoteColor [3]float32

	// symbolQuoteTextRenderer renders the symbol and quote text.
	symbolQuoteTextRenderer *gfx.TextRenderer

	// quoteFormatter is the function used to generate the quote text.
	quoteFormatter func(*ModelStock) string

	// refreshButton is the optional button to refresh the chart.
	refreshButton *Button

	// addButton is the optional button to add the symbol.
	addButton *Button

	// removeButton is the optional button to remove the symbol.
	removeButton *Button

	// rounding is only used to layout the symbol and quote text.
	rounding int

	// padding is only used to layout the symbol and quote text.
	padding int

	// loading is whether the data for the symbol is loading.
	loading bool
}

// ChartHeaderArgs are passed to NewChartHeader.
type ChartHeaderArgs struct {
	SymbolQuoteTextRenderer *gfx.TextRenderer
	QuoteFormatter          func(*ModelStock) string
	RefreshButton           bool
	AddButton               bool
	RemoveButton            bool
	Rounding                int
	Padding                 int
}

// NewChartHeader creates a new chart header.
func NewChartHeader(args *ChartHeaderArgs) *ChartHeader {
	ch := &ChartHeader{
		symbolQuoteTextRenderer: args.SymbolQuoteTextRenderer,
		quoteFormatter:          args.QuoteFormatter,
		rounding:                args.Rounding,
		padding:                 args.Padding,
	}
	if args.RefreshButton {
		ch.refreshButton = NewButton(chartRefreshButtonVAO)
	}
	if args.AddButton {
		ch.addButton = NewButton(chartAddButtonVAO)
	}
	if args.RemoveButton {
		ch.removeButton = NewButton(chartRemoveButtonVAO)
	}
	return ch
}

// SetState sets the ChartHeader's state.
func (ch *ChartHeader) SetState(state *ChartState) {
	if ch.refreshButton != nil {
		switch {
		// Not Loading -> Loading
		case !ch.loading && state.Loading:
			ch.refreshButton.StartSpinning()

		// Loading -> Not Loading
		case ch.loading && !state.Loading:
			ch.refreshButton.StopSpinning()
		}
	}
	ch.loading = state.Loading

	ch.symbol = state.Stock.Symbol
	ch.quoteText = ch.quoteFormatter(state.Stock)

	c := state.Stock.PercentChange()
	switch {
	case c > 0:
		ch.quoteColor = green

	case c < 0:
		ch.quoteColor = red

	default:
		ch.quoteColor = white
	}
}

// Update updates the ChartHeader.
func (ch *ChartHeader) Update() {
	if ch.refreshButton != nil {
		ch.refreshButton.Update()
	}
	if ch.addButton != nil {
		ch.addButton.Update()
	}
	if ch.removeButton != nil {
		ch.removeButton.Update()
	}
}

// Render renders the ChartHeader.
func (ch *ChartHeader) Render(vc ViewContext) (body image.Rectangle, addButtonClicked, refreshButtonClicked, removeButtonClicked bool) {
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

	// Render buttons in the upper right corner from right to left.
	buttonSize := image.Pt(r.Max.Y-pt.Y, r.Max.Y-pt.Y)
	vc.Bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}

	if ch.removeButton != nil {
		removeButtonClicked = ch.removeButton.Render(vc)
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	if ch.addButton != nil {
		addButtonClicked = ch.addButton.Render(vc)
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	if ch.refreshButton != nil {
		refreshButtonClicked = ch.refreshButton.Render(vc)
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	r.Max.Y = pt.Y

	return r, addButtonClicked, refreshButtonClicked, removeButtonClicked
}

// SetRefreshButtonClickCallback sets the callback for refresh button clicks.
func (ch *ChartHeader) SetRefreshButtonClickCallback(cb func()) {
	ch.refreshButton.SetClickCallback(cb)
}

// SetAddButtonClickCallback sets the callback for add button clicks.
func (ch *ChartHeader) SetAddButtonClickCallback(cb func()) {
	ch.addButton.SetClickCallback(cb)
}

// SetRemoveButtonClickCallback sets the callback for remove button clicks.
func (ch *ChartHeader) SetRemoveButtonClickCallback(cb func()) {
	ch.removeButton.SetClickCallback(cb)
}

// Close frees the resources backing the ChartHeader.
func (ch *ChartHeader) Close() {}
