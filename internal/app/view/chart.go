package view

import (
	"fmt"
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/gfx"
)

const (
	chartRounding                   = 10
	chartPadding                    = 5
	chartAxisVerticalPaddingPercent = .05
)

var (
	chartSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 24)
	chartFormatQuote             = func(st *model.Stock) string {
		if st.Price() != 0 {
			return fmt.Sprintf("%.2f %+5.2f %+5.2f%%", st.Price(), st.Change(), st.PercentChange())
		}
		return ""
	}
	chartLoadingText = NewCenteredText(chartSymbolQuoteTextRenderer, "LOADING...")
	chartErrorText   = NewCenteredText(chartSymbolQuoteTextRenderer, "ERROR", CenteredTextColor(orange))
)

// Constants for rendering a bubble behind an axis-label.
const (
	chartAxisLabelBubblePadding  = 4
	chartAxisLabelBubbleRounding = 6
)

var chartAxisLabelBubbleSpec = bubbleSpec{
	rounding: 6,
	padding:  4,
}

// Shared variables used by multiple chart components.
var (
	chartAxisLabelTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	chartGridHorizLine         = horizLineVAO(gray)
)

var (
	chartCursorHorizLine = horizLineVAO(lightGray)
	chartCursorVertLine  = vertLineVAO(lightGray)
)

// Chart shows a stock chart for a single stock.
type Chart struct {
	// header renders the header with the symbol, quote, and buttons.
	header *chartHeader

	// timeLines renders the vertical time lines.
	timeLines *chartTimeLines

	// prices renders the candlesticks.
	prices *ChartPrices

	// movingAverage renders the 25 day moving average.
	movingAverage25 *chartMovingAverage

	// movingAverage50 renders the 50 day moving average.
	movingAverage50 *chartMovingAverage

	// movingAverage200 renders the 200 day moving average.
	movingAverage200 *chartMovingAverage

	// volume renders the volume bars.
	volume *ChartVolume

	// dailyStochastics renders the daily stochastics.
	dailyStochastics *chartStochastics

	// weeklyStochastics renders the weekly stochastics.
	weeklyStochastics *chartStochastics

	// timeLabels renders the time labels.
	timeLabels *ChartTimeLabels

	// loading is true when data for the stock is being retrieved.
	loading bool

	// hasStockUpdated is true if the stock has reported data.
	hasStockUpdated bool

	// hasError is true there was a loading issue.
	hasError bool

	// fadeIn is fade-in animation.
	fadeIn *animation
}

// NewChart creates a new Chart.
func NewChart() *Chart {
	return &Chart{
		header: newChartHeader(&chartHeaderArgs{
			SymbolQuoteTextRenderer: chartSymbolQuoteTextRenderer,
			QuoteFormatter:          chartFormatQuote,
			ShowRefreshButton:       true,
			ShowAddButton:           true,
			Rounding:                chartRounding,
			Padding:                 chartPadding,
		}),
		timeLines:         newChartTimeLines(),
		prices:            NewChartPrices(),
		movingAverage25:   newChartMovingAverage(purple),
		movingAverage50:   newChartMovingAverage(yellow),
		movingAverage200:  newChartMovingAverage(white),
		volume:            NewChartVolume(),
		dailyStochastics:  newChartStochastics(yellow),
		weeklyStochastics: newChartStochastics(purple),
		timeLabels:        NewChartTimeLabels(),
		loading:           true,
		fadeIn:            newAnimation(1*fps, false),
	}
}

// SetLoading sets the Chart's loading state.
func (ch *Chart) SetLoading(loading bool) {
	ch.loading = loading
	ch.header.SetLoading(loading)
}

// SetError sets the Chart's error flag.
func (ch *Chart) SetError(error bool) {
	ch.hasError = error
	ch.header.SetError(error)
}

// SetStock sets the Chart's stock.
func (ch *Chart) SetStock(st *model.Stock) {
	if !ch.hasStockUpdated && !st.LastUpdateTime.IsZero() {
		ch.fadeIn.Start()
	}
	ch.hasStockUpdated = !st.LastUpdateTime.IsZero()

	ch.header.SetStock(st)
	ch.timeLines.SetData(st.DailyTradingSessionSeries)
	ch.prices.SetStock(st)
	ch.movingAverage25.SetData(st.DailyTradingSessionSeries, st.DailyMovingAverageSeries25)
	ch.movingAverage50.SetData(st.DailyTradingSessionSeries, st.DailyMovingAverageSeries50)
	ch.movingAverage200.SetData(st.DailyTradingSessionSeries, st.DailyMovingAverageSeries200)
	ch.volume.SetStock(st)
	ch.dailyStochastics.SetData(st.DailyStochasticSeries)
	ch.weeklyStochastics.SetData(st.WeeklyStochasticSeries)
	ch.timeLabels.SetStock(st)
}

// Update updates the Chart.
func (ch *Chart) Update() (animating bool) {
	if ch.header.Update() {
		animating = true
	}
	if ch.fadeIn.Update() {
		animating = true
	}
	return animating
}

