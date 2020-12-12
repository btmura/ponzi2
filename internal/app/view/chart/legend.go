package chart

import (
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
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
