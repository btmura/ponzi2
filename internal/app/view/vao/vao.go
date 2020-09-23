package vao

import (
	"bytes"
	"io"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view"
)

// Embed resources into the application. Get esc from github.com/mjibson/esc.
//go:generate esc -o bindata.go -pkg vao -include ".*(ply|png)" -modtime 1337 -private data

// DataLine returns a VAO of line segments that plot the percentage values on the Y-axis from -1 to 1 on the X-axis.
func DataLine(yPercentValues []float32, color view.Color) *gfx.VAO {
	if len(yPercentValues) < 2 {
		return gfx.EmptyVAO()
	}

	dx := 2.0 / float32(len(yPercentValues)) // (-1 to 1) on X-axis
	xc := func(i int) float32 {
		return -1.0 + dx*float32(i) + dx*0.5
	}
	yc := func(v float32) float32 {
		return 2.0*v - 1.0
	}

	data := &gfx.VAOVertexData{Mode: gfx.Lines}

	first := true
	var v uint16 // vertex index
	for i, val := range yPercentValues {
		if val <= 0 || val >= 1 {
			first = true
			continue
		}
		data.Vertices = append(data.Vertices, xc(i), yc(val), 0)
		data.Colors = append(data.Colors, color[0], color[1], color[2], color[3])
		if !first {
			data.Indices = append(data.Indices, v, v-1)
		}
		v++
		first = false
	}

	return gfx.NewVAO(data)
}

// VertRuleSet returns a set of vertical lines at different x values.
func VertRuleSet(xValues []float32, xRange [2]float32, color1, color2 view.Color) *gfx.VAO {
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
			color1[0], color1[1], color1[2], color1[3],
			color2[0], color2[1], color2[2], color2[3],
		)
		data.Indices = append(data.Indices, v, v+1)
		v += 2
	}

	return gfx.NewVAO(data)
}

// HorizLine returns a VAO of a horizontal line from (-1, 0) to (1, 0).
func HorizLine(color1, color2 view.Color) *gfx.VAO {
	return gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode: gfx.Lines,
			Vertices: []float32{
				-1, 0, 0,
				+1, 0, 0,
			},
			Colors: []float32{
				color1[0], color1[1], color1[2], color1[3],
				color2[0], color2[1], color2[2], color2[3],
			},
			Indices: []uint16{
				0, 1,
			},
		},
	)
}

// VertLine returns a VAO of a vertical line from (0, -1) to (0, 1).
func VertLine(color1, color2 view.Color) *gfx.VAO {
	return gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode: gfx.Lines,
			Vertices: []float32{
				0, -1, 0,
				0, +1, 0,
			},
			Colors: []float32{
				color1[0], color1[1], color1[2], color1[3],
				color2[0], color2[1], color2[2], color2[3],
			},
			Indices: []uint16{
				0, 1,
			},
		},
	)
}

// TexturedSquare returns a VAO that renders a square image.
func TexturedSquare(textureReader io.Reader) *gfx.VAO {
	return gfx.TexturedPLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/texturedsquareplane.ply")), textureReader)
}
