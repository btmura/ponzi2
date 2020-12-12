package chart

import (
	"fmt"
	"image"
	"time"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

const (
	legendBubbleMargin   = 10
	legendBubbleRounding = 10
	legendTablePadding   = 10
	legendFontSize       = 20
)

var (
	legendTextRenderer = gfx.NewTextRenderer(goregular.TTF, legendFontSize)

	// legendGeometricShapeRenderer renderers geometric shapes.
	// https://www.fileformat.info/info/unicode/font/dejavu_sans_mono/blockview.htm?block=geometric_shapes
	legendGeometricShapeRenderer = gfx.NewTextRenderer(_escFSMustByte(false, "/data/DejaVuSans.ttf"), legendFontSize)
)

type legendTable struct {
	bubble  *rect.Bubble
	rows    [][3]legendCell
	columns [3]legendColumn
}

func (l *legendTable) SetBounds(bounds image.Rectangle) {
	l.bubble.SetBounds(bounds)
}

func (l *legendTable) Render(fudge float32) {
	l.bubble.Render(fudge)

	lowerLeft := l.bubble.Bounds().Inset(legendTablePadding).Min
	for i := len(l.rows) - 1; i >= 0; i-- {
		{
			row := l.rows[i]
			pt := lowerLeft

			row[0].Render(pt)
			pt.X += l.columns[0].maxWidth + legendTablePadding

			row[1].Render(pt)
			pt.X += l.columns[1].maxWidth + legendTablePadding

			row[2].Render(pt)
		}
		lowerLeft.Y += legendTextRenderer.LineHeight()
	}
}

type legendCell struct {
	renderer *gfx.TextRenderer
	text     string
	color    view.Color
	size     image.Point
}

func (l *legendCell) Render(pt image.Point) {
	l.renderer.Render(l.text, pt, gfx.TextColor(l.color))
}

type legendColumn struct {
	maxWidth int
}

func legendText(text string) legendCell {
	return legendCell{
		renderer: legendTextRenderer,
		text:     text,
		color:    view.White,
		size:     legendTextRenderer.Measure(text),
	}
}

func symbol(text string, color view.Color) legendCell {
	return legendCell{
		renderer: legendGeometricShapeRenderer,
		text:     text,
		color:    color,
		size:     legendGeometricShapeRenderer.Measure(text),
	}
}

func whiteArrow(change float32) legendCell {
	switch {
	case change > 0:
		return symbol("△", view.White)
	case change < 0:
		return symbol("▽", view.White)
	default:
		return legendCell{}
	}
}

func colorArrow(change float32) legendCell {
	switch {
	case change > 0:
		return symbol("▲", view.Green)
	case change < 0:
		return symbol("▼", view.Red)
	default:
		return legendCell{}
	}
}

func symbolLabel(value, threshold float32) string {
	if value >= threshold {
		return "◼"
	}
	return "☒"
}

func typeLabel(avgType model.AverageType) string {
	switch avgType {
	case model.Simple:
		return "SMA"
	case model.Exponential:
		return "EMA"
	default:
		return "?"
	}
}

func formatFloat(value float32) string {
	return fmt.Sprintf("%.2f", value)
}

func formatChange(change float32) string {
	return fmt.Sprintf("%+.2f", change)
}

func formatPercentChange(percentChange float32) string {
	return fmt.Sprintf("%+.2f%%", percentChange)
}

func formatWeekday(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "M"
	case time.Tuesday:
		return "T"
	case time.Wednesday:
		return "W"
	case time.Thursday:
		return "R"
	case time.Friday:
		return "F"
	}
	return "?"
}
