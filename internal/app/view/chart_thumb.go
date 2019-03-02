package view

import (
	"golang.org/x/image/font/gofont/goregular"

	"gitlab.com/btmura/ponzi2/internal/app/gfx"
	"gitlab.com/btmura/ponzi2/internal/app/model"
	"gitlab.com/btmura/ponzi2/internal/util"
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
		return util.Error("missing data")
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
func (ch *ChartThumb) Render(vc viewContext) error {
	// Render the border around the chart.
	strokeRoundedRect(vc.Bounds, thumbChartRounding)

	// Render the header and the line below it.
	r, clicks := ch.header.Render(vc)
	renderRectTopDivider(r, horizLine)

	if !clicks.HasClicks() && vc.LeftClickInBounds() {
		*vc.ScheduledCallbacks = append(*vc.ScheduledCallbacks, ch.thumbClickCallback)
	}

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
	gfx.SetAlpha(old * ch.fadeIn.Value(vc.Fudge))
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

	renderCursorLines(dr, vc.MousePos)
	renderCursorLines(wr, vc.MousePos)

	ch.dailyStochastics.RenderCursorLabels(dr, dr, vc.MousePos)
	ch.weeklyStochastics.RenderCursorLabels(wr, wr, vc.MousePos)

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
