package chart

import (
	"fmt"
	"image"
	"math"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

const (
	legendBubbleMargin = 10
	legendTablePadding = 10
	legendFontSize     = 12
)

var (
	legendTextRenderer = gfx.NewTextRenderer(goregular.TTF, legendFontSize)

	// legendGeometricShapeRenderer renderers geometric shapes.
	// https://www.fileformat.info/info/unicode/font/dejavu_sans_mono/blockview.htm?block=geometric_shapes
	legendGeometricShapeRenderer = gfx.NewTextRenderer(_escFSMustByte(false, "/data/DejaVuSans.ttf"), legendFontSize)
)

// legend is a bubble that shows a trading session's stats where the mouse cursor is.
type legend struct {
	// data is the data necessary to render.
	data legendData
	// bounds is the bounds to draw the legend within.
	bounds image.Rectangle
	// mousePos is the current mouse position. Nil for no mouse input.
	mousePos *view.MousePosition

	// table has the rows and column updated  for rendering.
	table legendTable
	// tableBubble is the bubble to render the table within.
	tableBubble *rect.Bubble

	// needUpdate is true if the table and tableBubble need updating.
	needUpdate bool
	// renderable is true if there is something to render.
	renderable bool
}

type legendTable struct {
	rows    [][3]legendCell
	columns [3]legendColumn
}

type legendCell struct {
	renderer *gfx.TextRenderer
	text     string
	color    color.RGBA
	size     image.Point
}

func (l *legendCell) Render(pt image.Point) {
	l.renderer.Render(l.text, pt, gfx.TextColor(l.color))
}

type legendColumn struct {
	maxWidth int
}

func newLegend() *legend {
	return &legend{tableBubble: rect.NewBubble(10)}
}

type legendData struct {
	TradingSessionSeries   *model.TradingSessionSeries
	MovingAverageSeries25  *model.MovingAverageSeries
	MovingAverageSeries50  *model.MovingAverageSeries
	MovingAverageSeries200 *model.MovingAverageSeries
}

func (l *legend) SetData(data legendData) {
	l.data = data
	l.needUpdate = true
}

func (l *legend) SetBounds(bounds image.Rectangle) {
	if l.bounds == bounds {
		return
	}
	l.bounds = bounds
	l.needUpdate = true
}

func (l *legend) ProcessInput(input *view.Input) {
	if l.mousePos == input.MousePos {
		return
	}
	l.mousePos = input.MousePos
	l.needUpdate = true
}

func (l *legend) Update() (dirty bool) {
	if !l.needUpdate {
		return false
	}

	defer func() { l.needUpdate = false }()

	if l.mousePos == nil || !l.mousePos.WithinX(l.bounds) {
		l.renderable = false
		return true
	}

	if l.data.TradingSessionSeries == nil {
		l.renderable = false
		return true
	}

	ts := l.data.TradingSessionSeries.TradingSessions
	percent := float64(l.mousePos.X-l.bounds.Min.X) / float64(l.bounds.Dx())
	i := int(math.Floor(float64(len(ts)) * percent))
	if i >= len(ts) {
		i = len(ts) - 1
	}

	curr := ts[i]
	prev := curr
	if i > 0 {
		prev = ts[i-1]
	}

	formatFloat := func(value float32) string {
		return fmt.Sprintf("%.2f", value)
	}

	formatChange := func(change float32) string {
		return fmt.Sprintf("%+.2f", change)
	}

	formatPercentChange := func(percentChange float32) string {
		return fmt.Sprintf("%+.2f%%", percentChange)
	}

	var empty legendCell

	text := func(text string) legendCell {
		return legendCell{
			renderer: legendTextRenderer,
			text:     text,
			color:    color.White,
			size:     legendTextRenderer.Measure(text),
		}
	}

	symbol := func(text string, color color.RGBA) legendCell {
		return legendCell{
			renderer: legendGeometricShapeRenderer,
			text:     text,
			color:    color,
			size:     legendGeometricShapeRenderer.Measure(text),
		}
	}

	changeSymbol := func(change float32) legendCell {
		switch {
		case change > 0:
			return symbol("▲", color.Green)
		case change < 0:
			return symbol("▼", color.Red)
		default:
			return empty
		}
	}

	rows := [][3]legendCell{
		{empty, text(curr.Date.Format("1/2/06")), empty},
		{empty, empty, empty},
		{
			changeSymbol(curr.Open - prev.Open),
			text("Open"),
			text(formatFloat(curr.Open)),
		},
		{
			changeSymbol(curr.High - prev.High),
			text("High"),
			text(formatFloat(curr.High)),
		},
		{
			changeSymbol(curr.Low - prev.Low),
			text("Low"),
			text(formatFloat(curr.Low)),
		},
		{
			changeSymbol(curr.Close - prev.Close),
			text("Close"),
			text(formatFloat(curr.Close)),
		},
		{empty, empty, empty},
		{
			changeSymbol(curr.Change),
			text("Change"),
			text(formatChange(curr.Change)),
		},
		{empty, empty, text(formatPercentChange(curr.PercentChange))},
	}

	var m25, m50, m200 []*model.MovingAverage

	if s := l.data.MovingAverageSeries25; s != nil {
		m25 = s.MovingAverages
	}
	if s := l.data.MovingAverageSeries50; s != nil {
		m50 = s.MovingAverages
	}
	if s := l.data.MovingAverageSeries200; s != nil {
		m200 = s.MovingAverages
	}

	if len(m25) != 0 || len(m50) != 0 || len(m200) != 0 {
		rows = append(rows, [3]legendCell{empty, empty, empty})
	}

	if len(m25) != 0 {
		rows = append(rows, [3]legendCell{
			symbol("◼", color.Purple),
			text("SMA 25"),
			text(formatFloat(m25[i].Value)),
		})
	}

	if len(m50) != 0 {
		rows = append(rows, [3]legendCell{
			symbol("◼", color.Yellow),
			text("SMA 50"),
			text(formatFloat(m50[i].Value)),
		})
	}

	if len(m200) != 0 {
		rows = append(rows, [3]legendCell{
			symbol("◼", color.White),
			text("SMA 200"),
			text(formatFloat(m200[i].Value)),
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
	tableBounds = tableBounds.Add(l.bounds.Inset(legendBubbleMargin).Min)

	l.table = legendTable{rows, columns}
	l.tableBubble.SetBounds(tableBounds)
	l.renderable = true

	return true
}

func (l *legend) Render(fudge float32) {
	if !l.renderable {
		return
	}

	l.tableBubble.Render(fudge)

	lowerLeft := l.tableBubble.Bounds().Inset(legendTablePadding).Min
	for i := len(l.table.rows) - 1; i >= 0; i-- {
		{
			row := l.table.rows[i]
			pt := lowerLeft

			row[0].Render(pt)
			pt.X += l.table.columns[0].maxWidth + legendTablePadding

			row[1].Render(pt)
			pt.X += l.table.columns[1].maxWidth + legendTablePadding

			row[2].Render(pt)
		}
		lowerLeft.Y += legendTextRenderer.LineHeight()
	}
}

func (l *legend) Close() {
	l.renderable = false
}
