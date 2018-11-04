package view

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
)

const (
	chartRounding                   = 10
	chartPadding                    = 5
	chartAxisVerticalPaddingPercent = .05
)

var (
	chartSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 24)
	chartFormatQuote             = func(st *model.Stock) string {
		if q := st.Quote; q != nil {
			layout := "1/2/2006"
			if q.LatestTime.Hour() != 0 || q.LatestTime.Minute() != 0 || q.LatestTime.Second() != 0 || q.LatestTime.Nanosecond() != 0 {
				layout += " 3:04 PM"
			}
			return fmt.Sprintf("%.2f %+5.2f %+5.2f%% %v %s",
				q.LatestPrice,
				q.Change,
				q.ChangePercent*100,
				q.LatestSource,
				q.LatestTime.Format(layout))
		}
		return ""
	}
	chartLoadingText = newCenteredText(chartSymbolQuoteTextRenderer, "LOADING...")
	chartErrorText   = newCenteredText(chartSymbolQuoteTextRenderer, "ERROR", centeredTextColor(orange))
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
	prices *chartPrices

	// movingAverage renders the 25 day moving average.
	movingAverage25 *chartMovingAverage

	// movingAverage50 renders the 50 day moving average.
	movingAverage50 *chartMovingAverage

	// movingAverage200 renders the 200 day moving average.
	movingAverage200 *chartMovingAverage

	// volume renders the volume bars.
	volume *chartVolume

	// dailyStochastics renders the daily stochastics.
	dailyStochastics *chartStochastics

	// weeklyStochastics renders the weekly stochastics.
	weeklyStochastics *chartStochastics

	// timeLabels renders the time labels.
	timeLabels *chartTimeLabels

	// loading is true when data for the stock is being retrieved.
	loading bool

	// hasStockUpdated is true if the stock has reported data.
	hasStockUpdated bool

	// hasError is true there was a loading issue.
	hasError bool

	// fadeIn fades in the data after it loads.
	fadeIn *animation

	// pointer to TradingSessionSeries for use in trackline legend
	tlTradingSessions *model.TradingSessionSeries
	// pointers to moving averages for use in trackline legend
	tlMovingAverage25  *model.MovingAverageSeries
	tlMovingAverage50  *model.MovingAverageSeries
	tlMovingAverage200 *model.MovingAverageSeries
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
		prices:            newChartPrices(),
		movingAverage25:   newChartMovingAverage(purple),
		movingAverage50:   newChartMovingAverage(yellow),
		movingAverage200:  newChartMovingAverage(white),
		volume:            newChartVolume(),
		dailyStochastics:  newChartStochastics(yellow),
		weeklyStochastics: newChartStochastics(purple),
		timeLabels:        newChartTimeLabels(),
		loading:           true,
		fadeIn:            newAnimation(1 * fps),
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

// SetData sets the Chart's stock.
func (ch *Chart) SetData(st *model.Stock) {
	if !ch.hasStockUpdated && !st.LastUpdateTime.IsZero() {
		ch.fadeIn.Start()
	}
	ch.hasStockUpdated = !st.LastUpdateTime.IsZero()

	ts := st.DailyTradingSessionSeries

	ch.header.SetData(st)
	ch.timeLines.SetData(ts)
	ch.prices.SetData(ts)
	ch.movingAverage25.SetData(ts, st.DailyMovingAverageSeries25)
	ch.movingAverage50.SetData(ts, st.DailyMovingAverageSeries50)
	ch.movingAverage200.SetData(ts, st.DailyMovingAverageSeries200)
	ch.volume.SetData(ts)
	ch.dailyStochastics.SetData(st.DailyStochasticSeries)
	ch.weeklyStochastics.SetData(st.WeeklyStochasticSeries)
	ch.timeLabels.SetData(ts)
	ch.setTrackLineData(ts, st.DailyMovingAverageSeries25, st.DailyMovingAverageSeries50, st.DailyMovingAverageSeries200)
}

// Update updates the Chart.
func (ch *Chart) Update() (dirty bool) {
	if ch.header.Update() {
		dirty = true
	}
	if ch.fadeIn.Update() {
		dirty = true
	}
	return dirty
}

