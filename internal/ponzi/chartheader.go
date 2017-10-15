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

	// refreshButton is the button to refresh the chart.
	refreshButton *chartHeaderButton

	// addButton is the button to add the symbol.
	addButton *chartHeaderButton

	// removeButton is the button to remove the symbol.
	removeButton *chartHeaderButton

	// rounding is only used to layout the symbol and quote text.
	rounding int

	// padding is only used to layout the symbol and quote text.
	padding int

	// loading is whether the data for the symbol is loading.
	loading bool
}

// chartHeaderButton is a button with an additional enabled flag.
type chartHeaderButton struct {
	// Button is the underlying button.
	*Button

	// enabled is whether the button is present and clickable.
	enabled bool
}

// ChartHeaderArgs are passed to NewChartHeader.
type ChartHeaderArgs struct {
	SymbolQuoteTextRenderer *gfx.TextRenderer
	QuoteFormatter          func(*ModelStock) string
	ShowRefreshButton       bool
	ShowAddButton           bool
	ShowRemoveButton        bool
	Rounding                int
	Padding                 int
}

// NewChartHeader creates a new chart header.
func NewChartHeader(args *ChartHeaderArgs) *ChartHeader {
	return &ChartHeader{
		symbolQuoteTextRenderer: args.SymbolQuoteTextRenderer,
		quoteFormatter:          args.QuoteFormatter,
		refreshButton: &chartHeaderButton{
			Button:  NewButton(chartRefreshButtonVAO),
			enabled: args.ShowRefreshButton,
		},
		addButton: &chartHeaderButton{
			Button:  NewButton(chartAddButtonVAO),
			enabled: args.ShowAddButton,
		},
		removeButton: &chartHeaderButton{
			Button:  NewButton(chartRemoveButtonVAO),
			enabled: args.ShowRemoveButton,
		},
		rounding: args.Rounding,
		padding:  args.Padding,
	}
}

// SetLoading sets the ChartHeader's loading state.
func (ch *ChartHeader) SetLoading(loading bool) {
	switch {
	// Not Loading -> Loading
	case !ch.loading && loading:
		ch.refreshButton.StartSpinning()

	// Loading -> Not Loading
	case ch.loading && !loading:
		ch.refreshButton.StopSpinning()
	}
	ch.loading = loading
}

// SetStock sets the ChartHeader's stock.
func (ch *ChartHeader) SetStock(st *ModelStock) {
	ch.symbol = st.Symbol
	ch.quoteText = ch.quoteFormatter(st)

	c := st.PercentChange()
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
	ch.refreshButton.Update()
	ch.addButton.Update()
	ch.removeButton.Update()
}

// ChartHeaderClicks reports what buttons were clicked.
type ChartHeaderClicks struct {
	// AddButtonClicked is true if the add button was clicked.
	AddButtonClicked bool

	// RefreshButtonClicked is true if the refresh button was clicked.
	RefreshButtonClicked bool

	// RemoveButtonClicked is true if the remove button was clicked.
	RemoveButtonClicked bool
}

// HasClicks returns true if a clickable part of the header was clicked.
func (c ChartHeaderClicks) HasClicks() bool {
	return c.AddButtonClicked || c.RefreshButtonClicked || c.RemoveButtonClicked
}

// Render renders the ChartHeader.
func (ch *ChartHeader) Render(vc ViewContext) (body image.Rectangle, clicks ChartHeaderClicks) {
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

	if ch.removeButton.enabled {
		clicks.RemoveButtonClicked = ch.removeButton.Render(vc)
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	if ch.addButton.enabled {
		clicks.AddButtonClicked = ch.addButton.Render(vc)
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	if ch.refreshButton.enabled || ch.loading {
		clicks.RefreshButtonClicked = ch.refreshButton.Render(vc)
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	// Don't report clicks when the refresh button is just an indicator.
	if !ch.refreshButton.enabled {
		clicks.RefreshButtonClicked = false
	}

	r.Max.Y = pt.Y

	return r, clicks
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
