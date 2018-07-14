package view

import "github.com/btmura/ponzi2/internal/gfx"

func graphLineVAO(values []float32, valRange [2]float32, color [3]float32) *gfx.VAO {
	dx := 2.0 / float32(len(values)) // (-1 to 1) on X-axis
	xc := func(i int) float32 {
		return -1.0 + dx*float32(i) + dx*0.5
	}

	min, max := valRange[0], valRange[1]
	yc := func(v float32) float32 {
		return 2.0*(v-min)/(max-min) - 1.0
	}

	data := &gfx.VAOVertexData{Mode: gfx.Lines}
	if len(values) == 0 {
		return gfx.NewVAO(data)
	}

	first := true
	var v uint16 // vertex index
	for i, val := range values {
		if val < min || val > max {
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
