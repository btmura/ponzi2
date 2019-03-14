package view

import (
	"bytes"
	"io"

	"github.com/btmura/ponzi2/internal/app/gfx"
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

// horizRuleSetVAO returns a set of horizontal lines at different y values.
func horizRuleSetVAO(yValues []float32, yRange [2]float32, color [3]float32) *gfx.VAO {
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

// vertRuleSetVAO returns a set of vertical lines at different x values.
func vertRuleSetVAO(xValues []float32, xRange [2]float32, color [3]float32) *gfx.VAO {
	if len(xValues) < 2 {
		return gfx.EmptyVAO()
	}

	minX, maxX := xRange[0], xRange[1]
	xc := func(v float32) float32 {
		return 2.0*(v-minX)/(maxX-minX) - 1.0
	}

	data := &gfx.VAOVertexData{Mode: gfx.Lines}

	var v uint16 // vertex index
	for _, val := range xValues {
		if val < minX || val > maxX {
			continue
		}
		data.Vertices = append(data.Vertices,
			xc(val), -1, 0,
			xc(val), +1, 0,
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

// horizLineVAO returns a horizontal line from (-1, 0) to (1, 0).
func horizLineVAO(color [3]float32) *gfx.VAO {
	return gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode: gfx.Lines,
			Vertices: []float32{
				-1, 0, 0,
				+1, 0, 0,
			},
			Colors: []float32{
				color[0], color[1], color[2],
				color[0], color[1], color[2],
			},
			Indices: []uint16{
				0, 1,
			},
		},
	)
}

// vertLineVAO returns a vertical line from (0, -1) to (0, 1).
func vertLineVAO(color [3]float32) *gfx.VAO {
	return gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode: gfx.Lines,
			Vertices: []float32{
				0, -1, 0,
				0, +1, 0,
			},
			Colors: []float32{
				color[0], color[1], color[2],
				color[0], color[1], color[2],
			},
			Indices: []uint16{
				0, 1,
			},
		},
	)
}

// texturedSquareVAO returns a VAO that renders a square image.
func texturedSquareVAO(textureReader io.Reader) *gfx.VAO {
	return gfx.TexturedPLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/texturedsquareplane.ply")), textureReader)
}
