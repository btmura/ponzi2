package chart

import (
	"fmt"
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/app/view/status"
	"github.com/btmura/ponzi2/internal/app/view/text"
	"github.com/btmura/ponzi2/internal/app/view/vao"
	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/log"
)

// Embed resources into the application. Get esc from github.com/mjibson/esc.
//go:generate esc -o bindata.go -pkg chart -include ".*(ply|png)" -modtime 1337 -private data

const (
	chartRounding       = 10
	chartSectionPadding = 5
	chartTextPadding    = 20
)

var (
	chartSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 24)
	chartQuotePrinter            = func(q *model.Quote) string { return status.Join(status.PriceChange(q), status.SourceUpdate(q)) }
)

const axisLabelPadding = 4

var (
	axisLabelBubble       = rect.NewBubble(6)
	axisLabelTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
)

var (
	cursorHorizLine = vao.HorizLine(color.LightGray)
	cursorVertLine  = vao.VertLine(color.LightGray)
)

// ZoomChange specifies whether the user has zoomed in or not.
type ZoomChange int

// ZoomChange values.
//go:generate stringer -type=ZoomChange
const (
	ZoomChangeUnspecified ZoomChange = iota
	ZoomIn
	ZoomOut
)

// Chart shows a stock chart for a single stock.
type Chart struct {
	// frameBubble is the border around the entire chart.
	frameBubble *rect.Bubble

	// header renders the header with the symbol, quote, and buttons.
	header *header

	prices        *price
	priceAxis     *priceAxis
	priceCursor   *priceCursor
	priceTimeline *timeline

	movingAverage25  *movingAverage
	movingAverage50  *movingAverage
	movingAverage200 *movingAverage

	volume         *volume
	volumeAxis     *volumeAxis
	volumeCursor   *volumeCursor
	volumeTimeline *timeline

	dailyStochastics        *stochastic
	dailyStochasticAxis     *stochasticAxis
	dailyStochasticCursor   *stochasticCursor
	dailyStochasticTimeline *timeline

	weeklyStochastics        *stochastic
	weeklyStochasticAxis     *stochasticAxis
	weeklyStochasticCursor   *stochasticCursor
	weeklyStochasticTimeline *timeline

	timelineAxis   *timelineAxis
	timelineCursor *timelineCursor

	legend *legend

	// loadingTextBox renders the loading text shown when loading from a fresh state.
	loadingTextBox *text.Box

	// errorTextBox renders the error text.
	errorTextBox *text.Box

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

	// zoomChangeCallback is fired when the zoom is changed. Nil if no callback registered.
	zoomChangeCallback func(zoomChange ZoomChange)
}

// NewChart creates a new Chart.
func NewChart(fps int) *Chart {
	return &Chart{
		frameBubble: rect.NewBubble(chartRounding),
		header: newHeader(&headerArgs{
			SymbolQuoteTextRenderer: chartSymbolQuoteTextRenderer,
			QuotePrinter:            chartQuotePrinter,
			ShowRefreshButton:       true,
			ShowAddButton:           true,
			Rounding:                chartRounding,
			Padding:                 chartSectionPadding,
			FPS:                     fps,
		}),

		prices:        new(price),
		priceAxis:     new(priceAxis),
		priceCursor:   new(priceCursor),
		priceTimeline: new(timeline),

		movingAverage25:  newMovingAverage(color.Purple),
		movingAverage50:  newMovingAverage(color.Yellow),
		movingAverage200: newMovingAverage(color.White),

		volume:         new(volume),
		volumeAxis:     new(volumeAxis),
		volumeCursor:   new(volumeCursor),
		volumeTimeline: new(timeline),

		dailyStochastics:        newStochastic(color.Yellow),
		dailyStochasticAxis:     new(stochasticAxis),
		dailyStochasticCursor:   new(stochasticCursor),
		dailyStochasticTimeline: new(timeline),

		weeklyStochastics:        newStochastic(color.Purple),
		weeklyStochasticAxis:     new(stochasticAxis),
		weeklyStochasticCursor:   new(stochasticCursor),
		weeklyStochasticTimeline: new(timeline),

		timelineAxis:   new(timelineAxis),
		timelineCursor: new(timelineCursor),

		legend: new(legend),

		loadingTextBox: text.NewBox(chartSymbolQuoteTextRenderer, "LOADING...", text.Padding(chartTextPadding)),
		errorTextBox:   text.NewBox(chartSymbolQuoteTextRenderer, "ERROR", text.Color(color.Orange), text.Padding(chartTextPadding)),
		loading:        true,
		fadeIn:         animation.New(1 * fps),
	}
}