// Render renders the Chart.
func (ch *Chart) Render(vc viewContext) {
	gfx.SetAlpha(1)

	// Render the border around the chart.
	strokeRoundedRect(vc.Bounds, chartRounding)

	// Render the header and the line below it.
	r, _ := ch.header.Render(vc)
	renderRectTopDivider(r, horizLine)

	// Only show messages if no prior data to show.
	if !ch.hasStockUpdated {
		if ch.loading {
			chartLoadingText.Render(r)
			return
		}

		if ch.hasError {
			chartErrorText.Render(r)
			return
		}
	}

	gfx.SetAlpha(ch.fadeIn.Value(vc.Fudge))

	// Calculate percentage needed for the time labels.
	tperc := float32(ch.timeLabels.MaxLabelSize.Y+chartPadding*2) / float32(r.Dy())

	// Render the dividers between the sections.
	rects := sliceRect(r, tperc, 0.13, 0.13, 0.13)
	for i := 0; i < len(rects)-1; i++ {
		renderRectTopDivider(rects[i], horizLine)
	}

	pr, vr, dr, wr, tr := rects[4], rects[3], rects[2], rects[1], rects[0]

	// Create separate rects for each section's labels shown on the right.
	plr, vlr, dlr, wlr := pr, vr, dr, wr

	// Figure out width to trim off on the right of each rect for the labels.
	maxWidth := ch.prices.MaxLabelSize.X
	if w := ch.volume.MaxLabelSize.X; w > maxWidth {
		maxWidth = w
	}
	if w := ch.dailyStochastics.MaxLabelSize.X; w > maxWidth {
		maxWidth = w
	}
	if w := ch.weeklyStochastics.MaxLabelSize.X; w > maxWidth {
		maxWidth = w
	}
	maxWidth += chartPadding

	// Set left side of label rects.
	plr.Min.X = pr.Max.X - maxWidth
	vlr.Min.X = vr.Max.X - maxWidth
	dlr.Min.X = dr.Max.X - maxWidth
	wlr.Min.X = wr.Max.X - maxWidth

	// Trim off the label rects from the main rects.
	pr.Max.X = plr.Min.X
	vr.Max.X = vlr.Min.X
	dr.Max.X = dlr.Min.X
	wr.Max.X = wlr.Min.X

	// Time labels and its cursors labels overlap and use the same rect.
	tr.Max.X = wlr.Min.X
	tlr := tr

	// Pad all the rects.
	pr = pr.Inset(chartPadding)
	vr = vr.Inset(chartPadding)
	dr = dr.Inset(chartPadding)
	wr = wr.Inset(chartPadding)
	tr = tr.Inset(chartPadding)

	plr = plr.Inset(chartPadding)
	vlr = vlr.Inset(chartPadding)
	dlr = dlr.Inset(chartPadding)
	wlr = wlr.Inset(chartPadding)
	tlr = tlr.Inset(chartPadding)

	ch.timeLines.Render(pr)
	ch.timeLines.Render(vr)
	ch.timeLines.Render(dr)
	ch.timeLines.Render(wr)

	ch.prices.Render(pr)
	ch.movingAverage25.Render(pr)
	ch.movingAverage50.Render(pr)
	ch.movingAverage200.Render(pr)
	ch.volume.Render(vr)
	ch.dailyStochastics.Render(dr)
	ch.weeklyStochastics.Render(wr)
	ch.timeLabels.Render(tr)

	ch.prices.RenderAxisLabels(plr)
	ch.volume.RenderAxisLabels(vlr)
	ch.dailyStochastics.RenderAxisLabels(dlr)
	ch.weeklyStochastics.RenderAxisLabels(wlr)

	renderCursorLines(pr, vc.MousePos)
	renderCursorLines(vr, vc.MousePos)
	renderCursorLines(dr, vc.MousePos)
	renderCursorLines(wr, vc.MousePos)

	ch.prices.RenderCursorLabels(pr, plr, vc.MousePos)
	ch.volume.RenderCursorLabels(vr, vlr, vc.MousePos)
	ch.dailyStochastics.RenderCursorLabels(dr, dlr, vc.MousePos)
	ch.weeklyStochastics.RenderCursorLabels(wr, wlr, vc.MousePos)
	ch.timeLabels.RenderCursorLabels(tr, tlr, vc.MousePos)
}

// SetRefreshButtonClickCallback sets the callback for refresh button clicks.
func (ch *Chart) SetRefreshButtonClickCallback(cb func()) {
	ch.header.SetRefreshButtonClickCallback(cb)
}

// SetAddButtonClickCallback sets the callback for add button clicks.
func (ch *Chart) SetAddButtonClickCallback(cb func()) {
	ch.header.SetAddButtonClickCallback(cb)
}

// Close frees the resources backing the chart.
func (ch *Chart) Close() {
	if ch.header != nil {
		ch.header.Close()
	}
	if ch.timeLines != nil {
		ch.timeLines.Close()
	}
	if ch.prices != nil {
		ch.prices.Close()
	}
	if ch.movingAverage25 != nil {
		ch.movingAverage25.Close()
	}
	if ch.movingAverage50 != nil {
		ch.movingAverage50.Close()
	}
	if ch.movingAverage200 != nil {
		ch.movingAverage200.Close()
	}
	if ch.volume != nil {
		ch.volume.Close()
	}
	if ch.dailyStochastics != nil {
		ch.dailyStochastics.Close()
	}
	if ch.weeklyStochastics != nil {
		ch.weeklyStochastics.Close()
	}
	if ch.timeLabels != nil {
		ch.timeLabels.Close()
	}
}

func renderCursorLines(r image.Rectangle, mousePos image.Point) {
	if mousePos.In(r) {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, mousePos.Y, r.Max.X, mousePos.Y))
		chartCursorHorizLine.Render()
	}

	if mousePos.X >= r.Min.X && mousePos.X <= r.Max.X {
		gfx.SetModelMatrixRect(image.Rect(mousePos.X, r.Min.Y, mousePos.X, r.Max.Y))
		chartCursorVertLine.Render()
	}
}
