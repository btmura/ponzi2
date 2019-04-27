package chart

import (
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/centeredtext"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/app/view/status"
	"github.com/btmura/ponzi2/internal/errors"
)

const (
	thumbChartRounding = 6
	thumbChartPadding  = 2
)

var (
	thumbSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	thumbQuotePrinter            = func(q *model.Quote) string { return status.PriceChange(q) }
)

// ChartThumb shows a thumbnail for a stock.
type ChartThumb struct {
	// header renders the header with the symbol, quote, and buttons.
	header *chartHeader

	dailyStochastic         *chartStochastics
	dailyStochasticCursor   *chartStochasticCursor
	dailyStochasticTimeline *chartTimeline

	weeklyStochastic         *chartStochastics
	weeklyStochasticCursor   *chartStochasticCursor
	weeklyStochasticTimeline *chartTimeline

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

// NewChartThumb creates a ChartThumb.
func NewChartThumb(fps int) *ChartThumb {
	return &ChartThumb{
		header: newChartHeader(&chartHeaderArgs{
			SymbolQuoteTextRenderer: thumbSymbolQuoteTextRenderer,
			QuotePrinter:            thumbQuotePrinter,
			ShowRemoveButton:        true,
			Rounding:                thumbChartRounding,
			Padding:                 thumbChartPadding,
			FPS:                     fps,
		}),

		dailyStochastic:         newChartStochastics(yellow),
		dailyStochasticCursor:   new(chartStochasticCursor),
		dailyStochasticTimeline: new(chartTimeline),

		weeklyStochastic:         newChartStochastics(purple),
		weeklyStochasticCursor:   new(chartStochasticCursor),
		weeklyStochasticTimeline: new(chartTimeline),

		loadingText: centeredtext.New(thumbSymbolQuoteTextRenderer, "LOADING..."),
		errorText:   centeredtext.New(thumbSymbolQuoteTextRenderer, "ERROR", centeredtext.Color(orange)),
		loading:     true,
		fadeIn:      animation.New(1 * fps),
	}
}

// SetLoading toggles the Chart's loading indicator.
func (ch *ChartThumb) SetLoading(loading bool) {
	ch.loading = loading
	ch.header.SetLoading(loading)
}

// SetError toggles the Chart's error indicator.
func (ch *ChartThumb) SetError(error bool) {
	ch.hasError = error
	ch.header.SetError(error)
}

// SetData sets the data to be shown on the chart.
func (ch *ChartThumb) SetData(data *ChartData) error {
	if data == nil {
		return errors.Errorf("missing data")
	}

	if !ch.hasStockUpdated && data.Chart != nil {
		ch.fadeIn.Start()
	}
	ch.hasStockUpdated = data.Chart != nil

	if err := ch.header.SetData(data); err != nil {
		return err
	}

	dc := data.Chart

	if dc == nil {
		return nil
	}

	ch.dailyStochastic.SetData(dc.DailyStochasticSeries)
	ch.dailyStochasticCursor.SetData()
	if err := ch.dailyStochasticTimeline.SetData(dc.Range, dc.TradingSessionSeries); err != nil {
		return err
	}

	ch.weeklyStochastic.SetData(dc.WeeklyStochasticSeries)
	ch.weeklyStochasticCursor.SetData()
	if err := ch.weeklyStochasticTimeline.SetData(dc.Range, dc.TradingSessionSeries); err != nil {
		return err
	}

	return nil
}

// ProcessInput processes input.
func (ch *ChartThumb) ProcessInput(
	bounds image.Rectangle,
	mousePos image.Point,
	mouseLeftButtonReleased bool,
	scheduledCallbacks *[]func(),
) {
	ch.fullBounds = bounds
	ch.mousePos = mousePos

	r, clicks := ch.header.ProcessInput(bounds, mousePos, mouseLeftButtonReleased, scheduledCallbacks)
	ch.bodyBounds = r
	if !clicks.HasClicks() && mousePos.In(bounds) {
		*scheduledCallbacks = append(*scheduledCallbacks, ch.thumbClickCallback)
	}

	rects := rect.Slice(r, 0.5)
	for i := range rects {
		rects[i] = rects[i].Inset(thumbChartPadding)
	}

	dr, wr := rects[1], rects[0]

	ch.dailyStochastic.ProcessInput(dr)
	ch.dailyStochasticCursor.ProcessInput(dr, dr, ch.mousePos)
	ch.dailyStochasticTimeline.ProcessInput(dr)

	ch.weeklyStochastic.ProcessInput(wr)
	ch.weeklyStochasticCursor.ProcessInput(wr, wr, ch.mousePos)
	ch.weeklyStochasticTimeline.ProcessInput(wr)
}

// Update updates the ChartThumb.
func (ch *ChartThumb) Update() (dirty bool) {
	if ch.header.Update() {
		dirty = true
	}
	if ch.fadeIn.Update() {
		dirty = true
	}
	return dirty
}

// Render renders the ChartThumb.
func (ch *ChartThumb) Render(fudge float32) error {
	// Render the border around the chart.
	rect.StrokeRoundedRect(ch.fullBounds, thumbChartRounding)

	// Render the header and the line below it.
	ch.header.Render(fudge)

	r := ch.bodyBounds
	rect.RenderLineAtTop(r)

	// Only show messages if no prior data to show.
	if !ch.hasStockUpdated {
		if ch.loading {
			ch.loadingText.Render(fudge)
			return nil
		}

		if ch.hasError {
			ch.errorText.Render(fudge)
			return nil
		}
	}

	old := gfx.Alpha()
	gfx.SetAlpha(old * ch.fadeIn.Value(fudge))
	defer gfx.SetAlpha(old)

	rects := rect.Slice(r, 0.5)
	rect.RenderLineAtTop(rects[0])

	ch.dailyStochasticTimeline.Render(fudge)
	ch.weeklyStochasticTimeline.Render(fudge)

	ch.dailyStochastic.Render(fudge)
	ch.weeklyStochastic.Render(fudge)

	ch.dailyStochasticCursor.Render(fudge)
	ch.weeklyStochasticCursor.Render(fudge)

	return nil
}

// SetRemoveButtonClickCallback sets the callback for remove button clicks.
func (ch *ChartThumb) SetRemoveButtonClickCallback(cb func()) {
	ch.header.SetRemoveButtonClickCallback(cb)
}

// SetThumbClickCallback sets the callback for thumbnail clicks.
func (ch *ChartThumb) SetThumbClickCallback(cb func()) {
	ch.thumbClickCallback = cb
}

// Close frees the resources backing the chart thumbnail.
func (ch *ChartThumb) Close() {
	ch.header.Close()
	ch.dailyStochastic.Close()
	ch.dailyStochasticTimeline.Close()
	ch.weeklyStochastic.Close()
	ch.weeklyStochasticTimeline.Close()
	ch.thumbClickCallback = nil
}
