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

var volumeHorizRuleSet = vao.HorizRuleSet([]float32{0.2, 0.8}, [2]float32{0, 1}, color.TransparentGray, color.Gray)

// volume renders the volume bars and labels for a single stock.
type volume struct {
	// renderable is whether the ChartVolume can be rendered.
	renderable bool

	volumeRange [2]int

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// bars is the VAO with the colored volume bars.
	bars *gfx.VAO

	// avgLine is the VAO with the average volume line.
	avgLine *gfx.VAO

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

type volumeData struct {
	TradingSessionSeries *model.TradingSessionSeries
	AverageVolumeSeries  *model.AverageVolumeSeries
}

func (v *volume) SetData(data volumeData) {
	// Reset everything.
	v.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
	vs := data.AverageVolumeSeries
	if ts == nil || vs == nil {
		return
	}

	v.volumeRange = volumeRange(ts.TradingSessions)

	// Measure the max label size by creating a label with the max value.
	v.MaxLabelSize = makeVolumeLabel(v.volumeRange[1], 1).size

	v.bars = volumeBarsVAO(ts.TradingSessions, v.volumeRange[1])

	var values []float32
	for _, m := range vs.AverageVolumes {
		values = append(values, m.Value)
	}
	v.avgLine = vao.DataLine(values, [2]float32{float32(v.volumeRange[0]), float32(v.volumeRange[1])}, color.White)

	v.renderable = true
}

func (v *volume) SetBounds(bounds image.Rectangle) {
	v.bounds = bounds
}

func (v *volume) Render(float32) {
	if !v.renderable {
		return
	}

	gfx.SetModelMatrixRect(v.bounds)
	volumeHorizRuleSet.Render()
	v.bars.Render()
	v.avgLine.Render()
}

func (v *volume) Close() {
	v.renderable = false
	if v.bars != nil {
		v.bars.Delete()
	}
	if v.avgLine != nil {
		v.avgLine.Delete()
	}
}

func volumeRange(ts []*model.TradingSession) [2]int {
	var high int

	for _, s := range ts {
		if s.Volume > high {
			high = s.Volume
		}
	}

	return [2]int{0, high}
}

// volumeLabel is a right-justified Y-axis label with the volume.
type volumeLabel struct {
	percent float32
	text    string
	size    image.Point
}

func makeVolumeLabel(maxVolume int, perc float32) volumeLabel {
	t := volumeText(int(float32(maxVolume) * perc))
	return volumeLabel{
		percent: perc,
		text:    t,
		size:    axisLabelTextRenderer.Measure(t),
	}
}

func volumeText(v int) string {
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
	return t
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
		add := func(c [4]float32) {
			data.Colors = append(data.Colors,
				c[0], c[1], c[2], c[3],
				c[0], c[1], c[2], c[3],
				c[0], c[1], c[2], c[3],
				c[0], c[1], c[2], c[3],
			)
		}

		switch {
		case s.Close > s.Open:
			add(color.Green)

		case s.Close < s.Open:
			add(color.Red)

		default:
			add(color.White)
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
