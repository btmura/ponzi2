package view

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

	"golang.org/x/image/font/gofont/goregular"

	"gitlab.com/btmura/ponzi2/internal/app/gfx"
	"gitlab.com/btmura/ponzi2/internal/app/model"
	"gitlab.com/btmura/ponzi2/internal/util"
)

const (
	chartRounding                   = 10
	chartPadding                    = 5
	chartAxisVerticalPaddingPercent = .05
)

var (
	chartSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 24)
	chartQuotePrinter            = func(q *model.Quote) string { return join(priceStatus(q), updateStatus(q)) }
	chartLoadingText             = newCenteredText(chartSymbolQuoteTextRenderer, "LOADING...")
	chartErrorText               = newCenteredText(chartSymbolQuoteTextRenderer, "ERROR", centeredTextColor(orange))
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

	// showMovingAverages is whether to render the moving averages.
	showMovingAverages bool

	// showStochastics is whether to show stochastics.
	showStochastics bool
}

// NewChart creates a new Chart.
func NewChart() *Chart {
	return &Chart{
		header: newChartHeader(&chartHeaderArgs{
			SymbolQuoteTextRenderer: chartSymbolQuoteTextRenderer,
			QuotePrinter:            chartQuotePrinter,
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

// SetLoading toggles the Chart's loading indicator.
func (ch *Chart) SetLoading(loading bool) {
	ch.loading = loading
	ch.header.SetLoading(loading)
}

// SetError toggles the Chart's error indicator.
func (ch *Chart) SetError(error bool) {
	ch.hasError = error
	ch.header.SetError(error)
}

// ChartData has the data to be shown on the chart.
type ChartData struct {
	Symbol string
	Quote  *model.Quote
	Chart  *model.Chart
}

// SetData sets the data to be shown on the chart.
func (ch *Chart) SetData(data *ChartData) error {
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

	switch dc.Range {
	case model.OneDay:
		ch.showMovingAverages = false
		ch.showStochastics = false
	case model.OneYear:
		ch.showMovingAverages = true
		ch.showStochastics = true
	default:
		return util.Errorf("bad range: %v", dc.Range)
	}

	ts := dc.TradingSessionSeries

	if err := ch.timeLines.SetData(dc.Range, ts); err != nil {
		return err
	}

	ch.prices.SetData(ts)

	if ch.showMovingAverages {
		ch.movingAverage25.SetData(ts, dc.MovingAverageSeries25)
		ch.movingAverage50.SetData(ts, dc.MovingAverageSeries50)
		ch.movingAverage200.SetData(ts, dc.MovingAverageSeries200)
	}

	ch.volume.SetData(ts)

	if ch.showStochastics {
		ch.dailyStochastics.SetData(dc.DailyStochasticSeries)
		ch.weeklyStochastics.SetData(dc.WeeklyStochasticSeries)
	}

	if err := ch.timeLabels.SetData(dc.Range, ts); err != nil {
		return err
	}

	ch.setTrackLineData(ts, dc.MovingAverageSeries25, dc.MovingAverageSeries50, dc.MovingAverageSeries200)

	return nil
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
func (ch *Chart) Render(vc viewContext) error {
	// Render the border around the chart.
	strokeRoundedRect(vc.Bounds, chartRounding)

	// Render the header and the line below it.
	r, _ := ch.header.Render(vc)
	renderRectTopDivider(r, horizLine)

	// Only show messages if no prior data to show.
	if !ch.hasStockUpdated {
		if ch.loading {
			chartLoadingText.Render(r)
			return nil
		}

		if ch.hasError {
			chartErrorText.Render(r)
			return nil
		}
	}

	old := gfx.Alpha()
	gfx.SetAlpha(old * ch.fadeIn.Value(vc.Fudge))
	defer gfx.SetAlpha(old)

	// Calculate percentage needed for each section.
	const (
		volumePercent            = 0.13
		dailyStochasticsPercent  = 0.13
		weeklyStochasticsPercent = 0.13
	)
	timeLabelsPercent := float32(ch.timeLabels.MaxLabelSize.Y+chartPadding*2) / float32(r.Dy())

	// Divide up the rectangle into sections.
	var rects []image.Rectangle
	if ch.showStochastics {
		rects = sliceRect(r, timeLabelsPercent, weeklyStochasticsPercent, dailyStochasticsPercent, volumePercent)
	} else {
		rects = sliceRect(r, timeLabelsPercent, volumePercent)
	}

	// Render the dividers between the sections.
	for i := 0; i < len(rects)-1; i++ {
		renderRectTopDivider(rects[i], horizLine)
	}

	var pr, vr, dr, wr, tr image.Rectangle
	if ch.showStochastics {
		pr, vr, dr, wr, tr = rects[4], rects[3], rects[2], rects[1], rects[0]
	} else {
		pr, vr, tr = rects[2], rects[1], rects[0]
	}

	// Create separate rects for each section's labels shown on the right.
	plr, vlr, dlr, wlr := pr, vr, dr, wr

	// Figure out width to trim off on the right of each rect for the labels.
	maxWidth := ch.prices.MaxLabelSize.X
	if w := ch.volume.MaxLabelSize.X; w > maxWidth {
		maxWidth = w
	}
	if ch.showStochastics {
		if w := ch.dailyStochastics.MaxLabelSize.X; w > maxWidth {
			maxWidth = w
		}
		if w := ch.weeklyStochastics.MaxLabelSize.X; w > maxWidth {
			maxWidth = w
		}
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
	tr.Max.X = plr.Min.X
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
	if ch.showStochastics {
		ch.timeLines.Render(dr)
		ch.timeLines.Render(wr)
	}

	ch.prices.Render(pr)
	if ch.showMovingAverages {
		ch.movingAverage25.Render(pr)
		ch.movingAverage50.Render(pr)
		ch.movingAverage200.Render(pr)
	}
	ch.volume.Render(vr)
	if ch.showStochastics {
		ch.dailyStochastics.Render(dr)
		ch.weeklyStochastics.Render(wr)
	}
	ch.timeLabels.Render(tr)

	ch.prices.RenderAxisLabels(plr)
	ch.volume.RenderAxisLabels(vlr)
	if ch.showStochastics {
		ch.dailyStochastics.RenderAxisLabels(dlr)
		ch.weeklyStochastics.RenderAxisLabels(wlr)
	}

	renderCursorLines(pr, vc.MousePos)
	renderCursorLines(vr, vc.MousePos)
	if ch.showStochastics {
		renderCursorLines(dr, vc.MousePos)
		renderCursorLines(wr, vc.MousePos)
	}

	ch.prices.RenderCursorLabels(pr, plr, vc.MousePos)
	ch.volume.RenderCursorLabels(vr, vlr, vc.MousePos)
	if ch.showStochastics {
		ch.dailyStochastics.RenderCursorLabels(dr, dlr, vc.MousePos)
		ch.weeklyStochastics.RenderCursorLabels(wr, wlr, vc.MousePos)
	}

	if err := ch.timeLabels.RenderCursorLabels(tr, tlr, vc.MousePos); err != nil {
		return err
	}

	// Renders trackline legend. To display inline comment out the two lines below and uncomment the last line.
	if ch.showMovingAverages {
		ch.renderMATrackLineLegend(pr, llr, vc.MousePos)
	}
	ch.renderCandleTrackLineLegend(pr, llr, vc.MousePos)
	//ch.renderTrackLineLegendInline(pr, llr, vc.MousePos) // Stocks that have higher prices tend to overflow the chart on default window size

	return nil
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
