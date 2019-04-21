package view

import (
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/status"
)

const (
	chartRounding                   = 10
	chartPadding                    = 5
	chartAxisVerticalPaddingPercent = .05
)

var (
	chartSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 24)
	chartQuotePrinter            = func(q *model.Quote) string { return join(priceStatus(q), updateStatus(q)) }
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
	header *chartHeader

	prices        *chartPrices
	priceAxis     *chartPriceAxis
	priceCursor   *chartPriceCursor
	priceTimeline *chartTimeline

	movingAverage25  *chartMovingAverage
	movingAverage50  *chartMovingAverage
	movingAverage200 *chartMovingAverage

	volume         *chartVolume
	volumeAxis     *chartVolumeAxis
	volumeCursor   *chartVolumeCursor
	volumeTimeline *chartTimeline

	dailyStochastics        *chartStochastics
	dailyStochasticAxis     *chartStochasticAxis
	dailyStochasticCursor   *chartStochasticCursor
	dailyStochasticTimeline *chartTimeline

	weeklyStochastics        *chartStochastics
	weeklyStochasticAxis     *chartStochasticAxis
	weeklyStochasticCursor   *chartStochasticCursor
	weeklyStochasticTimeline *chartTimeline

	timelineAxis   *chartTimelineAxis
	timelineCursor *chartTimelineCursor

	legend *chartLegend

	// loadingText is the text shown when loading from a fresh state.
	loadingText *centeredText

	// errorText is the text shown when an error occurs from a fresh state.
	errorText *centeredText

	// loading is true when data for the stock is being retrieved.
	loading bool

	// hasStockUpdated is true if the stock has reported data.
	hasStockUpdated bool

	// hasError is true there was a loading issue.
	hasError bool

	// fadeIn fades in the data after it loads.
	fadeIn *animation.Animation

	// showMovingAverages is whether to render the moving averages.
	showMovingAverages bool

	// showStochastics is whether to show stochastics.
	showStochastics bool

	// fullBounds is the rect with global coords that should be drawn within.
	fullBounds image.Rectangle

	// bodyBounds is a sub-rect of fullBounds without the header.
	bodyBounds image.Rectangle

	// mousePos is the current mouse position.
	mousePos image.Point
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

		prices:        newChartPrices(),
		priceAxis:     new(chartPriceAxis),
		priceCursor:   new(chartPriceCursor),
		priceTimeline: new(chartTimeline),

		movingAverage25:  newChartMovingAverage(purple),
		movingAverage50:  newChartMovingAverage(yellow),
		movingAverage200: newChartMovingAverage(white),

		volume:         newChartVolume(),
		volumeAxis:     new(chartVolumeAxis),
		volumeCursor:   new(chartVolumeCursor),
		volumeTimeline: new(chartTimeline),

		dailyStochastics:        newChartStochastics(yellow),
		dailyStochasticAxis:     new(chartStochasticAxis),
		dailyStochasticCursor:   new(chartStochasticCursor),
		dailyStochasticTimeline: new(chartTimeline),

		weeklyStochastics:        newChartStochastics(purple),
		weeklyStochasticAxis:     new(chartStochasticAxis),
		weeklyStochasticCursor:   new(chartStochasticCursor),
		weeklyStochasticTimeline: new(chartTimeline),

		timelineAxis:   new(chartTimelineAxis),
		timelineCursor: new(chartTimelineCursor),

		legend: newChartLegend(),

		loadingText: newCenteredText(chartSymbolQuoteTextRenderer, "LOADING..."),
		errorText:   newCenteredText(chartSymbolQuoteTextRenderer, "ERROR", centeredTextColor(orange)),
		loading:     true,
		fadeIn:      animation.New(1 * fps),
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

	switch dc.Range {
	case model.OneDay:
		ch.showMovingAverages = false
		ch.showStochastics = false
	case model.OneYear:
		ch.showMovingAverages = true
		ch.showStochastics = true
	default:
		return status.Errorf("bad range: %v", dc.Range)
	}

	ts := dc.TradingSessionSeries

	ch.prices.SetData(ts)
	ch.priceAxis.SetData(ts)
	ch.priceCursor.SetData(ts)
	if err := ch.priceTimeline.SetData(dc.Range, ts); err != nil {
		return err
	}

	if ch.showMovingAverages {
		ch.movingAverage25.SetData(ts, dc.MovingAverageSeries25)
		ch.movingAverage50.SetData(ts, dc.MovingAverageSeries50)
		ch.movingAverage200.SetData(ts, dc.MovingAverageSeries200)
	}

	ch.volume.SetData(ts)
	ch.volumeAxis.SetData(ts)
	ch.volumeCursor.SetData(ts)
	if err := ch.volumeTimeline.SetData(dc.Range, ts); err != nil {
		return err
	}

	if ch.showStochastics {
		ch.dailyStochastics.SetData(dc.DailyStochasticSeries)
		ch.dailyStochasticAxis.SetData(dc.DailyStochasticSeries)
		ch.dailyStochasticCursor.SetData()
		if err := ch.dailyStochasticTimeline.SetData(dc.Range, ts); err != nil {
			return err
		}

		ch.weeklyStochastics.SetData(dc.WeeklyStochasticSeries)
		ch.weeklyStochasticAxis.SetData(dc.WeeklyStochasticSeries)
		ch.weeklyStochasticCursor.SetData()
		if err := ch.weeklyStochasticTimeline.SetData(dc.Range, ts); err != nil {
			return err
		}
	}

	if err := ch.timelineAxis.SetData(dc.Range, ts); err != nil {
		return err
	}

	if err := ch.timelineCursor.SetData(dc.Range, ts); err != nil {
		return err
	}

	ch.legend.SetData(ts, dc.MovingAverageSeries25, dc.MovingAverageSeries50, dc.MovingAverageSeries200, ch.showMovingAverages)

	return nil
}

