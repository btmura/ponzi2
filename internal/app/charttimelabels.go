package app

import (
	"image"
)

// ChartTimeLabels renders the time labels for a single stock.
type ChartTimeLabels struct {
	// renderable is whether the ChartTimeLabels can be rendered.
	renderable bool

	labels []chartTimeLabel
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

	ch.labels = chartTimeLabels(st.DailySessions)
	ch.renderable = true
}

// Render renders the ChartTimeLabels.
func (ch *ChartTimeLabels) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	for _, l := range ch.labels {
		x := r.Min.X + int(float32(r.Dx())*l.percent) - l.size.X/2
		y := r.Max.Y - l.size.Y
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
	}
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

func chartTimeLabels(ds []*ModelTradingSession) []chartTimeLabel {
	var ls []chartTimeLabel

	for i, s := range ds {
		if i == 0 {
			continue // Can't check previous week.
		}

		pm := ds[i-1].Date.Month()
		m := s.Date.Month()
		if pm == m {
			continue
		}

		mt := string(m.String()[0])

		ls = append(ls, chartTimeLabel{
			percent: float32(i) / float32(len(ds)),
			text:    mt,
			size:    chartAxisLabelTextRenderer.Measure(mt),
		})
	}

	return ls
}
