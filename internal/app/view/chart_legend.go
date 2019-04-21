package view

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

	"github.com/btmura/ponzi2/internal/app/model"
)

type chartLegend struct {
	tradingSessionSeries   *model.TradingSessionSeries
	movingAverageSeries25  *model.MovingAverageSeries
	movingAverageSeries50  *model.MovingAverageSeries
	movingAverageSeries200 *model.MovingAverageSeries
	showMovingAverages     bool

	priceRect image.Rectangle
	labelRect image.Rectangle
	mousePos  image.Point
}

func (ch *chartLegend) SetData(
	ts *model.TradingSessionSeries,
	ms25 *model.MovingAverageSeries,
	ms50 *model.MovingAverageSeries,
	ms200 *model.MovingAverageSeries,
	showMovingAverages bool) {

	ch.tradingSessionSeries = ts
	ch.movingAverageSeries25 = ms25
	ch.movingAverageSeries50 = ms50
	ch.movingAverageSeries200 = ms200
	ch.showMovingAverages = showMovingAverages
}

func (ch *chartLegend) ProcessInput(priceRect, labelRect image.Rectangle, mousePos image.Point) {
	ch.priceRect = priceRect
	ch.labelRect = labelRect
	ch.mousePos = mousePos
}

func (ch *chartLegend) Render(fudge float32) {
	// Renders trackline legend. To display inline comment out the two lines below and uncomment the last line.
	if ch.showMovingAverages {
		ch.renderMATrackLineLegend()
	}
	ch.renderCandleTrackLineLegend()
}

func (ch *chartLegend) renderMATrackLineLegend() {
	if !ch.mousePos.In(ch.priceRect) {
		return
	}

	// Render moving average trackline legend
	ts := ch.tradingSessionSeries.TradingSessions
	ms25 := ch.movingAverageSeries25.MovingAverages
	ms50 := ch.movingAverageSeries50.MovingAverages
	ms200 := ch.movingAverageSeries200.MovingAverages

	mal := chartLegendLabel{
		percent: float32(ch.mousePos.X-ch.priceRect.Min.X) / float32(ch.priceRect.Dx()),
	}

	i := int(math.Floor(float64(len(ts)) * float64(mal.percent)))
	if i >= len(ts) {
		i = len(ts) - 1
	}

	mal.text = ch.chartLegendMALabelText(ms25[i], ms50[i], ms200[i])
	mal.size = chartAxisLabelTextRenderer.Measure(mal.text)

	MAp := image.Point{
		X: ch.labelRect.Min.X + ch.labelRect.Dx()/2 - mal.size.X/2,
		Y: ch.labelRect.Min.Y + ch.labelRect.Dy() - int(math.Floor(float64(mal.size.Y)*float64(2.4))),
	}

	renderBubble(MAp, mal.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(mal.text, MAp, white)
}

func (ch *chartLegend) renderCandleTrackLineLegend() {
	if !ch.mousePos.In(ch.priceRect) {
		return
	}

	// Render candle trackline legend
	ts := ch.tradingSessionSeries.TradingSessions

	pl := chartLegendLabel{
		percent: float32(ch.mousePos.X-ch.priceRect.Min.X) / float32(ch.priceRect.Dx()),
	}

	i := int(math.Floor(float64(len(ts)) * float64(pl.percent)))
	if i >= len(ts) {
		i = len(ts) - 1
	}

	// Candlestick and moving average legends stacked as two chartAxisLabelBubbleSpec lines
	pl.text = ch.chartLegendCandleLabelText(ts[i])
	pl.size = chartAxisLabelTextRenderer.Measure(pl.text)

	ohlcvp := image.Point{
		X: ch.labelRect.Min.X + ch.labelRect.Dx()/2 - pl.size.X/2,
		Y: ch.labelRect.Min.Y + ch.labelRect.Dy() - pl.size.Y,
	}

	renderBubble(ohlcvp, pl.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(pl.text, ohlcvp, white)
}

type chartLegendLabel struct {
	percent float32
	text    string
	size    image.Point
}

func (ch *chartLegend) chartLegendMALabelText(ma25 *model.MovingAverage, ma50 *model.MovingAverage, ma200 *model.MovingAverage) string {
	legendMA := strings.Join([]string{
		"MA25:", strconv.FormatFloat(float64(ma25.Value), 'f', 1, 32),
		"   MA50:", strconv.FormatFloat(float64(ma50.Value), 'f', 1, 32),
		"   MA200:", strconv.FormatFloat(float64(ma200.Value), 'f', 1, 32),
	}, "")
	return fmt.Sprintf("%s", legendMA)
}

func (ch *chartLegend) chartLegendCandleLabelText(candle *model.TradingSession) string {
	legendOHLCVC := string(strings.Join([]string{
		"O:", strconv.FormatFloat(float64(candle.Open), 'f', 2, 32),
		"   H:", strconv.FormatFloat(float64(candle.High), 'f', 2, 32),
		"   L:", strconv.FormatFloat(float64(candle.Low), 'f', 2, 32),
		"   C:", strconv.FormatFloat(float64(candle.Close), 'f', 2, 32),
		"   V:", strconv.Itoa(candle.Volume), "   ",
		strconv.FormatFloat(float64(candle.Change), 'f', 2, 32), "%"}, ""))
	return fmt.Sprintf("%s", legendOHLCVC)
}