// SetLoading toggles the Chart's loading indicator.
func (ch *Chart) SetLoading(loading bool) {
	ch.loading = loading
	ch.header.SetLoading(loading)
}

// SetError toggles the Chart's error indicator.
func (ch *Chart) SetError(err error) {
	ch.hasError = err != nil
	if err != nil {
		ch.errorTextBox.SetText(fmt.Sprintf("ERROR: %v", err))
	}
	ch.header.SetError(err)
}

// Data is argument to SetData.
type Data struct {
	// Symbol is a required stock symbol.
	Symbol string

	// Quote is an optional quote. Nil when no data has been received yet.
	Quote *model.Quote

	// Chart is optional chart data. Nil when data hasn't been received yet.
	Chart *model.Chart
}

// SetData sets the data to be shown on the chart.
func (ch *Chart) SetData(data *Data) error {
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

	switch dc.Range {
	case model.OneDay:
		ch.showMovingAverages = false
		ch.showStochastics = false
	case model.OneYear:
		ch.showMovingAverages = true
		ch.showStochastics = true
	default:
		return errors.Errorf("bad range: %v", dc.Range)
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

func (ch *Chart) SetBounds(bounds image.Rectangle) {
	ch.fullBounds = bounds
}

func (ch *Chart) ProcessInput(input *view.Input) {
	bounds := ch.fullBounds

	ch.frameBubble.SetBounds(bounds)

	ch.header.SetBounds(bounds)
	r, _ := ch.header.ProcessInput(input)

	ch.bodyBounds = r
	ch.loadingTextBox.SetBounds(r)
	ch.errorTextBox.SetBounds(r)

	// Calculate percentage needed for each section.
	const (
		volumePercent            = 0.13
		dailyStochasticsPercent  = 0.13
		weeklyStochasticsPercent = 0.13
	)
	timeLabelsPercent := float32(ch.timelineAxis.MaxLabelSize.Y+chartSectionPadding*2) / float32(r.Dy())

	// Divide up the rectangle into sections.
	var rects []image.Rectangle
	if ch.showStochastics {
		rects = rect.Slice(r, timeLabelsPercent, weeklyStochasticsPercent, dailyStochasticsPercent, volumePercent)
	} else {
		rects = rect.Slice(r, timeLabelsPercent, volumePercent)
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
	maxWidth += chartSectionPadding

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
	pr = pr.Inset(chartSectionPadding)
	vr = vr.Inset(chartSectionPadding)
	dr = dr.Inset(chartSectionPadding)
	wr = wr.Inset(chartSectionPadding)
	tr = tr.Inset(chartSectionPadding)

	plr = plr.Inset(chartSectionPadding)
	vlr = vlr.Inset(chartSectionPadding)
	dlr = dlr.Inset(chartSectionPadding)
	wlr = wlr.Inset(chartSectionPadding)
	tlr = tlr.Inset(chartSectionPadding)

	ch.prices.SetBounds(pr)
	ch.priceCursor.SetBounds(pr, plr)
	ch.priceCursor.ProcessInput(input)
	ch.priceTimeline.SetBounds(pr)
	ch.movingAverage25.SetBounds(pr)
	ch.movingAverage50.SetBounds(pr)
	ch.movingAverage200.SetBounds(pr)

	ch.volume.SetBounds(vr)
	ch.volumeCursor.SetBounds(vr, vlr)
	ch.volumeCursor.ProcessInput(input)
	ch.volumeTimeline.SetBounds(vr)

	ch.dailyStochastics.SetBounds(dr)
	ch.dailyStochasticCursor.SetBounds(dr, dlr)
	ch.dailyStochasticCursor.ProcessInput(input)
	ch.dailyStochasticTimeline.SetBounds(dr)

	ch.weeklyStochastics.SetBounds(wr)
	ch.weeklyStochasticCursor.SetBounds(wr, wlr)
	ch.weeklyStochasticCursor.ProcessInput(input)
	ch.weeklyStochasticTimeline.SetBounds(wr)

	ch.timelineAxis.SetBounds(tr)
	ch.timelineCursor.SetBounds(tr, tlr)
	ch.timelineCursor.ProcessInput(input)

	ch.priceAxis.SetBounds(plr)
	ch.volumeAxis.SetBounds(vlr)
	ch.dailyStochasticAxis.SetBounds(dlr)
	ch.weeklyStochasticAxis.SetBounds(wlr)

	ch.legend.SetBounds(pr, llr)
	ch.legend.ProcessInput(input)

	if input.MouseScrolled.In(bounds) && ch.zoomChangeCallback != nil {
		zoomChange := ZoomChangeUnspecified
		switch input.MouseScrolled.Direction {
		case view.ScrollDown:
			zoomChange = ZoomOut
		case view.ScrollUp:
			zoomChange = ZoomIn
		default:
			log.Error("mouse scroll event missing direction")
			return
		}

		input.ScheduledCallbacks = append(input.ScheduledCallbacks, func() {
			if ch.zoomChangeCallback != nil {
				ch.zoomChangeCallback(zoomChange)
			}
		})
	}
}

// Update updates the Chart.
func (ch *Chart) Update() (dirty bool) {
	if ch.header.Update() {
		dirty = true
	}
	if ch.loadingTextBox.Update() {
		dirty = true
	}
	if ch.errorTextBox.Update() {
		dirty = true
	}
	if ch.fadeIn.Update() {
		dirty = true
	}
	return dirty
}

// Render renders the Chart.
func (ch *Chart) Render(fudge float32) {
	ch.frameBubble.Render(fudge)
	ch.header.Render(fudge)

	r := ch.bodyBounds
	rect.RenderLineAtTop(r)

	// Only show messages if no prior data to show.
	if !ch.hasStockUpdated {
		if ch.loading {
			ch.loadingTextBox.Render(fudge)
			return
		}

		if ch.hasError {
			ch.errorTextBox.Render(fudge)
			return
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
	timeLabelsPercent := float32(ch.timelineAxis.MaxLabelSize.Y+chartSectionPadding*2) / float32(r.Dy())

	// Divide up the rectangle into sections.
	var rects []image.Rectangle
	if ch.showStochastics {
		rects = rect.Slice(r, timeLabelsPercent, weeklyStochasticsPercent, dailyStochasticsPercent, volumePercent)
	} else {
		rects = rect.Slice(r, timeLabelsPercent, volumePercent)
	}

	// Render the dividers between the sections.
	for i := 0; i < len(rects)-1; i++ {
		rect.RenderLineAtTop(rects[i])
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
}

// SetRefreshButtonClickCallback sets the callback for refresh button clicks.
func (ch *Chart) SetRefreshButtonClickCallback(cb func()) {
	ch.header.SetRefreshButtonClickCallback(cb)
}

// SetAddButtonClickCallback sets the callback for add button clicks.
func (ch *Chart) SetAddButtonClickCallback(cb func()) {
	ch.header.SetAddButtonClickCallback(cb)
}

// SetZoomChangeCallback sets the callback for zoom changes.
func (ch *Chart) SetZoomChangeCallback(cb func(zoomChange ZoomChange)) {
	ch.zoomChangeCallback = cb
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
	ch.zoomChangeCallback = nil
}

func renderCursorLines(r image.Rectangle, mousePos image.Point) {
	if mousePos.In(r) {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, mousePos.Y, r.Max.X, mousePos.Y))
		cursorHorizLine.Render()
	}

	if mousePos.X >= r.Min.X && mousePos.X <= r.Max.X {
		gfx.SetModelMatrixRect(image.Rect(mousePos.X, r.Min.Y, mousePos.X, r.Max.Y))
		cursorVertLine.Render()
	}
}
