package view

import (
	"bytes"
	"io"

	"github.com/btmura/ponzi2/internal/gfx"
)

// horizColoredLineVAO returns a horizontal colored line segment
// from (-1, 0) to (1, 0).
func horizColoredLineVAO(leftColor, rightColor [3]float32) *gfx.VAO {
	return gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode: gfx.Lines,
			Vertices: []float32{
				-1, 0, 0,
				+1, 0, 0,
			},
			Colors: []float32{
				leftColor[0], leftColor[1], leftColor[2],
				rightColor[0], rightColor[1], rightColor[2],
			},
			Indices: []uint16{
				0, 1,
			},
		},
	)
}

// vertColoredLineVAO returns a vertical colored line segment
// from (0, -1) to (0, 1).
func vertColoredLineVAO(topColor, botColor [3]float32) *gfx.VAO {
	return gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode: gfx.Lines,
			Vertices: []float32{
				0, -1, 0,
				0, +1, 0,
			},
			Colors: []float32{
				topColor[0], topColor[1], topColor[2],
				botColor[0], botColor[1], botColor[2],
			},
			Indices: []uint16{
				0, 1,
			},
		},
	)
}

// squareImageVAO returns a VAO that renders a square image.
func squareImageVAO(textureReader io.Reader) *gfx.VAO {
	return gfx.TexturedPLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/texturedsquareplane.ply")), textureReader)
}
