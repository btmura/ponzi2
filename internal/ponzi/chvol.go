package ponzi

import (
	"fmt"
	"image"
	"strconv"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartVolume renders the volume bars and labels for a single stock.
type ChartVolume struct {
	renderable bool
	labels     []chartVolumeLabel
	volRects   *gfx.VAO
}

type chartVolumeLabel struct {
	percent float32
	text    string
	size    image.Point
}

// Update updates the ChartVolume with the stock.
func (ch *ChartVolume) Update(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	// Find the maximum volume.
	var maxVolume int
	for _, s := range st.DailySessions {
		if maxVolume < s.Volume {
			maxVolume = s.Volume
		}
	}

	// Create Y-axis labels for key percentages.
	makeLabel := func(perc float32) chartVolumeLabel {
		v := int(float32(maxVolume) * perc)

		var t string
		switch {
		case v > 1000000000:
			t = fmt.Sprintf("%dB", v/1000000000)
		case v > 1000000:
			t = fmt.Sprintf("%dM", v/1000000)
		case v > 1000:
			t = fmt.Sprintf("%dK", v/1000)
		default:
			t = strconv.Itoa(v)
		}

		return chartVolumeLabel{
			percent: perc,
			text:    t,
			size:    chartAxisLabelTextRenderer.Measure(t),
		}
	}

	ch.labels = append(ch.labels, makeLabel(.7))
	ch.labels = append(ch.labels, makeLabel(.3))

	ch.volRects = createChartVolumeBarsVAO(st.DailySessions, maxVolume)
	ch.renderable = true
}

func createChartVolumeBarsVAO(ds []*ModelTradingSession, maxVolume int) *gfx.VAO {
	var vertices []float32
	var colors []float32
	var indices []uint16

	barWidth := 2.0 / float32(len(ds)) // (-1 to 1) on X-axis
	leftX := -1.0 + barWidth*0.2
	rightX := -1.0 + barWidth*0.8

	calcY := func(value int) float32 {
		return 2*float32(value)/float32(maxVolume) - 1
	}

	for i, s := range ds {
		topY := calcY(s.Volume)
		botY := calcY(0)

		// Add the vertices needed to create the volume bar.
		vertices = append(vertices,
			leftX, topY, 0, // UL
			rightX, topY, 0, // UR
			leftX, botY, 0, // BL
			rightX, botY, 0, // BR
		)

		// Add the colors corresponding to the volume bar.
		switch {
		case s.Close > s.Open:
			colors = append(colors,
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
			)

		case s.Close < s.Open:
			colors = append(colors,
				red[0], red[1], red[2],
				red[0], red[1], red[2],
				red[0], red[1], red[2],
				red[0], red[1], red[2],
			)

		default:
			colors = append(colors,
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
			)
		}

		// idx is function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(i)*4 + j
		}

		// Use triangles for filled candlestick on lower closes.
		indices = append(indices,
			idx(0), idx(2), idx(1),
			idx(1), idx(2), idx(3),
		)

		// Move the X coordinates one bar over.
		leftX += barWidth
		rightX += barWidth
	}

	return gfx.NewVAO(
		gfx.Triangles,
		&gfx.VAOVertexData{
			Vertices: vertices,
			Colors:   colors,
			Indices:  indices,
		},
	)
}

// Render renders the volume bars.
func (ch *ChartVolume) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(r)
	ch.volRects.Render()

	sliceRenderHorizDividers(r, chartGridHorizLine, 0.3, 0.4)
}

// RenderLabels renders the Y-axis labels for the volume bars.
func (ch *ChartVolume) RenderLabels(r image.Rectangle) (maxLabelWidth int) {
	if !ch.renderable {
		return
	}

	var maxWidth int
	renderLabel := func(l chartVolumeLabel) {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
		if maxWidth < l.size.X {
			maxWidth = l.size.X
		}
	}
	for _, l := range ch.labels {
		renderLabel(l)
	}
	return maxWidth
}

// Close frees the resources backing the ChartVolume.
func (ch *ChartVolume) Close() {
	ch.renderable = false
	ch.labels = nil
	if ch.volRects != nil {
		ch.volRects.Delete()
		ch.volRects = nil
	}
}
