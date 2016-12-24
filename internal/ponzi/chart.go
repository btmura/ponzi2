package ponzi

import "github.com/go-gl/gl/v4.5-core/gl"

type chart struct {
	vao   uint32
	count int32
}

func createChart(stock *modelStock) *chart {
	vertices := []float32{
		0, 0,
		1, 0,
		0, 1,
		1, 1,
	}
	indices := []uint16{
		1, 2, 0,
		1, 3, 2,
	}

	vbo := createArrayBuffer(vertices)
	ibo := createElementArrayBuffer(indices)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.EnableVertexAttribArray(positionLocation)
	gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BindVertexArray(0)

	return &chart{
		vao:   vao,
		count: int32(len(indices)),
	}
}

func (c *chart) draw() {
	gl.BindVertexArray(c.vao)
	gl.DrawElements(gl.LINES, c.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}
