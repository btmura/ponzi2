package chart

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/button"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

var (
	addButtonVAO         = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/addbutton.png")))
	barButtonVAO         = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/barbutton.png")))
	candlestickButtonVAO = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/candlestickbutton.png")))
	errorIconVAO         = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/erroricon.png")))
	refreshButtonVAO     = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/refreshbutton.png")))
	removeButtonVAO      = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/removebutton.png")))
)

// header shows a header for charts and thumbnails with a clickable button.
type header struct {
	// symbol is the symbol to render.
	symbol string

	// quoteText is the text with the price information.
	quoteText string

	// quoteColor is the color to render the quote text.
	quoteColor view.Color

	// symbolQuoteTextRenderer renders the symbol and quote text.
	symbolQuoteTextRenderer *gfx.TextRenderer

	// quotePrinter is the function used to generate the quote text.
	quotePrinter func(*model.Quote) string

	// barButton is the button to show price bars.
	barButton *headerButton

	// candlestickButton is the button to show price candlesticks.
	candlestickButton *headerButton

	// refreshButton is the button to refresh the chart.
	refreshButton *headerButton

	// addButton is the button to add the symbol.
	addButton *headerButton

	// removeButton is the button to remove the symbol.
	removeButton *headerButton

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

// headerButton is a button with an additional enabled flag.
type headerButton struct {
	// Button is the underlying button.
	*button.Button

	// enabled is whether the button is present and clickable.
	enabled bool
}

// headerArgs are passed to newChartHeader.
type headerArgs struct {
	SymbolQuoteTextRenderer *gfx.TextRenderer
	QuotePrinter            func(*model.Quote) string
	ShowBarButton           bool
	ShowCandlestickButton   bool
	ShowRefreshButton       bool
	ShowAddButton           bool
	ShowRemoveButton        bool
	Rounding                int
	Padding                 int
	FPS                     int
}

func newHeader(args *headerArgs) *header {
	return &header{
		symbolQuoteTextRenderer: args.SymbolQuoteTextRenderer,
		quotePrinter:            args.QuotePrinter,
		barButton: &headerButton{
			Button:  button.New(barButtonVAO, args.FPS),
			enabled: args.ShowBarButton,
		},
		candlestickButton: &headerButton{
			Button:  button.New(candlestickButtonVAO, args.FPS),
			enabled: args.ShowCandlestickButton,
		},
		refreshButton: &headerButton{
			Button:  button.New(refreshButtonVAO, args.FPS),
			enabled: args.ShowRefreshButton,
		},
		addButton: &headerButton{
			Button:  button.New(addButtonVAO, args.FPS),
			enabled: args.ShowAddButton,
		},
		removeButton: &headerButton{
			Button:  button.New(removeButtonVAO, args.FPS),
			enabled: args.ShowRemoveButton,
		},
		rounding: args.Rounding,
		padding:  args.Padding,
		fadeIn:   animation.New(1 * args.FPS),
	}
}

// SetLoading toggles the Chart's loading indicator.
func (h *header) SetLoading(loading bool) {
	switch {
	// Not Loading -> Loading
	case !h.loading && loading:
		h.refreshButton.StartSpinning()

	// Loading -> Not Loading
	case h.loading && !loading:
		h.refreshButton.StopSpinning()
	}
	h.loading = loading
}

// SetError toggles the Chart's error indicator.
func (h *header) SetError(err error) {
	h.hasError = err != nil
}

// SetData sets the data to be shown on the chart.
func (h *header) SetData(data Data) {
	if !h.hasStockUpdated && data.Chart != nil {
		h.fadeIn.Start()
	}
	h.hasStockUpdated = data.Chart != nil

	h.symbol = data.Symbol

	h.quoteText = h.quotePrinter(data.Quote)

	var c float32
	if q := data.Quote; q != nil {
		c = q.ChangePercent
	}

	switch {
	case c > 0:
		h.quoteColor = view.Green

	case c < 0:
		h.quoteColor = view.Red

	default:
		h.quoteColor = view.White
	}
}

// headerClicks reports what buttons were clicked.
type headerClicks struct {
	// BarButtonClicked is true if the bar button was clicked.
	BarButtonClicked bool

	// CandlestickButtonClicked is true if the candlestick button wan clicked.
	CandlestickButtonClicked bool

	// AddButtonClicked is true if the add button was clicked.
	AddButtonClicked bool

	// RefreshButtonClicked is true if the refresh button was clicked.
	RefreshButtonClicked bool

	// RemoveButtonClicked is true if the remove button was clicked.
	RemoveButtonClicked bool
}

// HasClicks returns true if a clickable part of the header was clicked.
func (c headerClicks) HasClicks() bool {
	return c.BarButtonClicked ||
		c.CandlestickButtonClicked ||
		c.AddButtonClicked ||
		c.RefreshButtonClicked ||
		c.RemoveButtonClicked
}

func (h *header) SetBounds(bounds image.Rectangle) {
	h.bounds = bounds
}

func (h *header) ProcessInput(input *view.Input) (body image.Rectangle, clicks headerClicks) {
	height := h.padding + h.symbolQuoteTextRenderer.LineHeight() + h.padding
	buttonSize := image.Pt(height, height)

	// Render buttons in the upper right corner from right to left.
	r := h.bounds
	bounds := image.Rectangle{Min: r.Max.Sub(buttonSize), Max: r.Max}

	if h.removeButton.enabled {
		h.removeButton.SetBounds(bounds)
		clicks.RemoveButtonClicked = h.removeButton.ProcessInput(input)
		bounds = rect.Translate(bounds, -buttonSize.X, 0)
	}

	if h.addButton.enabled {
		h.addButton.SetBounds(bounds)
		clicks.AddButtonClicked = h.addButton.ProcessInput(input)
		bounds = rect.Translate(bounds, -buttonSize.X, 0)
	}

	if h.refreshButton.enabled || h.refreshButton.Spinning() {
		h.refreshButton.SetBounds(bounds)
		clicks.RefreshButtonClicked = h.refreshButton.ProcessInput(input)
		bounds = rect.Translate(bounds, -buttonSize.X, 0)
	}

	if h.candlestickButton.enabled {
		h.candlestickButton.SetBounds(bounds)
		clicks.RemoveButtonClicked = h.candlestickButton.ProcessInput(input)
		bounds = rect.Translate(bounds, -buttonSize.X, 0)
	}

	if h.barButton.enabled {
		h.barButton.SetBounds(bounds)
		clicks.RemoveButtonClicked = h.barButton.ProcessInput(input)
		bounds = rect.Translate(bounds, -buttonSize.X, 0)
	}

	// Don't report clicks when the refresh button is just an indicator.
	if !h.refreshButton.enabled {
		clicks.RefreshButtonClicked = false
	}

	r.Max.Y -= h.padding + h.symbolQuoteTextRenderer.LineHeight() + h.padding

	return r, clicks
}

func (h *header) Update() (dirty bool) {
	if h.barButton.Update() {
		dirty = true
	}
	if h.candlestickButton.Update() {
		dirty = true
	}
	if h.refreshButton.Update() {
		dirty = true
	}
	if h.addButton.Update() {
		dirty = true
	}
	if h.removeButton.Update() {
		dirty = true
	}
	if h.fadeIn.Update() {
		dirty = true
	}
	return dirty
}

// Render renders the ChartHeader.
func (h *header) Render(fudge float32) {
	height := h.padding + h.symbolQuoteTextRenderer.LineHeight() + h.padding
	buttonSize := image.Pt(height, height)

	// Render buttons in the upper right corner from right to left.
	r := h.bounds
	h.bounds = image.Rectangle{Min: r.Max.Sub(buttonSize), Max: r.Max}

	if h.removeButton.enabled {
		h.removeButton.Render(fudge)
		h.bounds = rect.Translate(h.bounds, -buttonSize.X, 0)
	}

	if h.addButton.enabled {
		h.addButton.Render(fudge)
		h.bounds = rect.Translate(h.bounds, -buttonSize.X, 0)
	}

	if h.refreshButton.enabled || h.refreshButton.Spinning() {
		h.refreshButton.Render(fudge)
		h.bounds = rect.Translate(h.bounds, -buttonSize.X, 0)
	}

	if h.candlestickButton.enabled {
		h.candlestickButton.Render(fudge)
		h.bounds = rect.Translate(h.bounds, -buttonSize.X, 0)
	}

	if h.barButton.enabled {
		h.barButton.Render(fudge)
		h.bounds = rect.Translate(h.bounds, -buttonSize.X, 0)
	}

	if h.hasError {
		gfx.SetModelMatrixRect(h.bounds)
		errorIconVAO.Render()
		h.bounds = rect.Translate(h.bounds, -buttonSize.X, 0)
	}

	buttonEdge := h.bounds.Min.X + buttonSize.X

	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)
	pt.Y -= h.padding + h.symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += h.rounding
		pt.X += h.symbolQuoteTextRenderer.Render(h.symbol, pt, gfx.TextColor(view.White))
		pt.X += h.padding

		if w := buttonEdge - pt.X; w > 0 {
			old := gfx.Alpha()
			gfx.SetAlpha(old * h.fadeIn.Value(fudge))
			pt.X += h.symbolQuoteTextRenderer.Render(h.quoteText, pt, gfx.TextColor(h.quoteColor), gfx.TextRenderMaxWidth(w))
			gfx.SetAlpha(old)
		}
	}
}

// SetBarButtonClickCallback sets the callback for bar button clicks.
func (h *header) SetBarButtonClickCallback(cb func()) {
	h.barButton.SetClickCallback(cb)
}

// SetCandlestickButtonClickCallback sets the callback for candlestick clicks.
func (h *header) SetCandlestickButtonClickCallback(cb func()) {
	h.candlestickButton.SetClickCallback(cb)
}

// SetRefreshButtonClickCallback sets the callback for refresh button clicks.
func (h *header) SetRefreshButtonClickCallback(cb func()) {
	h.refreshButton.SetClickCallback(cb)
}

// SetAddButtonClickCallback sets the callback for add button clicks.
func (h *header) SetAddButtonClickCallback(cb func()) {
	h.addButton.SetClickCallback(cb)
}

// SetRemoveButtonClickCallback sets the callback for remove button clicks.
func (h *header) SetRemoveButtonClickCallback(cb func()) {
	h.removeButton.SetClickCallback(cb)
}

// Close frees the resources backing the ChartHeader.
func (h *header) Close() {
	h.barButton.Close()
	h.candlestickButton.Close()
	h.refreshButton.Close()
	h.addButton.Close()
	h.removeButton.Close()
}
