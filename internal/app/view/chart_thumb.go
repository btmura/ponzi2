package view

import (
	"fmt"
	"time"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/gfx"
)

const (
	thumbChartRounding = 6
	thumbChartPadding  = 2
)

var (
	thumbSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	thumbFormatQuote             = func(st *model.Stock) string {
		if st.Price() != 0 {
			return fmt.Sprintf(" %.2f %+5.2f%% ", st.Price(), st.PercentChange())
		}
		return ""
	}
	thumbLoadingText = NewCenteredText(thumbSymbolQuoteTextRenderer, "LOADING...")
	thumbErrorText   = NewCenteredText(thumbSymbolQuoteTextRenderer, "ERROR", CenteredTextColor(orange))
)

// ChartThumb shows a thumbnail for a stock.
type ChartThumb struct {
	// header renders the header with the symbol, quote, and buttons.
	header *ChartHeader

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

	// fadeIn is fade-in animation.
	fadeIn *animation
}

// NewChartThumb creates a ChartThumb.
func NewChartThumb() *ChartThumb {
	return &ChartThumb{
		header: NewChartHeader(&ChartHeaderArgs{
			SymbolQuoteTextRenderer: thumbSymbolQuoteTextRenderer,
			QuoteFormatter:          thumbFormatQuote,
			ShowRemoveButton:        true,
			Rounding:                thumbChartRounding,
			Padding:                 thumbChartPadding,
		}),
		timeLines:         newChartTimeLines(),
		dailyStochastics:  newChartStochastics(yellow),
		weeklyStochastics: newChartStochastics(purple),
		loading:           true,
		fadeIn:            newAnimation(1 * time.Second),
	}
}

// SetLoading sets the ChartThumb's loading state.
func (ch *ChartThumb) SetLoading(loading bool) {
	ch.loading = loading
	ch.header.SetLoading(loading)
}

// SetError sets the ChartThumb's error flag.
func (ch *ChartThumb) SetError(error bool) {
	ch.hasError = error
	ch.header.SetError(error)
}

// SetStock sets the ChartThumb's stock.
func (ch *ChartThumb) SetStock(st *model.Stock) {
	ch.hasStockUpdated = !st.LastUpdateTime.IsZero()
	ch.header.SetStock(st)
	ch.timeLines.SetData(st.DailyTradingSessionSeries)
	ch.dailyStochastics.SetData(st.DailyStochasticSeries)
	ch.weeklyStochastics.SetData(st.WeeklyStochasticSeries)
	ch.fadeIn.Start()
}

// Update updates the ChartThumb.
func (ch *ChartThumb) Update() (animating bool) {
	if ch.header.Update() {
		animating = true
	}
	if ch.fadeIn.Update() {
		animating = true
	}
	return animating
}

// Render renders the ChartThumb.
func (ch *ChartThumb) Render(vc viewContext) {
	gfx.SetAlpha(1)

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
			return
		}

		if ch.hasError {
			thumbErrorText.Render(r)
			return
		}
	}

	gfx.SetAlpha(ch.fadeIn.Value(vc.Fudge))

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
