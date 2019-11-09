package chart

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
)

type legend struct {
	tradingSessionSeries   *model.TradingSessionSeries
	movingAverageSeries25  *model.MovingAverageSeries
	movingAverageSeries50  *model.MovingAverageSeries
	movingAverageSeries200 *model.MovingAverageSeries
	showMovingAverages     bool

	priceRect image.Rectangle
	labelRect image.Rectangle
	mousePos  image.Point
}

func (l *legend) SetData(
	ts *model.TradingSessionSeries,
	ms25 *model.MovingAverageSeries,
	ms50 *model.MovingAverageSeries,
	ms200 *model.MovingAverageSeries,
	showMovingAverages bool) {

	l.tradingSessionSeries = ts
	l.movingAverageSeries25 = ms25
	l.movingAverageSeries50 = ms50
	l.movingAverageSeries200 = ms200
	l.showMovingAverages = showMovingAverages
}

func (l *legend) SetBounds(priceRect, labelRect image.Rectangle) {
	l.priceRect = priceRect
	l.labelRect = labelRect
}

func (l *legend) ProcessInput(mousePos image.Point) {
	l.mousePos = mousePos
}

func (l *legend) Render(fudge float32) {
	// Renders trackline legend. To display inline comment out the two lines below and uncomment the last line.
	if l.showMovingAverages {
		l.renderMATrackLineLegend(fudge)
	}
	l.renderCandleTrackLineLegend(fudge)
}

func (l *legend) renderMATrackLineLegend(fudge float32) {
	if !l.mousePos.In(l.priceRect) {
		return
	}

	// Render moving average trackline legend
	ts := l.tradingSessionSeries.TradingSessions
	ms25 := l.movingAverageSeries25.MovingAverages
	ms50 := l.movingAverageSeries50.MovingAverages
	ms200 := l.movingAverageSeries200.MovingAverages

	mal := legendLabel{
		percent: float32(l.mousePos.X-l.priceRect.Min.X) / float32(l.priceRect.Dx()),
	}

	i := int(math.Floor(float64(len(ts)) * float64(mal.percent)))
	if i >= len(ts) {
		i = len(ts) - 1
	}

	mal.text = l.chartLegendMALabelText(ms25[i], ms50[i], ms200[i])
	mal.size = axisLabelTextRenderer.Measure(mal.text)

	textPt := image.Point{
		X: l.labelRect.Min.X + l.labelRect.Dx()/2 - mal.size.X/2,
		Y: l.labelRect.Min.Y + l.labelRect.Dy() - int(math.Floor(float64(mal.size.Y)*float64(2.4))),
	}
	bubbleRect := image.Rectangle{
		Min: textPt,
		Max: textPt.Add(mal.size),
	}
	bubbleRect = bubbleRect.Inset(-axisLabelPadding)

	axisLabelBubble.SetBounds(bubbleRect)
	axisLabelBubble.Render(fudge)
	axisLabelTextRenderer.Render(mal.text, textPt, gfx.TextColor(color.White))
}

func (l *legend) renderCandleTrackLineLegend(fudge float32) {
	if !l.mousePos.In(l.priceRect) {
		return
	}

	// Render candle trackline legend
	ts := l.tradingSessionSeries.TradingSessions

	pl := legendLabel{
		percent: float32(l.mousePos.X-l.priceRect.Min.X) / float32(l.priceRect.Dx()),
	}

	i := int(math.Floor(float64(len(ts)) * float64(pl.percent)))
	if i >= len(ts) {
		i = len(ts) - 1
	}

	// Candlestick and moving average legends stacked as two chartAxisLabelBubbleSpec lines
	pl.text = l.chartLegendCandleLabelText(ts[i])
	pl.size = axisLabelTextRenderer.Measure(pl.text)

	textPt := image.Point{
		X: l.labelRect.Min.X + l.labelRect.Dx()/2 - pl.size.X/2,
		Y: l.labelRect.Min.Y + l.labelRect.Dy() - pl.size.Y,
	}
	bubbleRect := image.Rectangle{
		Min: textPt,
		Max: textPt.Add(pl.size),
	}
	bubbleRect = bubbleRect.Inset(-axisLabelPadding)

	axisLabelBubble.SetBounds(bubbleRect)
	axisLabelBubble.Render(fudge)
	axisLabelTextRenderer.Render(pl.text, textPt, gfx.TextColor(color.White))
}

type legendLabel struct {
	percent float32
	text    string
	size    image.Point
}

func (l *legend) chartLegendMALabelText(ma25 *model.MovingAverage, ma50 *model.MovingAverage, ma200 *model.MovingAverage) string {
	legendMA := strings.Join([]string{
		"MA25:", strconv.FormatFloat(float64(ma25.Value), 'f', 1, 32),
		"   MA50:", strconv.FormatFloat(float64(ma50.Value), 'f', 1, 32),
		"   MA200:", strconv.FormatFloat(float64(ma200.Value), 'f', 1, 32),
	}, "")
	return fmt.Sprintf("%s", legendMA)
}

func (l *legend) chartLegendCandleLabelText(candle *model.TradingSession) string {
	legendOHLCVC := string(strings.Join([]string{
		"O:", strconv.FormatFloat(float64(candle.Open), 'f', 2, 32),
		"   H:", strconv.FormatFloat(float64(candle.High), 'f', 2, 32),
		"   L:", strconv.FormatFloat(float64(candle.Low), 'f', 2, 32),
		"   C:", strconv.FormatFloat(float64(candle.Close), 'f', 2, 32),
		"   V:", strconv.Itoa(candle.Volume), "   ",
		strconv.FormatFloat(float64(candle.Change), 'f', 2, 32), "%"}, ""))
	return fmt.Sprintf("%s", legendOHLCVC)
}
