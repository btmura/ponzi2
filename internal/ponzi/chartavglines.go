package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartAvgLines renders a moving average line for a single stock.
type ChartAvgLines struct {
	renderable bool
	lines      *gfx.VAO
}

// NewChartAvgLines creates a new ChartAvgLines.
func NewChartAvgLines() *ChartAvgLines {
	return &ChartAvgLines{}
}

// SetStock sets the ChartAvgLine's stock.
func (ch *ChartAvgLines) SetStock(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	ch.lines = createChartAvgLinesVAO(st.DailySessions)
	ch.renderable = true
}

func createChartAvgLinesVAO(ds []*ModelTradingSession) *gfx.VAO {
	var minPrice float32
	var maxPrice float32
	for _, s := range ds {
		if minPrice > s.Low {
			minPrice = s.Low
		}
		if maxPrice < s.High {
			maxPrice = s.High
		}
	}

	data := &gfx.VAOVertexData{}
	var v uint16 // vertex index

	dx := 2.0 / float32(len(ds)) // (-1 to 1) on X-axis
	calcX := func(i int) float32 {
		return -1.0 + dx*float32(i) + dx*0.5
	}
	calcY := func(v float32) float32 {
		return 2.0*(v-minPrice)/(maxPrice-minPrice) - 1.0
	}

	first := true
	for i, s := range ds {
		if s.MovingAverage25 == 0 {
			continue
		}
		data.Vertices = append(data.Vertices, calcX(i), calcY(s.MovingAverage25), 0)
		data.Colors = append(data.Colors, purple[0], purple[1], purple[2])
		if !first {
			data.Indices = append(data.Indices, v, v-1)
		}
		v++
		first = false
	}

	first = true
	for i, s := range ds {
		if s.MovingAverage50 == 0 {
			continue
		}

		data.Vertices = append(data.Vertices, calcX(i), calcY(s.MovingAverage50), 0)
		data.Colors = append(data.Colors, yellow[0], yellow[1], yellow[2])
		if !first {
			data.Indices = append(data.Indices, v, v-1)
		}
		v++
		first = false
	}

	first = true
	for i, s := range ds {
		if s.MovingAverage200 == 0 {
			continue
		}

		data.Vertices = append(data.Vertices, calcX(i), calcY(s.MovingAverage200), 0)
		data.Colors = append(data.Colors, white[0], white[1], white[2])
		if !first {
			data.Indices = append(data.Indices, v, v-1)
		}
		v++
		first = false
	}

	return gfx.NewVAO(gfx.Lines, data)
}

// Render renders the ChartAvgLine.
func (ch *ChartAvgLines) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(r)
	ch.lines.Render()
}

// Close frees the resources backing the ChartAvgLine.
func (ch *ChartAvgLines) Close() {
	ch.renderable = false
	if ch.lines != nil {
		ch.lines.Delete()
		ch.lines = nil
	}
}