// Render renders the Chart.
func (ch *Chart) Render(vc viewContext) {
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

	old := gfx.Alpha()
	gfx.SetAlpha(old * ch.fadeIn.Value(vc.Fudge))
	defer gfx.SetAlpha(old)

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

	// Legend labels and its cursors labels overlap and use the same rect.
	// pr.Max.X = plr.Min.X
	llr := pr

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

	// Renders trackline legend. To display inline comment out the two lines below and uncomment the last line.
	ch.renderMATrackLineLegend(pr, llr, vc.MousePos)
	ch.renderCandleTrackLineLegend(pr, llr, vc.MousePos)
	//ch.renderTrackLineLegendInline(pr, llr, vc.MousePos) // Stocks that have higher prices tend to overflow the chart on default window size
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

type chartLegendLabel struct {
	percent float32
	text    string
	size    image.Point
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

func (ch *Chart) setTrackLineData(ts *model.TradingSessionSeries, ms25 *model.MovingAverageSeries, ms50 *model.MovingAverageSeries, ms200 *model.MovingAverageSeries) {
	ch.tlTradingSessions = ts
	ch.tlMovingAverage25 = ms25
	ch.tlMovingAverage50 = ms50
	ch.tlMovingAverage200 = ms200
}

func (ch *Chart) chartLegendLabelText(candle *model.TradingSession, ma25 *model.MovingAverage, ma50 *model.MovingAverage, ma200 *model.MovingAverage) string {
	legendOHLCVC := string(strings.Join([]string{"O:", strconv.FormatFloat(float64(candle.Open), 'f', 2, 32), "   H:", strconv.FormatFloat(float64(candle.High), 'f', 2, 32), "   L:", strconv.FormatFloat(float64(candle.Low), 'f', 2, 32), "   C:", strconv.FormatFloat(float64(candle.Close), 'f', 2, 32), "   V:", strconv.Itoa(candle.Volume), "   ", strconv.FormatFloat(float64(candle.Change), 'f', 2, 32), "%"}, ""))
	legendMA := strings.Join([]string{" MA25:", strconv.FormatFloat(float64(ma25.Value), 'f', 1, 32), "   MA50:", strconv.FormatFloat(float64(ma50.Value), 'f', 1, 32), "   MA200:", strconv.FormatFloat(float64(ma200.Value), 'f', 1, 32)}, "")
	legend := strings.Join([]string{legendOHLCVC, legendMA}, "   ")
	return fmt.Sprintf("%s", legend)
}

func (ch *Chart) chartLegendCandleLabelText(candle *model.TradingSession) string {
	legendOHLCVC := string(strings.Join([]string{"O:", strconv.FormatFloat(float64(candle.Open), 'f', 2, 32), "   H:", strconv.FormatFloat(float64(candle.High), 'f', 2, 32), "   L:", strconv.FormatFloat(float64(candle.Low), 'f', 2, 32), "   C:", strconv.FormatFloat(float64(candle.Close), 'f', 2, 32), "   V:", strconv.Itoa(candle.Volume), "   ", strconv.FormatFloat(float64(candle.Change), 'f', 2, 32), "%"}, ""))
	return fmt.Sprintf("%s", legendOHLCVC)
}

func (ch *Chart) chartLegendMALabelText(ma25 *model.MovingAverage, ma50 *model.MovingAverage, ma200 *model.MovingAverage) string {
	legendMA := strings.Join([]string{"MA25:", strconv.FormatFloat(float64(ma25.Value), 'f', 1, 32), "   MA50:", strconv.FormatFloat(float64(ma50.Value), 'f', 1, 32), "   MA200:", strconv.FormatFloat(float64(ma200.Value), 'f', 1, 32)}, "")
	return fmt.Sprintf("%s", legendMA)
}

func (ch *Chart) renderCandleTrackLineLegend(mainRect, labelRect image.Rectangle, mousePos image.Point) {
	if !mousePos.In(mainRect) {
		return
	}

	// Render candle trackline legend
	ts := ch.tlTradingSessions.TradingSessions

	pl := chartLegendLabel{
		percent: float32(mousePos.X-mainRect.Min.X) / float32(mainRect.Dx()),
	}

	i := int(math.Floor(float64(len(ts)) * float64(pl.percent)))
	if i >= len(ts) {
		i = len(ts) - 1
	}

	// Candlestick and moving average legends stacked as two chartAxisLabelBubbleSpec lines
	pl.text = ch.chartLegendCandleLabelText(ts[i])
	pl.size = chartAxisLabelTextRenderer.Measure(pl.text)

	ohlcvp := image.Point{
		X: labelRect.Min.X + labelRect.Dx()/2 - pl.size.X/2,
		Y: labelRect.Min.Y + labelRect.Dy() - pl.size.Y,
	}

	renderBubble(ohlcvp, pl.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(pl.text, ohlcvp, white)
}

func (ch *Chart) renderMATrackLineLegend(mainRect, labelRect image.Rectangle, mousePos image.Point) {
	if !mousePos.In(mainRect) {
		return
	}

	// Render moving average trackline legend
	ts := ch.tlTradingSessions.TradingSessions
	ms25 := ch.tlMovingAverage25.MovingAverages
	ms50 := ch.tlMovingAverage50.MovingAverages
	ms200 := ch.tlMovingAverage200.MovingAverages

	mal := chartLegendLabel{
		percent: float32(mousePos.X-mainRect.Min.X) / float32(mainRect.Dx()),
	}

	i := int(math.Floor(float64(len(ts)) * float64(mal.percent)))
	if i >= len(ts) {
		i = len(ts) - 1
	}

	mal.text = ch.chartLegendMALabelText(ms25[i], ms50[i], ms200[i])
	mal.size = chartAxisLabelTextRenderer.Measure(mal.text)

	MAp := image.Point{
		X: labelRect.Min.X + labelRect.Dx()/2 - mal.size.X/2,
		Y: labelRect.Min.Y + labelRect.Dy() - int(math.Floor(float64(mal.size.Y)*float64(2.4))),
	}

	renderBubble(MAp, mal.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(mal.text, MAp, white)
}

func (ch *Chart) renderTrackLineLegendInline(mainRect, labelRect image.Rectangle, mousePos image.Point) {
	if !mousePos.In(mainRect) {
		return
	}

	// Render inline trackline legend
	ts := ch.tlTradingSessions.TradingSessions
	ms25 := ch.tlMovingAverage25.MovingAverages
	ms50 := ch.tlMovingAverage50.MovingAverages
	ms200 := ch.tlMovingAverage200.MovingAverages

	l := chartLegendLabel{
		percent: float32(mousePos.X-mainRect.Min.X) / float32(mainRect.Dx()),
	}

	i := int(math.Floor(float64(len(ts)) * float64(l.percent)))
	if i >= len(ts) {
		i = len(ts) - 1
	}

	l.text = ch.chartLegendLabelText(ts[i], ms25[i], ms50[i], ms200[i])
	l.size = chartAxisLabelTextRenderer.Measure(l.text)

	tp := image.Point{
		X: labelRect.Min.X + labelRect.Dx()/2 - l.size.X/2,
		Y: labelRect.Min.Y + labelRect.Dy() - l.size.Y,
	}

	renderBubble(tp, l.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)

}
