package chart

import (
	"fmt"
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

// volumeLegend is a bubble that shows a trading session's stats where the mouse cursor is.
type volumeLegend struct {
	// data is the data necessary to render.
	data volumeLegendData
	// bounds is the bounds to draw the volumeLegend within.
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

func newVolumeLegend() *volumeLegend {
	return &volumeLegend{}
}

type volumeLegendData struct {
	Interval             model.Interval
	TradingSessionSeries *model.TradingSessionSeries
	AverageVolumeSeries  *model.AverageSeries
}

func (v *volumeLegend) SetData(data volumeLegendData) {
	v.data = data
	v.needUpdate = true
}

func (v *volumeLegend) SetBounds(bounds image.Rectangle) {
	if v.bounds == bounds {
		return
	}
	v.bounds = bounds
	v.needUpdate = true
}

func (v *volumeLegend) ProcessInput(input *view.Input) {
	if v.mousePos == input.MousePos {
		return
	}
	v.mousePos = input.MousePos
	v.needUpdate = true
}

func (v *volumeLegend) Update() (dirty bool) {
	if !v.needUpdate {
		return false
	}

	defer func() { v.needUpdate = false }()

	if v.data.TradingSessionSeries == nil {
		v.renderable = false
		return true
	}

	tss := v.data.TradingSessionSeries.TradingSessions
	if len(tss) == 0 {
		v.renderable = false
		return true
	}

	i, ts := len(tss)-1, tss[len(tss)-1]
	if v.mousePos.WithinX(v.bounds) {
		i, ts = tradingSessionAtX(tss, v.bounds, v.mousePos.X)
	}

	curr := ts
	prev := curr
	if i > 0 {
		prev = tss[i-1]
	}

	if curr.Volume == 0 {
		v.renderable = false
		return true
	}

	var empty legendCell
	var rows [][3]legendCell

	dv := curr.Volume - prev.Volume
	rows = append(rows,
		[3]legendCell{
			whiteArrow(float32(dv)),
			legendText("Volume"),
			legendText(volumeText(curr.Volume)),
		},
		[3]legendCell{empty, empty, empty},
		[3]legendCell{
			colorArrow(float32(dv)),
			legendText("Change"),
			legendText(volumeChangeText(dv)),
		},
		[3]legendCell{
			empty,
			empty,
			legendText(formatPercentChange(curr.VolumePercentChange)),
		},
	)

	if series := v.data.AverageVolumeSeries; len(series.Values) == len(tss) {
		value := series.Values[i].Value
		rows = append(rows,
			[3]legendCell{empty, empty, empty},
			[3]legendCell{
				symbol(symbolLabel(float32(curr.Volume), value), view.Red),
				legendText(fmt.Sprintf("%s %d", typeLabel(series.Type), series.Intervals)),
				legendText(volumeText(int(value))),
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

	// Move the table to the upper left.
	bounds := v.bounds.Inset(legendBubbleMargin)
	tableBounds = tableBounds.Add(image.Pt(
		bounds.Min.X,
		bounds.Max.Y-tableBounds.Dy(),
	))

	// Move the table to the right if the mouse is in the bounds.
	if v.mousePos.In(tableBounds) {
		tableBounds = tableBounds.Add(image.Pt(bounds.Dx()-tableBounds.Dx(), 0))
	}

	v.table = legendTable{
		bubble:  rect.NewBubble(legendBubbleRounding),
		rows:    rows,
		columns: columns,
	}
	v.table.SetBounds(tableBounds)
	v.renderable = true

	return true
}

func (v *volumeLegend) Render(fudge float32) {
	if !v.renderable {
		return
	}

	v.table.Render(fudge)
}

func (v *volumeLegend) Close() {
	v.renderable = false
}
