package view

import (
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/status"
)

const (
	thumbChartRounding = 6
	thumbChartPadding  = 2
)

var (
	thumbSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	thumbQuotePrinter            = func(q *model.Quote) string { return priceStatus(q) }
	thumbLoadingText             = newCenteredText(thumbSymbolQuoteTextRenderer, "LOADING...")
	thumbErrorText               = newCenteredText(thumbSymbolQuoteTextRenderer, "ERROR", centeredTextColor(orange))
)

// ChartThumb shows a thumbnail for a stock.
type ChartThumb struct {
	// header renders the header with the symbol, quote, and buttons.
	header *chartHeader

	// timeLines renders the vertical time lines.
	timeLines *chartTimeLines

	// dailyStochastics renders the daily stochastics.
	dailyStochastics *chartStochastics

	// weeklyStochastics renders the weekly stochastics.
	weeklyStochastics *chartStochastics

	// thumbClickCallback is the callback to schedule when the thumb is clicked.
	thumbClickCallback func()

	// loading is true when data for the stock is being retrieved.
	loading bool

	// hasStockUpdated is true if the stock has reported data.
	hasStockUpdated bool

	// hasError is true there was a loading issue.
	hasError bool

	// fadeIn fades in the data after it loads.
	fadeIn *animation

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle

	// mousePos is the current mouse position.
	mousePos image.Point
}

// NewChartThumb creates a ChartThumb.
func NewChartThumb() *ChartThumb {
	return &ChartThumb{
		header: newChartHeader(&chartHeaderArgs{
			SymbolQuoteTextRenderer: thumbSymbolQuoteTextRenderer,
			QuotePrinter:            thumbQuotePrinter,
			ShowRemoveButton:        true,
			Rounding:                thumbChartRounding,
			Padding:                 thumbChartPadding,
		}),
		timeLines:         newChartTimeLines(),
		dailyStochastics:  newChartStochastics(yellow),
		weeklyStochastics: newChartStochastics(purple),
		loading:           true,
		fadeIn:            newAnimation(1 * fps),
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
		return status.Error("missing data")
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

	if err := ch.timeLines.SetData(dc.Range, dc.TradingSessionSeries); err != nil {
		return err
	}

	ch.dailyStochastics.SetData(dc.DailyStochasticSeries)
	ch.weeklyStochastics.SetData(dc.WeeklyStochasticSeries)

	return nil
}

// ProcessInput processes input.
func (ch *ChartThumb) ProcessInput(ic inputContext) {
	ch.bounds = ic.Bounds
	ch.mousePos = ic.MousePos
	clicks := ch.header.ProcessInput(ic)
	if !clicks.HasClicks() && ic.LeftClickInBounds() {
		*ic.ScheduledCallbacks = append(*ic.ScheduledCallbacks, ch.thumbClickCallback)
	}
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
func (ch *ChartThumb) Render(rc renderContext) error {
	// Render the border around the chart.
	strokeRoundedRect(ch.bounds, thumbChartRounding)

	// Render the header and the line below it.
	r := ch.header.Render(rc)
	renderRectTopDivider(r, horizLine)

	// Only show messages if no prior data to show.
	if !ch.hasStockUpdated {
		if ch.loading {
			thumbLoadingText.Render(r)
			return nil
		}

		if ch.hasError {
			thumbErrorText.Render(r)
			return nil
		}
	}

	old := gfx.Alpha()
	gfx.SetAlpha(old * ch.fadeIn.Value(rc.Fudge))
	defer gfx.SetAlpha(old)

	rects := sliceRect(r, 0.5)
	renderRectTopDivider(rects[0], horizLine)

	for i := range rects {
		rects[i] = rects[i].Inset(thumbChartPadding)
	}

	dr, wr := rects[1], rects[0]

	ch.timeLines.Render(dr)
	ch.timeLines.Render(wr)

	ch.dailyStochastics.Render(dr)
	ch.weeklyStochastics.Render(wr)

	renderCursorLines(dr, ch.mousePos)
	renderCursorLines(wr, ch.mousePos)

	ch.dailyStochastics.RenderCursorLabels(dr, dr, ch.mousePos)
	ch.weeklyStochastics.RenderCursorLabels(wr, wr, ch.mousePos)

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
	if ch.header != nil {
		ch.header.Close()
	}
	if ch.timeLines != nil {
		ch.timeLines.Close()
	}
	if ch.dailyStochastics != nil {
		ch.dailyStochastics.Close()
	}
	if ch.weeklyStochastics != nil {
		ch.weeklyStochastics.Close()
	}
	ch.thumbClickCallback = nil
}
