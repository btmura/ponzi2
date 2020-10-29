package chart

import (
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

	price         *price
	priceCursor   *priceCursor
	priceTimeline *timeline

	movingAverages []*movingAverage

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

		price:         newPrice(priceStyle),
		priceCursor:   new(priceCursor),
		priceTimeline: newTimeline(view.TransparentLightGray, view.LightGray, view.TransparentGray, view.Gray),

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

	t.price.SetStyle(newPriceStyle)
	t.volume.SetStyle(newPriceStyle)
}

// SetLoading toggles the Chart's loading indicator.
func (t *Thumb) SetLoading(loading bool) {
	t.loading = loading
	t.header.SetLoading(loading)
}

// SetErrorMessage sets or resets an error message on charts and thumbnails that match the symbol.
// An empty error message clears any previously set error messages.
func (t *Thumb) SetErrorMessage(errorMessage string) {
	t.hasError = errorMessage != ""
	if errorMessage != "" {
		t.errorTextBox.SetText(errorMessage)
	}
	t.header.SetErrorMessage(errorMessage)
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

	var mas []*model.MovingAverageSeries
	for _, ma := range dc.MovingAverageSeriesSet {
		mas = append(mas, ma)
	}

	vs := dc.AverageVolumeSeries

	if ts == nil || vs == nil {
		return
	}

	tl := len(ts.TradingSessions)
	if vl := len(vs.Values); tl != vl {
		logger.Errorf("volume has different lengths: %d vs %d", tl, vl)
		return
	}

	for _, ma := range mas {
		if ml := len(ma.Values); ml != tl {
			logger.Errorf("moving average has different length: %d vs %d", tl, ml)
			return
		}
	}

	const days = 20
	if l := len(ts.TradingSessions); l > days {
		ts = ts.DeepCopy()
		ts.TradingSessions = ts.TradingSessions[l-days:]
	}
	for i, m := range mas {
		if l := len(m.Values); l > days {
			m = m.DeepCopy()
			m.Values = m.Values[l-days:]
			mas[i] = m
		}
	}
	if l := len(vs.Values); l > days {
		vs = vs.DeepCopy()
		vs.Values = vs.Values[l-days:]
	}

	t.price.SetData(priceData{ts})
	t.priceCursor.SetData(priceCursorData{ts})
	t.priceTimeline.SetData(timelineData{dc.Interval, ts})

	for _, ma := range t.movingAverages {
		ma.Close()
	}

	t.movingAverages = nil
	for _, ma := range mas {
		m := newMovingAverage(movingAverageColors[dc.Interval][ma.Intervals])
		m.SetData(movingAverageData{ts, ma})
		t.movingAverages = append(t.movingAverages, m)
	}

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

	t.price.SetBounds(pr)
	t.priceCursor.SetBounds(pr, pr)
	t.priceTimeline.SetBounds(pr)

	for _, ma := range t.movingAverages {
		ma.SetBounds(pr)
	}

	t.volume.SetBounds(vr)
	t.volumeCursor.SetBounds(vr, vr)
	t.volumeTimeline.SetBounds(vr)

	t.priceCursor.ProcessInput(input)
	t.volumeCursor.ProcessInput(input)
}

// Update updates the Thumb.
func (t *Thumb) Update() (dirty bool) {
	if t.header.Update() {
		dirty = true
	}
	if t.price.Update() {
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
	t.price.Render(fudge)
	for _, ma := range t.movingAverages {
		ma.Render(fudge)
	}
	t.priceCursor.Render(fudge)

	t.volumeTimeline.Render(fudge)
	t.volume.Render(fudge)
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
	t.price.Close()
	t.priceCursor.Close()
	t.priceTimeline.Close()
	for _, ma := range t.movingAverages {
		ma.Close()
	}
	t.volume.Close()
	t.volumeCursor.Close()
	t.volumeTimeline.Close()
	t.thumbClickCallback = nil
}