// ProcessInput processes input.
func (ch *Chart) ProcessInput(ic inputContext) {
	ch.fullBounds = ic.Bounds
	ch.mousePos = ic.MousePos

	r, _ := ch.header.ProcessInput(ic)
	ch.bodyBounds = r

	ch.loadingText.ProcessInput(ic.Bounds)
	ch.errorText.ProcessInput(ic.Bounds)

	// Calculate percentage needed for each section.
	const (
		volumePercent            = 0.13
		dailyStochasticsPercent  = 0.13
		weeklyStochasticsPercent = 0.13
	)
	timeLabelsPercent := float32(ch.timelineAxis.MaxLabelSize.Y+chartPadding*2) / float32(r.Dy())

	// Divide up the rectangle into sections.
	var rects []image.Rectangle
	if ch.showStochastics {
		rects = sliceRect(r, timeLabelsPercent, weeklyStochasticsPercent, dailyStochasticsPercent, volumePercent)
	} else {
		rects = sliceRect(r, timeLabelsPercent, volumePercent)
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

	ic.Bounds = pr
	ch.prices.ProcessInput(ic)
	ch.priceCursor.ProcessInput(pr, plr, ch.mousePos)
	ch.priceTimeline.ProcessInput(pr)
	ch.movingAverage25.ProcessInput(ic)
	ch.movingAverage50.ProcessInput(ic)
	ch.movingAverage200.ProcessInput(ic)

	ic.Bounds = vr
	ch.volume.ProcessInput(ic)
	ch.volumeCursor.ProcessInput(vr, vlr, ch.mousePos)
	ch.volumeTimeline.ProcessInput(vr)

	ic.Bounds = dr
	ch.dailyStochastics.ProcessInput(ic)
	ch.dailyStochasticCursor.ProcessInput(dr, dlr, ch.mousePos)
	ch.dailyStochasticTimeline.ProcessInput(dr)

	ic.Bounds = wr
	ch.weeklyStochastics.ProcessInput(ic)
	ch.weeklyStochasticCursor.ProcessInput(wr, wlr, ch.mousePos)
	ch.weeklyStochasticTimeline.ProcessInput(wr)

	ch.timelineAxis.ProcessInput(tr)
	ch.timelineCursor.ProcessInput(tr, tlr, ch.mousePos)

	ch.priceAxis.ProcessInput(plr)
	ch.volumeAxis.ProcessInput(vlr)
	ch.dailyStochasticAxis.ProcessInput(dlr)
	ch.weeklyStochasticAxis.ProcessInput(wlr)

	ch.legend.ProcessInput(pr, llr, ic.MousePos)
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
func (ch *Chart) Render(fudge float32) error {
	// Render the border around the chart.
	strokeRoundedRect(ch.fullBounds, chartRounding)

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

	// Calculate percentage needed for each section.
	const (
		volumePercent            = 0.13
		dailyStochasticsPercent  = 0.13
		weeklyStochasticsPercent = 0.13
	)
	timeLabelsPercent := float32(ch.timelineAxis.MaxLabelSize.Y+chartPadding*2) / float32(r.Dy())

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

	ch.priceTimeline.Render(fudge)
	ch.volumeTimeline.Render(fudge)
	if ch.showStochastics {
		ch.dailyStochasticTimeline.Render(fudge)
		ch.weeklyStochasticTimeline.Render(fudge)
	}

	ch.prices.Render(fudge)
	ch.priceAxis.Render(fudge)
	ch.priceCursor.Render(fudge)

	if ch.showMovingAverages {
		ch.movingAverage25.Render(fudge)
		ch.movingAverage50.Render(fudge)
		ch.movingAverage200.Render(fudge)
	}

	ch.volume.Render(fudge)
	ch.volumeAxis.Render(fudge)
	ch.volumeCursor.Render(fudge)

	if ch.showStochastics {
		ch.dailyStochastics.Render(fudge)
		ch.dailyStochasticAxis.Render(fudge)
		ch.dailyStochasticCursor.Render(fudge)

		ch.weeklyStochastics.Render(fudge)
		ch.weeklyStochasticAxis.Render(fudge)
		ch.weeklyStochasticCursor.Render(fudge)
	}

	ch.timelineAxis.Render(fudge)
	ch.timelineCursor.Render(fudge)

	ch.legend.Render(fudge)

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
	ch.header.Close()
	ch.prices.Close()
	ch.priceAxis.Close()
	ch.priceCursor.Close()
	ch.priceTimeline.Close()
	ch.movingAverage25.Close()
	ch.movingAverage50.Close()
	ch.movingAverage200.Close()
	ch.volume.Close()
	ch.volumeAxis.Close()
	ch.volumeCursor.Close()
	ch.volumeTimeline.Close()
	ch.dailyStochastics.Close()
	ch.dailyStochasticAxis.Close()
	ch.dailyStochasticCursor.Close()
	ch.dailyStochasticTimeline.Close()
	ch.weeklyStochastics.Close()
	ch.weeklyStochasticAxis.Close()
	ch.weeklyStochasticCursor.Close()
	ch.weeklyStochasticTimeline.Close()
	ch.timelineAxis.Close()
	ch.timelineCursor.Close()
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
