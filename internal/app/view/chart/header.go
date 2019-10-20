package chart

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/button"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/app/view/vao"
	"github.com/btmura/ponzi2/internal/errors"
)

var (
	addButtonVAO     = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/addbutton.png")))
	refreshButtonVAO = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/refreshbutton.png")))
	removeButtonVAO  = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/removebutton.png")))
	errorIconVAO     = vao.TexturedSquare(bytes.NewReader(_escFSMustByte(false, "/data/erroricon.png")))
)

// header shows a header for charts and thumbnails with a clickable button.
type header struct {
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
func (h *header) SetError(err error ) {
	h.hasError = err != nil
}

// SetData sets the data to be shown on the chart.
func (h *header) SetData(data *Data) error {
	if data == nil {
		return errors.Errorf("missing data")
	}

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
		h.quoteColor = color.Green

	case c < 0:
		h.quoteColor = color.Red

	default:
		h.quoteColor = color.White
	}

	return nil
}

// headerClicks reports what buttons were clicked.
type headerClicks struct {
	// AddButtonClicked is true if the add button was clicked.
	AddButtonClicked bool

	// RefreshButtonClicked is true if the refresh button was clicked.
	RefreshButtonClicked bool

	// RemoveButtonClicked is true if the remove button was clicked.
	RemoveButtonClicked bool
}

// HasClicks returns true if a clickable part of the header was clicked.
func (c headerClicks) HasClicks() bool {
	return c.AddButtonClicked || c.RefreshButtonClicked || c.RemoveButtonClicked
}

// ProcessInput processes input.
func (h *header) ProcessInput(
	bounds image.Rectangle,
	mousePos image.Point,
	mouseLeftButtonReleased bool,
	scheduledCallbacks *[]func(),
) (
	body image.Rectangle,
	clicks headerClicks,
) {
	h.bounds = bounds

	height := h.padding + h.symbolQuoteTextRenderer.LineHeight() + h.padding
	buttonSize := image.Pt(height, height)

	// Render buttons in the upper right corner from right to left.
	r := bounds
	bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}

	if h.removeButton.enabled {
		clicks.RemoveButtonClicked = h.removeButton.ProcessInput(bounds, mousePos, mouseLeftButtonReleased, scheduledCallbacks)
		bounds = rect.Translate(bounds, -buttonSize.X, 0)
	}

	if h.addButton.enabled {
		clicks.AddButtonClicked = h.addButton.ProcessInput(bounds, mousePos, mouseLeftButtonReleased, scheduledCallbacks)
		bounds = rect.Translate(bounds, -buttonSize.X, 0)
	}

	if h.refreshButton.enabled || h.refreshButton.Spinning() {
		clicks.RefreshButtonClicked = h.refreshButton.ProcessInput(bounds, mousePos, mouseLeftButtonReleased, scheduledCallbacks)
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
	h.bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}

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
		pt.X += h.symbolQuoteTextRenderer.Render(h.symbol, pt, color.White)
		pt.X += h.padding

		if w := buttonEdge - pt.X; w > 0 {
			old := gfx.Alpha()
			gfx.SetAlpha(old * h.fadeIn.Value(fudge))
			pt.X += h.symbolQuoteTextRenderer.Render(h.quoteText, pt, h.quoteColor, gfx.TextRenderMaxWidth(w))
			gfx.SetAlpha(old)
		}
	}
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
	h.refreshButton.Close()
	h.addButton.Close()
	h.removeButton.Close()
}
