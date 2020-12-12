package chart

import (
	"fmt"
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

// priceLegend is a bubble that shows a trading session's stats where the mouse cursor is.
type priceLegend struct {
	// data is the data necessary to render.
	data priceLegendData
	// bounds is the bounds to draw the priceLegend within.
	bounds image.Rectangle
	// mousePos is the current mouse position. Nil for no mouse input.
	mousePos *view.MousePosition

	// table renders the information of a single trading session.
	table legendTable

	// needUpdate is true if the table and tableBubble need updating.
	needUpdate bool
	// renderable is true if there is something to render.
	renderable bool
}

func newPriceLegend() *priceLegend {
	return &priceLegend{}
}

type priceLegendData struct {
	Interval               model.Interval
	TradingSessionSeries   *model.TradingSessionSeries
	MovingAverageSeriesSet []*model.AverageSeries
}

func (p *priceLegend) SetData(data priceLegendData) {
	p.data = data
	p.needUpdate = true
}

func (p *priceLegend) SetBounds(bounds image.Rectangle) {
	if p.bounds == bounds {
		return
	}
	p.bounds = bounds
	p.needUpdate = true
}

func (p *priceLegend) ProcessInput(input *view.Input) {
	if p.mousePos == input.MousePos {
		return
	}
	p.mousePos = input.MousePos
	p.needUpdate = true
}

func (p *priceLegend) Update() (dirty bool) {
	if !p.needUpdate {
		return false
	}

	defer func() { p.needUpdate = false }()

	if p.data.TradingSessionSeries == nil {
		p.renderable = false
		return true
	}

	tss := p.data.TradingSessionSeries.TradingSessions
	if len(tss) == 0 {
		p.renderable = false
		return true
	}

	i, ts := len(tss)-1, tss[len(tss)-1]
	if p.mousePos.WithinX(p.bounds) {
		i, ts = tradingSessionAtX(tss, p.bounds, p.mousePos.X)
	}

	curr := ts
	prev := curr
	if i > 0 {
		prev = tss[i-1]
	}

	var empty legendCell

	rows := [][3]legendCell{
		{
			legendText(formatWeekday(curr.Date.Weekday())),
			legendText(curr.Date.Format("1/2/06")),
			empty,
		},
		{empty, empty, empty},
		{
			whiteArrow(curr.Open - prev.Open),
			legendText("Open"),
			legendText(formatFloat(curr.Open)),
		},
		{
			whiteArrow(curr.High - prev.High),
			legendText("High"),
			legendText(formatFloat(curr.High)),
		},
		{
			whiteArrow(curr.Low - prev.Low),
			legendText("Low"),
			legendText(formatFloat(curr.Low)),
		},
		{
			whiteArrow(curr.Close - prev.Close),
			legendText("Close"),
			legendText(formatFloat(curr.Close)),
		},
		{empty, empty, empty},
		{
			colorArrow(curr.Change),
			legendText("Change"),
			legendText(formatChange(curr.Change)),
		},
		{empty, empty, legendText(formatPercentChange(curr.PercentChange))},
	}

	if len(p.data.MovingAverageSeriesSet) != 0 {
		rows = append(rows, [3]legendCell{empty, empty, empty})
	}

	for _, series := range p.data.MovingAverageSeriesSet {
		value := series.Values[i].Value
		rows = append(rows, [3]legendCell{
			symbol(symbolLabel(curr.Close, value), movingAverageColors[p.data.Interval][series.Intervals]),
			legendText(fmt.Sprintf("%s %d", typeLabel(series.Type), series.Intervals)),
			legendText(formatFloat(value)),
		})
	}

	columns := [3]legendColumn{}
	for i := range rows {
		for j := range columns {
			if w := rows[i][j].size.X; w > columns[j].maxWidth {
				columns[j].maxWidth = w
			}
		}
	}

	tableBounds := image.Rect(
		0,
		0,
		legendTablePadding+columns[0].maxWidth+
			legendTablePadding+columns[1].maxWidth+
			legendTablePadding+columns[2].maxWidth+legendTablePadding,
		legendTablePadding+len(rows)*legendTextRenderer.LineHeight()+legendTablePadding,
	)

	// Move the table to the lower left.
	bounds := p.bounds.Inset(legendBubbleMargin)
	tableBounds = tableBounds.Add(bounds.Min)

	// Move the table to the right if the mouse is in the bounds.
	if p.mousePos.In(tableBounds) {
		tableBounds = tableBounds.Add(image.Pt(bounds.Dx()-tableBounds.Dx(), 0))
	}

	p.table = legendTable{
		bubble:  rect.NewBubble(legendBubbleRounding),
		rows:    rows,
		columns: columns,
	}
	p.table.SetBounds(tableBounds)
	p.renderable = true

	return true
}

func (p *priceLegend) Render(fudge float32) {
	if !p.renderable {
		return
	}

	p.table.Render(fudge)
}

func (p *priceLegend) Close() {
	p.renderable = false
}
