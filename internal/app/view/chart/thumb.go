package chart

import (
	"fmt"
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/app/view/status"
	"github.com/btmura/ponzi2/internal/app/view/text"
	"github.com/btmura/ponzi2/internal/logger"
)

const (
	thumbRounding       = 6
	thumbSectionPadding = 2
	thumbTextPadding    = 10
	thumbVolumePercent  = 0.4
)

var (
	thumbSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	thumbQuotePrinter            = func(q *model.Quote) string { return status.PriceChange(q) }
)

// Thumb shows a thumbnail for a stock.
type Thumb struct {
	// frameBubble is the border around the entire thumb.
	frameBubble *rect.Bubble

	// header renders the header with the symbol, quote, and buttons.
	header *header

	prices        *price
	priceCursor   *priceCursor
	priceTimeline *timeline

	movingAverage21  *movingAverage
	movingAverage50  *movingAverage
	movingAverage200 *movingAverage

	volume         *volume
	volumeCursor   *volumeCursor
	volumeTimeline *timeline

	// loadingTextBox renders the loading text shown when loading from a fresh state.
	loadingTextBox *text.Box

	// errorTextBox renders the error text.
	errorTextBox *text.Box

	// thumbClickCallback is the callback to schedule when the thumb is clicked.
	thumbClickCallback func()

	// loading is true when data for the stock is being retrieved.
	loading bool

	// hasStockUpdated is true if the stock has reported data.
	hasStockUpdated bool

	// hasError is true there was a loading issue.
	hasError bool

	// fadeIn fades in the data after it loads.
	fadeIn *animation.Animation

	// bounds is the rect with global coords that should be drawn within.
	bounds image.Rectangle

	// bodyBounds is a sub-rect of bounds without the header.
	bodyBounds image.Rectangle

	// sectionDividers are bounds of the sections inside the body to render dividers.
	sectionDividers []image.Rectangle
}

// NewThumb creates a Thumb.
func NewThumb(priceStyle PriceStyle) *Thumb {
	if priceStyle == PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return nil
	}

	return &Thumb{
		frameBubble: rect.NewBubble(thumbRounding),

		header: newHeader(&headerArgs{
			SymbolQuoteTextRenderer: thumbSymbolQuoteTextRenderer,
			QuotePrinter:            thumbQuotePrinter,
			ShowRemoveButton:        true,
			Rounding:                thumbRounding,
			Padding:                 thumbSectionPadding,
		}),

		prices:        newPrice(priceStyle),
		priceCursor:   new(priceCursor),
		priceTimeline: newTimeline(view.TransparentLightGray, view.LightGray, view.TransparentGray, view.Gray),

		movingAverage21:  newMovingAverage(view.Green),
		movingAverage50:  newMovingAverage(view.Red),
		movingAverage200: newMovingAverage(view.White),

		volume:         newVolume(priceStyle),
		volumeCursor:   new(volumeCursor),
		volumeTimeline: newTimeline(view.LightGray, view.TransparentLightGray, view.Gray, view.TransparentGray),

		loadingTextBox: text.NewBox(thumbSymbolQuoteTextRenderer, "LOADING...", text.Padding(thumbTextPadding)),
		errorTextBox:   text.NewBox(thumbSymbolQuoteTextRenderer, "ERROR", text.Color(view.Orange), text.Padding(thumbTextPadding)),
		loading:        true,
		fadeIn:         animation.New(1 * view.FPS),
	}
}

// SetPriceStyle sets the style of the thumbnail.
func (t *Thumb) SetPriceStyle(newPriceStyle PriceStyle) {
	if newPriceStyle == PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return
	}

	t.prices.SetStyle(newPriceStyle)
	t.volume.SetStyle(newPriceStyle)
}

// SetLoading toggles the Chart's loading indicator.
func (t *Thumb) SetLoading(loading bool) {
	t.loading = loading
	t.header.SetLoading(loading)
}

// SetError toggles the Chart's error indicator.
func (t *Thumb) SetError(err error) {
	t.hasError = err != nil
	if err != nil {
		t.errorTextBox.SetText(fmt.Sprintf("ERROR: %v", err))
	}
	t.header.SetError(err)
}

