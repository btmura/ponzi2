package chart

import (
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/centeredtext"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/app/view/status"
	"github.com/btmura/ponzi2/internal/errors"
)

const (
	thumbRounding = 6
	thumbPadding  = 2
)

var (
	thumbSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	thumbQuotePrinter            = func(q *model.Quote) string { return status.PriceChange(q) }
)

// Thumb shows a thumbnail for a stock.
type Thumb struct {
	// header renders the header with the symbol, quote, and buttons.
	header *header

	dailyStochastic         *stochastic
	dailyStochasticCursor   *stochasticCursor
	dailyStochasticTimeline *timeline

	weeklyStochastic         *stochastic
	weeklyStochasticCursor   *stochasticCursor
	weeklyStochasticTimeline *timeline

	// loadingText is the text shown when loading from a fresh state.
	loadingText *centeredtext.CenteredText

	// errorText is the text shown when an error occurs from a fresh state.
	errorText *centeredtext.CenteredText

	// thumbClickCallback is the callback to schedule when the thumb is clicked.
	thumbClickCallback func()

	// loading is true when data for the stock is being retrieved.
	loading bool

	// hasStockUpdated is true if the stock has reported data.
	hasStockUpdated bool

	// hasError is true there was a loading issue.
	hasError bool

	// fadeIn fades in the data after it loads.
	fadeIn *animation.Animation

	// fullBounds is the rect with global coords that should be drawn within.
	fullBounds image.Rectangle

	// bodyBounds is a sub-rect of fullBounds without the header.
	bodyBounds image.Rectangle

	// mousePos is the current mouse position.
	mousePos image.Point
}

// NewThumb creates a Thumb.
func NewThumb(fps int) *Thumb {
	return &Thumb{
		header: newHeader(&headerArgs{
			SymbolQuoteTextRenderer: thumbSymbolQuoteTextRenderer,
			QuotePrinter:            thumbQuotePrinter,
			ShowRemoveButton:        true,
			Rounding:                thumbRounding,
			Padding:                 thumbPadding,
			FPS:                     fps,
		}),

		dailyStochastic:         newStochastic(color.Yellow),
		dailyStochasticCursor:   new(stochasticCursor),
		dailyStochasticTimeline: new(timeline),

		weeklyStochastic:         newStochastic(color.Purple),
		weeklyStochasticCursor:   new(stochasticCursor),
		weeklyStochasticTimeline: new(timeline),

		loadingText: centeredtext.New(thumbSymbolQuoteTextRenderer, "LOADING..."),
		errorText:   centeredtext.New(thumbSymbolQuoteTextRenderer, "ERROR", centeredtext.Color(color.Orange)),
		loading:     true,
		fadeIn:      animation.New(1 * fps),
	}
}

// SetLoading toggles the Chart's loading indicator.
func (t *Thumb) SetLoading(loading bool) {
	t.loading = loading
	t.header.SetLoading(loading)
}

// SetError toggles the Chart's error indicator.
func (t *Thumb) SetError(error bool) {
	t.hasError = error
	t.header.SetError(error)
}

// SetData sets the data to be shown on the chart.
func (t *Thumb) SetData(data *Data) error {
	if data == nil {
		return errors.Errorf("missing data")
	}

	if !t.hasStockUpdated && data.Chart != nil {
		t.fadeIn.Start()
	}
	t.hasStockUpdated = data.Chart != nil

	if err := t.header.SetData(data); err != nil {
		return err
	}

	dc := data.Chart

	if dc == nil {
		return nil
	}

	t.dailyStochastic.SetData(dc.DailyStochasticSeries)
	t.dailyStochasticCursor.SetData()
	if err := t.dailyStochasticTimeline.SetData(dc.Range, dc.TradingSessionSeries); err != nil {
		return err
	}

	t.weeklyStochastic.SetData(dc.WeeklyStochasticSeries)
	t.weeklyStochasticCursor.SetData()
	if err := t.weeklyStochasticTimeline.SetData(dc.Range, dc.TradingSessionSeries); err != nil {
		return err
	}

	return nil
}

// ProcessInput processes input.
func (t *Thumb) ProcessInput(
	bounds image.Rectangle,
	mousePos image.Point,
	mouseLeftButtonReleased bool,
	scheduledCallbacks *[]func(),
) {
	t.fullBounds = bounds
	t.mousePos = mousePos

	r, clicks := t.header.ProcessInput(bounds, mousePos, mouseLeftButtonReleased, scheduledCallbacks)
	t.bodyBounds = r
	if !clicks.HasClicks() && mouseLeftButtonReleased && mousePos.In(bounds) {
		*scheduledCallbacks = append(*scheduledCallbacks, t.thumbClickCallback)
	}

	rects := rect.Slice(r, 0.5)
	for i := range rects {
		rects[i] = rects[i].Inset(thumbPadding)
	}

	dr, wr := rects[1], rects[0]

	t.dailyStochastic.ProcessInput(dr)
	t.dailyStochasticCursor.ProcessInput(dr, dr, t.mousePos)
	t.dailyStochasticTimeline.ProcessInput(dr)

	t.weeklyStochastic.ProcessInput(wr)
	t.weeklyStochasticCursor.ProcessInput(wr, wr, t.mousePos)
	t.weeklyStochasticTimeline.ProcessInput(wr)
}

// Update updates the Thumb.
func (t *Thumb) Update() (dirty bool) {
	if t.header.Update() {
		dirty = true
	}
	if t.fadeIn.Update() {
		dirty = true
	}
	return dirty
}

// Render renders the Thumb.
func (t *Thumb) Render(fudge float32) error {
	// Render the border around the chart.
	rect.StrokeRoundedRect(t.fullBounds, thumbRounding)

	// Render the header and the line below it.
	t.header.Render(fudge)

	r := t.bodyBounds
	rect.RenderLineAtTop(r)

	// Only show messages if no prior data to show.
	if !t.hasStockUpdated {
		if t.loading {
			t.loadingText.Render(fudge)
			return nil
		}

		if t.hasError {
			t.errorText.Render(fudge)
			return nil
		}
	}

	old := gfx.Alpha()
	gfx.SetAlpha(old * t.fadeIn.Value(fudge))
	defer gfx.SetAlpha(old)

	rects := rect.Slice(r, 0.5)
	rect.RenderLineAtTop(rects[0])

	t.dailyStochasticTimeline.Render(fudge)
	t.weeklyStochasticTimeline.Render(fudge)

	t.dailyStochastic.Render(fudge)
	t.weeklyStochastic.Render(fudge)

	t.dailyStochasticCursor.Render(fudge)
	t.weeklyStochasticCursor.Render(fudge)

	return nil
}

// SetRemoveButtonClickCallback sets the callback for remove button clicks.
func (t *Thumb) SetRemoveButtonClickCallback(cb func()) {
	t.header.SetRemoveButtonClickCallback(cb)
}

// SetThumbClickCallback sets the callback for thumbnail clicks.
func (t *Thumb) SetThumbClickCallback(cb func()) {
	t.thumbClickCallback = cb
}

// Close frees the resources backing the chart thumbnail.
func (t *Thumb) Close() {
	t.header.Close()
	t.dailyStochastic.Close()
	t.dailyStochasticTimeline.Close()
	t.weeklyStochastic.Close()
	t.weeklyStochasticTimeline.Close()
	t.thumbClickCallback = nil
}
