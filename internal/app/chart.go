package app

import (
	"fmt"
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
)

const (
	chartRounding = 10
	chartPadding  = 5
)

var (
	chartSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 24)
	chartFormatQuote             = func(st *ModelStock) string {
		if st.Price() != 0 {
			return fmt.Sprintf("%.2f %+5.2f %+5.2f%%", st.Price(), st.Change(), st.PercentChange()*100.0)
		}
		return ""
	}
	chartLoadingText = NewCenteredText(chartSymbolQuoteTextRenderer, "LOADING...", white)
	chartErrorText   = NewCenteredText(chartSymbolQuoteTextRenderer, "ERROR", orange)
)

// Constants for rendering a bubble behind an axis-label.
const (
	chartAxisLabelBubblePadding  = 4
	chartAxisLabelBubbleRounding = 6
)

// Shared variables used by multiple chart components.
var (
	chartAxisLabelTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	chartGridHorizLine         = gfx.HorizColoredLineVAO(gray, gray)
)

var (
	chartCursorHorizLine = gfx.HorizColoredLineVAO(lightGray, lightGray)
	chartCursorVertLine  = gfx.VertColoredLineVAO(lightGray, lightGray)
)

// Chart shows a stock chart for a single stock.
type Chart struct {
	// header renders the header with the symbol, quote, and buttons.
	header *ChartHeader

	// weekLines renders the vertical weekly lines.
	weekLines *ChartWeekLines

	// prices renders the candlesticks.
	prices *ChartPrices

	// avgLines renders the moving average lines.
	avgLines *ChartAvgLines

	// volume renders the volume bars.
	volume *ChartVolume

	// dailyStochastics renders the daily stochastics.
	dailyStochastics *ChartStochastics

	// weeklyStochastics renders the weekly stochastics.
	weeklyStochastics *ChartStochastics

	// loading is true when data for the stock is being retrieved.
	loading bool

	// hasStockUpdated is true if the stock has reported data.
	hasStockUpdated bool

	// hasError is true there was a loading issue.
	hasError bool
}

// NewChart creates a new Chart.
func NewChart() *Chart {
	return &Chart{
		header: NewChartHeader(&ChartHeaderArgs{
			SymbolQuoteTextRenderer: chartSymbolQuoteTextRenderer,
			QuoteFormatter:          chartFormatQuote,
			ShowRefreshButton:       true,
			ShowAddButton:           true,
			Rounding:                chartRounding,
			Padding:                 chartPadding,
		}),
		weekLines:         NewChartWeekLines(),
		prices:            NewChartPrices(),
		avgLines:          NewChartAvgLines(),
		volume:            NewChartVolume(),
		dailyStochastics:  NewChartStochastics(DailyInterval),
		weeklyStochastics: NewChartStochastics(WeeklyInterval),
		loading:           true,
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
func (ch *Chart) SetStock(st *ModelStock) {
	ch.hasStockUpdated = !st.LastUpdateTime.IsZero()
	ch.header.SetStock(st)
	ch.weekLines.SetStock(st)
	ch.prices.SetStock(st)
	ch.avgLines.SetStock(st)
	ch.volume.SetStock(st)
	ch.dailyStochastics.SetStock(st)
	ch.weeklyStochastics.SetStock(st)
}

// Update updates the Chart.
func (ch *Chart) Update() {
	ch.header.Update()
}

// Render renders the Chart.
func (ch *Chart) Render(vc ViewContext) {
	// Render the border around the chart.
	strokeRoundedRect(vc.Bounds, chartRounding)

	// Render the header and the line below it.
	r, _ := ch.header.Render(vc)
	rects := sliceRect(r, 0.13, 0.13, 0.13, 0.61)
	renderHorizDivider(rects[3], horizLine)

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

	// Render the rest of the dividers.
	for i := 0; i < 3; i++ {
		renderHorizDivider(rects[i], horizLine)
	}

	pr, vr, dr, wr := rects[3], rects[2], rects[1], rects[0]

	// Create separate rects for each section's labels.
	plr, vlr, dlr, wlr := pr, vr, dr, wr

	maxWidth := ch.prices.MaxLabelSize.X
	if ch.volume.MaxLabelSize.X > maxWidth {
		maxWidth = ch.volume.MaxLabelSize.X
	}
	if ch.dailyStochastics.MaxLabelSize.X > maxWidth {
		maxWidth = ch.dailyStochastics.MaxLabelSize.X
	}
	if ch.weeklyStochastics.MaxLabelSize.X > maxWidth {
		maxWidth = ch.weeklyStochastics.MaxLabelSize.X
	}
	maxWidth += chartPadding

	plr.Min.X = pr.Max.X - maxWidth
	vlr.Min.X = vr.Max.X - maxWidth
	dlr.Min.X = dr.Max.X - maxWidth
	wlr.Min.X = wr.Max.X - maxWidth

	// Trim off the label rects from the main rects.
	pr.Max.X = plr.Min.X
	vr.Max.X = vlr.Min.X
	dr.Max.X = dlr.Min.X
	wr.Max.X = wlr.Min.X

	// Pad all the rects.
	pr = pr.Inset(chartPadding)
	vr = vr.Inset(chartPadding)
	dr = dr.Inset(chartPadding)
	wr = wr.Inset(chartPadding)

	plr = plr.Inset(chartPadding)
	vlr = vlr.Inset(chartPadding)
	dlr = dlr.Inset(chartPadding)
	wlr = wlr.Inset(chartPadding)

	ch.weekLines.Render(pr)
	ch.weekLines.Render(vr)
	ch.weekLines.Render(dr)
	ch.weekLines.Render(wr)

	ch.prices.Render(pr)
	ch.avgLines.Render(pr)
	ch.volume.Render(vr)
	ch.dailyStochastics.Render(dr)
	ch.weeklyStochastics.Render(wr)

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
		ch.header = nil
	}
	if ch.weekLines != nil {
		ch.weekLines.Close()
		ch.weekLines = nil
	}
	if ch.prices != nil {
		ch.prices.Close()
		ch.prices = nil
	}
	if ch.volume != nil {
		ch.volume.Close()
		ch.volume = nil
	}
	if ch.dailyStochastics != nil {
		ch.dailyStochastics.Close()
		ch.dailyStochastics = nil
	}
	if ch.weeklyStochastics != nil {
		ch.weeklyStochastics.Close()
		ch.weeklyStochastics = nil
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
