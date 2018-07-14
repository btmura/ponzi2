package view

import (
	"bytes"
	"io"

	"github.com/btmura/ponzi2/internal/gfx"
)

func dataLineVAO(yValues []float32, yRange [2]float32, color [3]float32) *gfx.VAO {
	if len(yValues) < 2 {
		return gfx.EmptyVAO()
	}

	dx := 2.0 / float32(len(yValues)) // (-1 to 1) on X-axis
	xc := func(i int) float32 {
		return -1.0 + dx*float32(i) + dx*0.5
	}

	minY, maxY := yRange[0], yRange[1]
	yc := func(v float32) float32 {
		return 2.0*(v-minY)/(maxY-minY) - 1.0
	}

	data := &gfx.VAOVertexData{Mode: gfx.Lines}

	first := true
	var v uint16 // vertex index
	for i, val := range yValues {
		if val < minY || val > maxY {
			continue
		}
		data.Vertices = append(data.Vertices, xc(i), yc(val), 0)
		data.Colors = append(data.Colors, color[0], color[1], color[2])
		if !first {
			data.Indices = append(data.Indices, v, v-1)
		}
		v++
		first = false
	}

	return gfx.NewVAO(data)
}

func horizLineSetVAO(yValues []float32, yRange [2]float32, color [3]float32) *gfx.VAO {
	if len(yValues) < 2 {
		return gfx.EmptyVAO()
	}

	minY, maxY := yRange[0], yRange[1]
	yc := func(v float32) float32 {
		return 2.0*(v-minY)/(maxY-minY) - 1.0
	}

	data := &gfx.VAOVertexData{Mode: gfx.Lines}

	var v uint16 // vertex index
	for _, val := range yValues {
		if val < minY || val > maxY {
			continue
		}
		data.Vertices = append(data.Vertices,
			-1, yc(val), 0,
			+1, yc(val), 0,
		)
		data.Colors = append(data.Colors,
			color[0], color[1], color[2],
			color[0], color[1], color[2],
		)
		data.Indices = append(data.Indices, v, v+1)
		v += 2
	}

	return gfx.NewVAO(data)
}

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
