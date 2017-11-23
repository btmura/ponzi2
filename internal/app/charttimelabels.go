package app

import (
	"image"
	"math"
	"time"
)

// ChartTimeLabels renders the time labels for a single stock.
type ChartTimeLabels struct {
	// renderable is whether the ChartTimeLabels can be rendered.
	renderable bool

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// labels bundle rendering measurements for time labels.
	labels []chartTimeLabel

	// dates are session dates shown for the cursor.
	dates []time.Time
}

// NewChartTimeLabels creates a new ChartTimeLabels.
func NewChartTimeLabels() *ChartTimeLabels {
	return &ChartTimeLabels{}
}

// SetStock sets the ChartTimeLabels' stock.
func (ch *ChartTimeLabels) SetStock(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	ch.MaxLabelSize = chartAxisLabelTextRenderer.Measure(chartTimeLabelText(time.December))

	ch.labels = chartTimeLabels(st.DailySessions)

	ch.dates = nil
	for _, s := range st.DailySessions {
		ch.dates = append(ch.dates, s.Date)
	}

	ch.renderable = true
}

// Render renders the ChartTimeLabels.
func (ch *ChartTimeLabels) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	for _, l := range ch.labels {
		tp := image.Point{
			X: r.Min.X + int(float32(r.Dx())*l.percent) - l.size.X/2,
			Y: r.Min.Y + r.Dy()/2 - l.size.Y/2,
		}
		chartAxisLabelTextRenderer.Render(l.text, tp, white)
	}
}

// RenderCursorLabels renders a label for the value under the mouse cursor.
func (ch *ChartTimeLabels) RenderCursorLabels(mainRect, labelRect image.Rectangle, mousePos image.Point) {
	if !ch.renderable {
		return
	}

	if mousePos.X < mainRect.Min.X || mousePos.X > mainRect.Max.X {
		return
	}

	l := chartTimeLabel{
		percent: float32(mousePos.X-mainRect.Min.X) / float32(mainRect.Dx()),
	}

	i := int(math.Floor(float64(len(ch.dates))*float64(l.percent) + 0.5))
	if i >= len(ch.dates) {
		i = len(ch.dates) - 1
	}
	l.text = ch.dates[i].Format("1/2/06")
	l.size = chartAxisLabelTextRenderer.Measure(l.text)

	tp := image.Point{
		X: mousePos.X - l.size.X/2,
		Y: labelRect.Min.Y + labelRect.Dy()/2 - l.size.Y/2,
	}

	renderBubble(tp, l.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)
}

// Close frees the resources backing the ChartTimeLabels.
func (ch *ChartTimeLabels) Close() {
	ch.renderable = false
}

type chartTimeLabel struct {
	percent float32
	text    string
	size    image.Point
}

func chartTimeLabelText(month time.Month) string {
	return string(month.String()[0:3])
}

func chartTimeLabels(ds []*ModelTradingSession) []chartTimeLabel {
	var ls []chartTimeLabel

	for i, s := range ds {
		if i == 0 {
			continue // Can't check previous month.
		}

		pm := ds[i-1].Date.Month()
		m := s.Date.Month()
		if pm == m {
			continue
		}

		txt := chartTimeLabelText(m)

		ls = append(ls, chartTimeLabel{
			percent: float32(i) / float32(len(ds)),
			text:    txt,
			size:    chartAxisLabelTextRenderer.Measure(txt),
		})
	}

	return ls
}
