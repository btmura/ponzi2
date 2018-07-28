package view

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/gfx"
)

var (
	chartAddButtonVAO     = texturedSquareVAO(bytes.NewReader(_escFSMustByte(false, "/data/addbutton.png")))
	chartRefreshButtonVAO = texturedSquareVAO(bytes.NewReader(_escFSMustByte(false, "/data/refreshbutton.png")))
	chartRemoveButtonVAO  = texturedSquareVAO(bytes.NewReader(_escFSMustByte(false, "/data/removebutton.png")))
	chartErrorIconVAO     = texturedSquareVAO(bytes.NewReader(_escFSMustByte(false, "/data/erroricon.png")))
)

// chartHeader shows a header for charts and thumbnails with a clickable button.
type chartHeader struct {
	// symbol is the symbol to render.
	symbol string

	// quoteText is the text with the price information.
	quoteText string

	// quoteColor is the color to render the quote text.
	quoteColor [3]float32

	// symbolQuoteTextRenderer renders the symbol and quote text.
	symbolQuoteTextRenderer *gfx.TextRenderer

	// quoteFormatter is the function used to generate the quote text.
	quoteFormatter func(*model.Stock) string

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

	// hasStockUpdated is true if the stock has reported data.
	hasStockUpdated bool

	// hasError is true there was a loading issue.
	hasError bool

	// fadeIn fades in the quote text after data loads.
	fadeIn *animation
}

// chartHeaderButton is a button with an additional enabled flag.
type chartHeaderButton struct {
	// Button is the underlying button.
	*button

	// enabled is whether the button is present and clickable.
	enabled bool
}

// chartHeaderArgs are passed to newChartHeader.
type chartHeaderArgs struct {
	SymbolQuoteTextRenderer *gfx.TextRenderer
	QuoteFormatter          func(*model.Stock) string
	ShowRefreshButton       bool
	ShowAddButton           bool
	ShowRemoveButton        bool
	Rounding                int
	Padding                 int
}

func newChartHeader(args *chartHeaderArgs) *chartHeader {
	return &chartHeader{
		symbolQuoteTextRenderer: args.SymbolQuoteTextRenderer,
		quoteFormatter:          args.QuoteFormatter,
		refreshButton: &chartHeaderButton{
			button:  newButton(chartRefreshButtonVAO),
			enabled: args.ShowRefreshButton,
		},
		addButton: &chartHeaderButton{
			button:  newButton(chartAddButtonVAO),
			enabled: args.ShowAddButton,
		},
		removeButton: &chartHeaderButton{
			button:  newButton(chartRemoveButtonVAO),
			enabled: args.ShowRemoveButton,
		},
		rounding: args.Rounding,
		padding:  args.Padding,
		fadeIn:   newAnimation(1*fps, false),
	}
}

func (ch *chartHeader) SetLoading(loading bool) {
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

func (ch *chartHeader) SetError(error bool) {
	ch.hasError = error
}

func (ch *chartHeader) SetData(st *model.Stock) {
	if !ch.hasStockUpdated && !st.LastUpdateTime.IsZero() {
		ch.fadeIn.Start()
	}
	ch.hasStockUpdated = !st.LastUpdateTime.IsZero()

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

func (ch *chartHeader) Update() (dirty bool) {
	if ch.refreshButton.Update() {
		dirty = true
	}
	if ch.addButton.Update() {
		dirty = true
	}
	if ch.removeButton.Update() {
		dirty = true
	}
	if ch.fadeIn.Update() {
		dirty = true
	}
	return dirty
}

// chartHeaderClicks reports what buttons were clicked.
type chartHeaderClicks struct {
	// AddButtonClicked is true if the add button was clicked.
	AddButtonClicked bool

	// RefreshButtonClicked is true if the refresh button was clicked.
	RefreshButtonClicked bool

	// RemoveButtonClicked is true if the remove button was clicked.
	RemoveButtonClicked bool
}

// HasClicks returns true if a clickable part of the header was clicked.
func (c chartHeaderClicks) HasClicks() bool {
	return c.AddButtonClicked || c.RefreshButtonClicked || c.RemoveButtonClicked
}

// Render renders the ChartHeader.
func (ch *chartHeader) Render(vc viewContext) (body image.Rectangle, clicks chartHeaderClicks) {
	// Start rendering from the top left. Track position with point.
	r := vc.Bounds
	pt := image.Pt(r.Min.X, r.Max.Y)
	pt.Y -= ch.padding + ch.symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += ch.rounding
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.symbol, pt, white)
		pt.X += ch.padding

		gfx.SetAlpha(ch.fadeIn.Value(vc.Fudge))
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.quoteText, pt, ch.quoteColor)
		gfx.SetAlpha(1)

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

	if ch.hasError {
		gfx.SetModelMatrixRect(vc.Bounds)
		chartErrorIconVAO.Render()
		vc.Bounds = transRect(vc.Bounds, -buttonSize.X, 0)
	}

	r.Max.Y = pt.Y

	return r, clicks
}

// SetRefreshButtonClickCallback sets the callback for refresh button clicks.
func (ch *chartHeader) SetRefreshButtonClickCallback(cb func()) {
	ch.refreshButton.SetClickCallback(cb)
}

// SetAddButtonClickCallback sets the callback for add button clicks.
func (ch *chartHeader) SetAddButtonClickCallback(cb func()) {
	ch.addButton.SetClickCallback(cb)
}

// SetRemoveButtonClickCallback sets the callback for remove button clicks.
func (ch *chartHeader) SetRemoveButtonClickCallback(cb func()) {
	ch.removeButton.SetClickCallback(cb)
}

// Close frees the resources backing the ChartHeader.
func (ch *chartHeader) Close() {}
