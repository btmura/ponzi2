package view

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/button"
	"github.com/btmura/ponzi2/internal/status"
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

	// quotePrinter is the function used to generate the quote text.
	quotePrinter func(*model.Quote) string

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
	fadeIn *animation.Animation

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

// chartHeaderButton is a button with an additional enabled flag.
type chartHeaderButton struct {
	// Button is the underlying button.
	*button.Button

	// enabled is whether the button is present and clickable.
	enabled bool
}

// chartHeaderArgs are passed to newChartHeader.
type chartHeaderArgs struct {
	SymbolQuoteTextRenderer *gfx.TextRenderer
	QuotePrinter            func(*model.Quote) string
	ShowRefreshButton       bool
	ShowAddButton           bool
	ShowRemoveButton        bool
	Rounding                int
	Padding                 int
}

func newChartHeader(args *chartHeaderArgs) *chartHeader {
	return &chartHeader{
		symbolQuoteTextRenderer: args.SymbolQuoteTextRenderer,
		quotePrinter:            args.QuotePrinter,
		refreshButton: &chartHeaderButton{
			Button:  button.New(chartRefreshButtonVAO, fps),
			enabled: args.ShowRefreshButton,
		},
		addButton: &chartHeaderButton{
			Button:  button.New(chartAddButtonVAO, fps),
			enabled: args.ShowAddButton,
		},
		removeButton: &chartHeaderButton{
			Button:  button.New(chartRemoveButtonVAO, fps),
			enabled: args.ShowRemoveButton,
		},
		rounding: args.Rounding,
		padding:  args.Padding,
		fadeIn:   animation.New(1 * fps),
	}
}

// SetLoading toggles the Chart's loading indicator.
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

// SetError toggles the Chart's error indicator.
func (ch *chartHeader) SetError(error bool) {
	ch.hasError = error
}

// SetData sets the data to be shown on the chart.
func (ch *chartHeader) SetData(data *ChartData) error {
	if data == nil {
		return status.Error("missing data")
	}

	if !ch.hasStockUpdated && data.Chart != nil {
		ch.fadeIn.Start()
	}
	ch.hasStockUpdated = data.Chart != nil

	ch.symbol = data.Symbol

	ch.quoteText = ch.quotePrinter(data.Quote)

	var c float32
	if q := data.Quote; q != nil {
		c = q.ChangePercent
	}

	switch {
	case c > 0:
		ch.quoteColor = green

	case c < 0:
		ch.quoteColor = red

	default:
		ch.quoteColor = white
	}

	return nil
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

// ProcessInput processes input.
func (ch *chartHeader) ProcessInput(ic inputContext) (body image.Rectangle, clicks chartHeaderClicks) {
	ch.bounds = ic.Bounds

	h := ch.padding + ch.symbolQuoteTextRenderer.LineHeight() + ch.padding
	buttonSize := image.Pt(h, h)

	// Render buttons in the upper right corner from right to left.
	r := ic.Bounds
	ic.Bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}

	if ch.removeButton.enabled {
		clicks.RemoveButtonClicked = ch.removeButton.ProcessInput(ic.Bounds, ic.MousePos, ic.MouseLeftButtonReleased, ic.ScheduledCallbacks)
		ic.Bounds = transRect(ic.Bounds, -buttonSize.X, 0)
	}

	if ch.addButton.enabled {
		clicks.AddButtonClicked = ch.addButton.ProcessInput(ic.Bounds, ic.MousePos, ic.MouseLeftButtonReleased, ic.ScheduledCallbacks)
		ic.Bounds = transRect(ic.Bounds, -buttonSize.X, 0)
	}

	if ch.refreshButton.enabled || ch.refreshButton.Spinning() {
		clicks.RefreshButtonClicked = ch.refreshButton.ProcessInput(ic.Bounds, ic.MousePos, ic.MouseLeftButtonReleased, ic.ScheduledCallbacks)
		ic.Bounds = transRect(ic.Bounds, -buttonSize.X, 0)
	}

	// Don't report clicks when the refresh button is just an indicator.
	if !ch.refreshButton.enabled {
		clicks.RefreshButtonClicked = false
	}

	r.Max.Y -= ch.padding + ch.symbolQuoteTextRenderer.LineHeight() + ch.padding

	return r, clicks
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

// Render renders the ChartHeader.
func (ch *chartHeader) Render(fudge float32) {
	h := ch.padding + ch.symbolQuoteTextRenderer.LineHeight() + ch.padding
	buttonSize := image.Pt(h, h)

	// Render buttons in the upper right corner from right to left.
	r := ch.bounds
	ch.bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}

	if ch.removeButton.enabled {
		ch.removeButton.Render(fudge)
		ch.bounds = transRect(ch.bounds, -buttonSize.X, 0)
	}

	if ch.addButton.enabled {
		ch.addButton.Render(fudge)
		ch.bounds = transRect(ch.bounds, -buttonSize.X, 0)
	}

	if ch.refreshButton.enabled || ch.refreshButton.Spinning() {
		ch.refreshButton.Render(fudge)
		ch.bounds = transRect(ch.bounds, -buttonSize.X, 0)
	}

	if ch.hasError {
		gfx.SetModelMatrixRect(ch.bounds)
		chartErrorIconVAO.Render()
		ch.bounds = transRect(ch.bounds, -buttonSize.X, 0)
	}

	buttonEdge := ch.bounds.Min.X + buttonSize.X

	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)
	pt.Y -= ch.padding + ch.symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += ch.rounding
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.symbol, pt, white)
		pt.X += ch.padding

		if w := buttonEdge - pt.X; w > 0 {
			old := gfx.Alpha()
			gfx.SetAlpha(old * ch.fadeIn.Value(fudge))
			pt.X += ch.symbolQuoteTextRenderer.Render(ch.quoteText, pt, ch.quoteColor, gfx.TextRenderMaxWidth(w))
			gfx.SetAlpha(old)
		}
	}
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
func (ch *chartHeader) Close() {
	ch.refreshButton.Close()
	ch.addButton.Close()
	ch.removeButton.Close()
}
