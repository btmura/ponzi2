package chart

import (
	"image"
	"math"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/vao"
	"github.com/btmura/ponzi2/internal/logger"
)

// volume renders the volume bars and labels for a single stock.
type volume struct {
	// renderable is whether the ChartVolume can be rendered.
	renderable bool

	// priceStyle is the price style whether bars or candlesticks.
	priceStyle PriceStyle

	// faders has the faders needed to fade in and out the bars and candlesticks.
	faders map[PriceStyle]*view.Fader

	// barLines are the volume bars colored to go with price bars.
	barLines *gfx.VAO

	// stickLines are the volume bars colored to go with candlesticks.
	stickLines *gfx.VAO

	// avgLine is the VAO with the average volume line.
	avgLine *gfx.VAO

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func newVolume(priceStyle PriceStyle) *volume {
	if priceStyle == PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return nil
	}

	return &volume{
		priceStyle: priceStyle,
		faders: map[PriceStyle]*view.Fader{
			Bar:         view.NewStoppedFader(1 * view.FPS),
			Candlestick: view.NewStoppedFader(1 * view.FPS),
		},
	}
}

// SetPriceStyle sets the priceStyle whether bars or candlesticks.
func (v *volume) SetStyle(newStyle PriceStyle) {
	if newStyle == PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return
	}

	if newStyle == v.priceStyle {
		return
	}

	v.priceStyle = newStyle
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

	yRange := volumeRange(ts.TradingSessions)

	v.barLines = volumeLineVAO(ts.TradingSessions, yRange, Bar)
	v.stickLines = volumeLineVAO(ts.TradingSessions, yRange, Candlestick)
	v.avgLine = volumeDataLine(vs.Values, yRange)

	v.renderable = true
}

func (v *volume) SetBounds(bounds image.Rectangle) {
	v.bounds = bounds
}

func (v *volume) Update() (dirty bool) {
	for s, fader := range v.faders {
		if s == v.priceStyle {
			fader.FadeIn()
		} else {
			fader.FadeOut()
		}
	}

	for _, fader := range v.faders {
		if fader.Update() {
			dirty = true
		}
	}
	return dirty
}

func (v *volume) Render(fudge float32) {
	if !v.renderable {
		return
	}

	gfx.SetModelMatrixRect(v.bounds)

	for style, fader := range v.faders {
		fader.Render(fudge, func() {
			switch style {
			case Bar:
				v.barLines.Render()

			case Candlestick:
				v.stickLines.Render()
			}
		})
	}

	v.avgLine.Render()
}

func (v *volume) Close() {
	v.renderable = false
	if v.barLines != nil {
		v.barLines.Delete()
	}
	if v.stickLines != nil {
		v.stickLines.Delete()
	}
	if v.avgLine != nil {
		v.avgLine.Delete()
	}
}

func volumeRange(ts []*model.TradingSession) [2]int {
	if len(ts) == 0 {
		return [2]int{0, 0}
	}

	low, high := math.MaxInt64, 0
	for _, s := range ts {
		if s.Volume != 0 && s.Volume < low {
			low = s.Volume
		}
		if s.Volume != 0 && s.Volume > high {
			high = s.Volume
		}
	}

	if low > high {
		return [2]int{0, 0}
	}

	return [2]int{low, high}
}

func volumePercent(volumeRange [2]int, value int) (percent float32) {
	log := func(value int) float64 {
		if value == 0 {
			return 0
		}
		return math.Log(float64(value))
	}
	percent = float32((log(value) - log(volumeRange[0])) / (log(volumeRange[1]) - log(volumeRange[0])))
	if percent >= 0 {
		return percent
	}
	return 0
}

func volumeValue(volumeRange [2]int, percent float32) (value int) {
	log := func(value int) float64 {
		if value == 0 {
			return 0
		}
		return math.Log(float64(value))
	}
	return int(
		math.Pow(
			math.E,
			float64(percent)*(log(volumeRange[1])-log(volumeRange[0]))+log(volumeRange[0]),
		),
	)
}

func volumeLineVAO(ts []*model.TradingSession, volumeRange [2]int, priceStyle PriceStyle) *gfx.VAO {
	var vertices []float32
	var colors []float32
	var lineIndices []uint16

	dx := 2.0 / float32(len(ts)) // (-1 to 1) on X-axis
	calcX := func(i int) (centerX float32) {
		x := -1.0 + dx*float32(i)
		return x + dx*.5
	}
	calcY := func(value int) (topY, botY float32) {
		return 2*volumePercent(volumeRange, value) - 1, -1
	}

	for i, s := range ts {
		centerX := calcX(i)
		topY, botY := calcY(s.Volume)

		// Add the vertices needed to create the volume bar.
		idxOffset := len(vertices) / 3
		vertices = append(vertices,
			centerX, topY, 0, // 0
			centerX, botY, 0, // 1
		)

		// Add the colors corresponding to the vertices.
		c := view.White
		switch priceStyle {
		case Bar:
			switch {
			case s.Source == model.RealTimePrice:
				c = view.Yellow
			case s.Change > 0:
				c = view.Blue
			case s.Change < 0:
				c = view.Red
			}

		case Candlestick:
			switch {
			case s.Source == model.RealTimePrice:
				c = view.Yellow
			case s.Close > s.Open:
				c = view.Blue
			case s.Close < s.Open:
				c = view.Red
			}
		}

		colors = append(colors,
			c[0], c[1], c[2], c[3], // 0
			c[0], c[1], c[2], c[3], // 1
		)

		// idx is function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(idxOffset) + j
		}

		// Add the vertex indices to render the bars.
		lineIndices = append(lineIndices,
			idx(0), idx(1),
		)
	}

	return gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode:     gfx.Lines,
			Vertices: vertices,
			Colors:   colors,
			Indices:  lineIndices,
		},
	)
}

func volumeDataLine(vs []*model.AverageVolumeValue, yRange [2]int) *gfx.VAO {
	var yPercentValues []float32
	for _, v := range vs {
		yPercentValues = append(yPercentValues, volumePercent(yRange, int(v.Value)))
	}
	return vao.DataLine(yPercentValues, view.Red)
}
