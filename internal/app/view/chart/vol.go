package chart

import (
	"fmt"
	"image"
	"strconv"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

var volumeHorizRuleSet = vao.HorizRuleSet([]float32{0.2, 0.8}, [2]float32{0, 1}, color.Gray)

// volume renders the volume bars and labels for a single stock.
type volume struct {
	// renderable is whether the ChartVolume can be rendered.
	renderable bool

	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// volBars is the VAO with the colored volume bars.
	volBars *gfx.VAO

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func (v *volume) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	v.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	// Find the maximum volume.
	v.maxVolume = 0
	for _, s := range ts.TradingSessions {
		if v.maxVolume < s.Volume {
			v.maxVolume = s.Volume
		}
	}

	// Measure the max label size by creating a label with the max value.
	v.MaxLabelSize = makeVolumeLabel(v.maxVolume, 1).size

	v.volBars = volumeBarsVAO(ts.TradingSessions, v.maxVolume)

	v.renderable = true
}

func (v *volume) ProcessInput(bounds image.Rectangle) {
	v.bounds = bounds
}

func (v *volume) Render(fudge float32) {
	if !v.renderable {
		return
	}

	gfx.SetModelMatrixRect(v.bounds)

	// Render lines for the 20% and 80% levels.
	volumeHorizRuleSet.Render()

	// Render the volume bars.
	v.volBars.Render()
}

func (v *volume) Close() {
	v.renderable = false
	if v.volBars != nil {
		v.volBars.Delete()
	}
}

// volumeLabel is a right-justified Y-axis label with the volume.
type volumeLabel struct {
	percent float32
	text    string
	size    image.Point
}

func makeVolumeLabel(maxVolume int, perc float32) volumeLabel {
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

	return volumeLabel{
		percent: perc,
		text:    t,
		size:    chartAxisLabelTextRenderer.Measure(t),
	}
}

func volumeBarsVAO(ds []*model.TradingSession, maxVolume int) *gfx.VAO {
	data := &gfx.VAOVertexData{Mode: gfx.Triangles}

	dx := 2.0 / float32(len(ds)) // (-1 to 1) on X-axis
	calcX := func(i int) (leftX, rightX float32) {
		x := -1.0 + dx*float32(i)
		return x + dx*0.2, x + dx*0.8
	}
	calcY := func(v int) (topY, botY float32) {
		return 2*float32(v)/float32(maxVolume) - 1, -1
	}

	for i, s := range ds {
		leftX, rightX := calcX(i)
		topY, botY := calcY(s.Volume)

		// Add the vertices needed to create the volume bar.
		data.Vertices = append(data.Vertices,
			leftX, topY, 0, // UL
			rightX, topY, 0, // UR
			leftX, botY, 0, // BL
			rightX, botY, 0, // BR
		)

		// Add the colors corresponding to the volume bar.
		switch {
		case s.Close > s.Open:
			data.Colors = append(data.Colors,
				color.Green[0], color.Green[1], color.Green[2],
				color.Green[0], color.Green[1], color.Green[2],
				color.Green[0], color.Green[1], color.Green[2],
				color.Green[0], color.Green[1], color.Green[2],
			)

		case s.Close < s.Open:
			data.Colors = append(data.Colors,
				color.Red[0], color.Red[1], color.Red[2],
				color.Red[0], color.Red[1], color.Red[2],
				color.Red[0], color.Red[1], color.Red[2],
				color.Red[0], color.Red[1], color.Red[2],
			)

		default:
			data.Colors = append(data.Colors,
				color.Yellow[0], color.Yellow[1], color.Yellow[2],
				color.Yellow[0], color.Yellow[1], color.Yellow[2],
				color.Yellow[0], color.Yellow[1], color.Yellow[2],
				color.Yellow[0], color.Yellow[1], color.Yellow[2],
			)
		}

		// idx is function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(i)*4 + j
		}

		// Use triangles for filled candlestick on lower closes.
		data.Indices = append(data.Indices,
			idx(0), idx(2), idx(1),
			idx(1), idx(2), idx(3),
		)
	}

	return gfx.NewVAO(data)
}
