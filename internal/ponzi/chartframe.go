package ponzi

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

const roundAmount = 10

type chartFrame struct {
	stock           *modelStock
	roundedCornerNW *gfx.VAO
	roundedCornerNE *gfx.VAO
	roundedCornerSE *gfx.VAO
	roundedCornerSW *gfx.VAO
	horizDivider    *gfx.VAO
	vertDivider     *gfx.VAO
}

func createChartFrame(stock *modelStock) *chartFrame {
	return &chartFrame{
		stock:           stock,
		roundedCornerNW: gfx.ReadPLYVAO(bytes.NewReader(MustAsset("roundedCornerNW.ply"))),
		roundedCornerNE: gfx.ReadPLYVAO(bytes.NewReader(MustAsset("roundedCornerNE.ply"))),
		roundedCornerSE: gfx.ReadPLYVAO(bytes.NewReader(MustAsset("roundedCornerSE.ply"))),
		roundedCornerSW: gfx.ReadPLYVAO(bytes.NewReader(MustAsset("roundedCornerSW.ply"))),
		horizDivider:    gfx.HorizColoredLineVAO(white, white),
		vertDivider:     gfx.VertColoredLineVAO(white, white),
	}
}

func (ch *chartFrame) render(r image.Rectangle) []image.Rectangle {
	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gfx.SetColorMixAmount(1)

	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y-roundAmount, r.Min.X+roundAmount, r.Max.Y))
	ch.roundedCornerNW.Render()

	gfx.SetModelMatrixRect(image.Rect(r.Max.X-roundAmount, r.Max.Y-roundAmount, r.Max.X, r.Max.Y))
	ch.roundedCornerNE.Render()

	gfx.SetModelMatrixRect(image.Rect(r.Max.X-roundAmount, r.Min.Y, r.Max.X, r.Min.Y+roundAmount))
	ch.roundedCornerSE.Render()

	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Min.Y, r.Min.X+roundAmount, r.Min.Y+roundAmount))
	ch.roundedCornerSW.Render()

	// North Line
	gfx.SetModelMatrixRect(image.Rect(r.Min.X+roundAmount, r.Max.Y, r.Max.X-roundAmount, r.Max.Y))
	ch.horizDivider.Render()

	// South Line
	gfx.SetModelMatrixRect(image.Rect(r.Min.X+roundAmount, r.Min.Y, r.Max.X-roundAmount, r.Min.Y))
	ch.horizDivider.Render()

	// West Line
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Min.Y+roundAmount, r.Min.X, r.Max.Y-roundAmount))
	ch.vertDivider.Render()

	// East Line
	gfx.SetModelMatrixRect(image.Rect(r.Max.X, r.Min.Y+roundAmount, r.Max.X, r.Max.Y-roundAmount))
	ch.vertDivider.Render()

	gfx.SetModelMatrixRect(r)

	//
	// Render the symbol, quote, and add button.
	//

	const pad = 5
	pt.Y -= pad + symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += roundAmount
		pt.X += symbolQuoteTextRenderer.Render(ch.stock.symbol, pt, white)
		pt.X += pad
		pt.X += symbolQuoteTextRenderer.Render(formatQuote(ch.stock.quote), pt, quoteColor(ch.stock.quote))
	}
	pt.Y -= pad

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = pt.Y
	gfx.SetColorMixAmount(1)

	rects := sliceRectangle(r, 0.13, 0.13, 0.13)
	for _, r := range rects {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		ch.horizDivider.Render()
	}
	return rects
}

func (ch *chartFrame) close() {
	ch.roundedCornerNW.Delete()
	ch.roundedCornerNW = nil
	ch.roundedCornerNE.Delete()
	ch.roundedCornerNE = nil
	ch.roundedCornerSE.Delete()
	ch.roundedCornerSE = nil
	ch.roundedCornerSW.Delete()
	ch.roundedCornerSW = nil
	ch.horizDivider.Delete()
	ch.horizDivider = nil
	ch.vertDivider.Delete()
	ch.vertDivider = nil
}
