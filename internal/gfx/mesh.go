package gfx

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/golang/glog"

	gl2 "github.com/btmura/ponzi2/internal/gl"
	"github.com/btmura/ponzi2/internal/obj"
)

// mesh is an OBJ file object with a bunch of vertex buffer objects.
type mesh struct {
	// ID is the object's ID in the OBJ file.
	ID string

	// vao is the vertex array object name.
	vao uint32

	// count is how many elements to render using gl.DrawElements.
	count int32
}

// createMeshes creates a slice of meshes from a slice of objs.
func createMeshes(objs []*obj.Object) []*mesh {
	var vertexTable []*obj.Vertex
	var normalTable []*obj.Normal
	var texCoordTable []*obj.TexCoord

	var vertices []float32
	var normals []float32
	var texCoords []float32

	elementIndexMap := map[obj.FaceElement]uint16{}
	var nextIndex uint16

	var meshes []*mesh
	var iboNames []uint32

	for _, o := range objs {
		for _, v := range o.Vertices {
			vertexTable = append(vertexTable, v)
		}
		for _, n := range o.Normals {
			normalTable = append(normalTable, n)
		}
		for _, tc := range o.TexCoords {
			texCoordTable = append(texCoordTable, tc)
		}

		var indices []uint16
		for _, f := range o.Faces {
			for _, e := range f {
				if _, exists := elementIndexMap[e]; !exists {
					elementIndexMap[e] = nextIndex
					nextIndex++

					v := vertexTable[e.VertexIndex-1]
					vertices = append(vertices, v.X, v.Y, v.Z)

					n := normalTable[e.NormalIndex-1]
					normals = append(normals, n.X, n.Y, n.Z)

					// Flip the y-axis to convert from OBJ to OpenGL.
					// OpenGL considers the origin to be lower left.
					// OBJ considers the origin to be upper left.
					tc := texCoordTable[e.TexCoordIndex-1]
					texCoords = append(texCoords, tc.S, 1.0-tc.T)
				}

				indices = append(indices, elementIndexMap[e])
			}
		}

		meshes = append(meshes, &mesh{
			ID:    o.ID,
			count: int32(len(indices)),
		})
		iboNames = append(iboNames, gl2.ElementArrayBuffer(indices))
	}

	glog.Infof("vertices: %d", len(vertexTable))
	glog.Infof("normals: %d", len(normalTable))
	glog.Infof("texCoords: %d", len(texCoordTable))

	vbo := gl2.ArrayBuffer(vertices)
	nbo := gl2.ArrayBuffer(normals)
	tbo := gl2.ArrayBuffer(texCoords)

	for i, m := range meshes {
		gl.GenVertexArrays(1, &m.vao)
		gl.BindVertexArray(m.vao)

		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ARRAY_BUFFER, nbo)
		gl.EnableVertexAttribArray(normalLocation)
		gl.VertexAttribPointer(normalLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ARRAY_BUFFER, tbo)
		gl.EnableVertexAttribArray(texCoordLocation)
		gl.VertexAttribPointer(texCoordLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, iboNames[i])
		gl.BindVertexArray(0)
	}

	return meshes
}

// drawElements draws the Mesh's elements.
func (m *mesh) drawElements() {
	gl.BindVertexArray(m.vao)
	gl.DrawElements(gl.TRIANGLES, m.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}
