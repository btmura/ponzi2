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
)

// ChartThumb shows a thumbnail for a stock.
type ChartThumb struct {
	// header renders the header with the symbol, quote, and buttons.
	header *chartHeader

	dailyStochastics             *chartStochastics
	dailyStochasticsCursorLabels *chartStochasticsCursorLabels
	dailyStochasticsTimeLines    *chartTimeLines

	weeklyStochastics             *chartStochastics
	weeklyStochasticsCursorLabels *chartStochasticsCursorLabels
	weeklyStochasticsTimeLines    *chartTimeLines

	// loadingText is the text shown when loading from a fresh state.
	loadingText *centeredText

	// errorText is the text shown when an error occurs from a fresh state.
	errorText *centeredText

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

	// fullBounds is the rect with global coords that should be drawn within.
	fullBounds image.Rectangle

	// bodyBounds is a sub-rect of fullBounds without the header.
	bodyBounds image.Rectangle

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

		dailyStochastics:             newChartStochastics(yellow),
		dailyStochasticsCursorLabels: newChartStochasticsCursorLabels(),
		dailyStochasticsTimeLines:    newChartTimeLines(),

		weeklyStochastics:             newChartStochastics(purple),
		weeklyStochasticsCursorLabels: newChartStochasticsCursorLabels(),
		weeklyStochasticsTimeLines:    newChartTimeLines(),

		loadingText: newCenteredText(thumbSymbolQuoteTextRenderer, "LOADING..."),
		errorText:   newCenteredText(thumbSymbolQuoteTextRenderer, "ERROR", centeredTextColor(orange)),
		loading:     true,
		fadeIn:      newAnimation(1 * fps),
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

	if err := ch.dailyStochasticsTimeLines.SetData(dc.Range, dc.TradingSessionSeries); err != nil {
		return err
	}

	ch.dailyStochastics.SetData(dc.DailyStochasticSeries)

	if err := ch.weeklyStochasticsTimeLines.SetData(dc.Range, dc.TradingSessionSeries); err != nil {
		return err
	}

	ch.weeklyStochastics.SetData(dc.WeeklyStochasticSeries)

	return nil
}

// ProcessInput processes input.
func (ch *ChartThumb) ProcessInput(ic inputContext) {
	ch.fullBounds = ic.Bounds
	ch.mousePos = ic.MousePos

	r, clicks := ch.header.ProcessInput(ic)
	ch.bodyBounds = r
	if !clicks.HasClicks() && ic.LeftClickInBounds() {
		*ic.ScheduledCallbacks = append(*ic.ScheduledCallbacks, ch.thumbClickCallback)
	}

	rects := sliceRect(r, 0.5)
	for i := range rects {
		rects[i] = rects[i].Inset(thumbChartPadding)
	}

	dr, wr := rects[1], rects[0]

	ic.Bounds = dr
	ch.dailyStochastics.ProcessInput(ic)
	ch.dailyStochasticsCursorLabels.ProcessInput(ic, dr)
	ch.dailyStochasticsTimeLines.ProcessInput(ic)

	ic.Bounds = wr
	ch.weeklyStochastics.ProcessInput(ic)
	ch.weeklyStochasticsCursorLabels.ProcessInput(ic, wr)
	ch.weeklyStochasticsTimeLines.ProcessInput(ic)
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
	strokeRoundedRect(ch.fullBounds, thumbChartRounding)

	// Render the header and the line below it.
	ch.header.Render(fudge)

	r := ch.bodyBounds
	renderRectTopDivider(r, horizLine)

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

	rects := sliceRect(r, 0.5)
	renderRectTopDivider(rects[0], horizLine)

	for i := range rects {
		rects[i] = rects[i].Inset(thumbChartPadding)
	}

	dr, wr := rects[1], rects[0]

	ch.dailyStochasticsTimeLines.Render(fudge)
	ch.weeklyStochasticsTimeLines.Render(fudge)

	ch.dailyStochastics.Render(fudge)
	ch.weeklyStochastics.Render(fudge)

	renderCursorLines(dr, ch.mousePos)
	renderCursorLines(wr, ch.mousePos)

	ch.dailyStochastics.Render(fudge)
	ch.weeklyStochastics.Render(fudge)

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

	if ch.dailyStochasticsTimeLines != nil {
		ch.dailyStochasticsTimeLines.Close()
	}
	if ch.dailyStochastics != nil {
		ch.dailyStochastics.Close()
	}

	if ch.weeklyStochasticsTimeLines != nil {
		ch.weeklyStochasticsTimeLines.Close()
	}
	if ch.weeklyStochastics != nil {
		ch.weeklyStochastics.Close()
	}

	ch.thumbClickCallback = nil
}
