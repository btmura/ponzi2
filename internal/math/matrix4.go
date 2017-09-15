// Package math has functions to work with matrices.
package math

import (
	"fmt"
)

// Matrix4 is a 4x4 matrix.
type Matrix4 [16]float32

// OrthoMatrix creates a orthographic projection matrix.
func OrthoMatrix(width, height, depth float32) Matrix4 {
	return Matrix4{
		2 / float32(width), 0, 0, 0,
		0, 2 / float32(height), 0, 0,
		0, 0, 2 / float32(depth), 0,
		-1, -1, 0, 1,
	}
}

// TranslationMatrix creates a translation matrix.
func TranslationMatrix(x, y, z float32) Matrix4 {
	return Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

// ScaleMatrix creates a scale matrix.
func ScaleMatrix(sx, sy, sz float32) Matrix4 {
	return Matrix4{
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	}
}

// Mult returns a fresh matrix that is the product of the receiver and the argument.
func (m Matrix4) Mult(n Matrix4) Matrix4 {
	return Matrix4{
		m[0]*n[0] + m[1]*n[4] + m[2]*n[8] + m[3]*n[12],
		m[0]*n[1] + m[1]*n[5] + m[2]*n[9] + m[3]*n[13],
		m[0]*n[2] + m[1]*n[6] + m[2]*n[10] + m[3]*n[14],
		m[0]*n[3] + m[1]*n[7] + m[2]*n[11] + m[3]*n[15],

		m[4]*n[0] + m[5]*n[4] + m[6]*n[8] + m[7]*n[12],
		m[4]*n[1] + m[5]*n[5] + m[6]*n[9] + m[7]*n[13],
		m[4]*n[2] + m[5]*n[6] + m[6]*n[10] + m[7]*n[14],
		m[4]*n[3] + m[5]*n[7] + m[6]*n[11] + m[7]*n[15],

		m[8]*n[0] + m[9]*n[4] + m[10]*n[8] + m[11]*n[12],
		m[8]*n[1] + m[9]*n[5] + m[10]*n[9] + m[11]*n[13],
		m[8]*n[2] + m[9]*n[6] + m[10]*n[10] + m[11]*n[14],
		m[8]*n[3] + m[9]*n[7] + m[10]*n[11] + m[11]*n[15],

		m[12]*n[0] + m[13]*n[4] + m[14]*n[8] + m[15]*n[12],
		m[12]*n[1] + m[13]*n[5] + m[14]*n[9] + m[15]*n[13],
		m[12]*n[2] + m[13]*n[6] + m[14]*n[10] + m[15]*n[14],
		m[12]*n[3] + m[13]*n[7] + m[14]*n[11] + m[15]*n[15],
	}
}

func (m Matrix4) String() string {
	return fmt.Sprintf("%8.2f %8.2f %8.2f %8.2f\n%8.2f %8.2f %8.2f %8.2f\n%8.2f %8.2f %8.2f %8.2f\n%8.2f %8.2f %8.2f %8.2f", m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15])
}