// SetData sets the data to be shown on the chart.
func (t *Thumb) SetData(data Data) {
	if !t.hasStockUpdated && data.Chart != nil {
		t.fadeIn.Start()
	}
	t.hasStockUpdated = data.Chart != nil

	t.header.SetData(data)

	dc := data.Chart
	if dc == nil {
		return
	}

	ts := dc.TradingSessionSeries
	m21 := dc.MovingAverageSeries21
	m50 := dc.MovingAverageSeries50
	m200 := dc.MovingAverageSeries200
	vs := dc.AverageVolumeSeries

	if ts == nil || m21 == nil || m50 == nil || m200 == nil || vs == nil {
		return
	}

	if l1, l2, l3, l4, l5 := len(ts.TradingSessions), len(m21.MovingAverages), len(m50.MovingAverages), len(m200.MovingAverages), len(vs.AverageVolumes); l1 != l2 || l2 != l3 || l3 != l4 || l4 != l5 {
		logger.Errorf("bad data has different lengths: %d %d %d %d %d", l1, l2, l3, l4, l5)
		return
	}

	const days = 20
	if l := len(ts.TradingSessions); l > days {
		ts = ts.DeepCopy()
		ts.TradingSessions = ts.TradingSessions[l-days:]
	}
	if l := len(m21.MovingAverages); l > days {
		m21 = m21.DeepCopy()
		m21.MovingAverages = m21.MovingAverages[l-days:]
	}
	if l := len(m50.MovingAverages); l > days {
		m50 = m50.DeepCopy()
		m50.MovingAverages = m50.MovingAverages[l-days:]
	}
	if l := len(m200.MovingAverages); l > days {
		m200 = m200.DeepCopy()
		m200.MovingAverages = m200.MovingAverages[l-days:]
	}
	if l := len(vs.AverageVolumes); l > days {
		vs = vs.DeepCopy()
		vs.AverageVolumes = vs.AverageVolumes[l-days:]
	}

	t.prices.SetData(priceData{ts})
	t.priceCursor.SetData(priceCursorData{ts})
	t.priceTimeline.SetData(timelineData{dc.Interval, ts})

	t.movingAverage21.SetData(movingAverageData{ts, m21})
	t.movingAverage50.SetData(movingAverageData{ts, m50})
	t.movingAverage200.SetData(movingAverageData{ts, m200})

	t.volume.SetData(volumeData{ts, vs})
	t.volumeCursor.SetData(volumeCursorData{ts})
	t.volumeTimeline.SetData(timelineData{dc.Interval, ts})
}

// SetBounds sets the bounds to draw within.
func (t *Thumb) SetBounds(bounds image.Rectangle) {
	t.bounds = bounds
}

// ProcessInput processes input.
func (t *Thumb) ProcessInput(input *view.Input) {
	t.frameBubble.SetBounds(t.bounds)

	t.header.SetBounds(t.bounds)
	r, clicks := t.header.ProcessInput(input)

	t.bodyBounds = r
	t.loadingTextBox.SetBounds(r)
	t.errorTextBox.SetBounds(r)

	if !clicks.HasClicks() && input.MouseLeftButtonClicked.In(t.bounds) {
		input.AddFiredCallback(t.thumbClickCallback)
	}

	// Divide up the rectangle into sections.
	rects := rect.Slice(r, thumbVolumePercent)

	pr, vr := rects[1], rects[0]

	t.sectionDividers = []image.Rectangle{vr}

	// Pad all the rects.
	pr = pr.Inset(thumbSectionPadding)
	vr = vr.Inset(thumbSectionPadding)

	t.prices.SetBounds(pr)
	t.priceCursor.SetBounds(pr, pr)
	t.priceCursor.ProcessInput(input)
	t.priceTimeline.SetBounds(pr)
	t.movingAverage21.SetBounds(pr)
	t.movingAverage50.SetBounds(pr)
	t.movingAverage200.SetBounds(pr)

	t.volume.SetBounds(vr)
	t.volumeCursor.SetBounds(vr, vr)
	t.volumeCursor.ProcessInput(input)
	t.volumeTimeline.SetBounds(vr)
}

// Update updates the Thumb.
func (t *Thumb) Update() (dirty bool) {
	if t.header.Update() {
		dirty = true
	}
	if t.prices.Update() {
		dirty = true
	}
	if t.volume.Update() {
		dirty = true
	}
	if t.loadingTextBox.Update() {
		dirty = true
	}
	if t.errorTextBox.Update() {
		dirty = true
	}
	if t.fadeIn.Update() {
		dirty = true
	}
	return dirty
}

// Render renders the Thumb.
func (t *Thumb) Render(fudge float32) {
	t.frameBubble.Render(fudge)
	t.header.Render(fudge)
	rect.RenderLineAtTop(t.bodyBounds)

	// Only show messages if no prior data to show.
	if !t.hasStockUpdated {
		if t.loading {
			t.loadingTextBox.Render(fudge)
			return
		}

		if t.hasError {
			t.errorTextBox.Render(fudge)
			return
		}
	}

	old := gfx.Alpha()
	gfx.SetAlpha(old * t.fadeIn.Value(fudge))
	defer gfx.SetAlpha(old)

	// Render the dividers between the sections.
	for _, r := range t.sectionDividers {
		rect.RenderLineAtTop(r)
	}

	t.priceTimeline.Render(fudge)
	t.volumeTimeline.Render(fudge)

	t.prices.Render(fudge)
	t.movingAverage21.Render(fudge)
	t.movingAverage50.Render(fudge)
	t.movingAverage200.Render(fudge)
	t.volume.Render(fudge)

	t.priceCursor.Render(fudge)
	t.volumeCursor.Render(fudge)
}

// SetRemoveButtonClickCallback sets the callback for remove button clicks.
func (t *Thumb) SetRemoveButtonClickCallback(cb func()) {
	t.header.SetRemoveButtonClickCallback(cb)
}

// SetThumbClickCallback sets the callback for thumbnail clicks.
func (t *Thumb) SetThumbClickCallback(cb func()) {
	t.thumbClickCallback = cb
}

// Close frees the resources backing the chart thumbnail.
func (t *Thumb) Close() {
	t.header.Close()
	t.prices.Close()
	t.priceCursor.Close()
	t.priceTimeline.Close()
	t.movingAverage21.Close()
	t.movingAverage50.Close()
	t.movingAverage200.Close()
	t.volume.Close()
	t.volumeCursor.Close()
	t.volumeTimeline.Close()
	t.thumbClickCallback = nil
}
