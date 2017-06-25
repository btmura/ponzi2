package math2

import (
	"fmt"
	"math"
)

// Matrix4 is a 4x4 matrix.
type Matrix4 [16]float32

func PerspectiveMatrix(fovRadians, aspect, near, far float32) Matrix4 {
	f := float32(math.Tan(math.Pi*0.5 - 0.5*float64(fovRadians)))
	rangeInv := 1.0 / (near - far)
	return Matrix4{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (near + far) * rangeInv, -1,
		0, 0, near * far * rangeInv * 2, 0,
	}
}

func OrthoMatrix(width, height, depth float32) Matrix4 {
	return Matrix4{
		2 / float32(width), 0, 0, 0,
		0, 2 / float32(height), 0, 0,
		0, 0, 2 / float32(depth), 0,
		-1, -1, 0, 1,
	}
}

func ViewMatrix(cameraPosition, target, up Vector3) Matrix4 {
	cameraMatrix := lookAtMatrix(cameraPosition, target, up)
	return cameraMatrix.Inverse()
}

func lookAtMatrix(cameraPosition, target, up Vector3) Matrix4 {
	zAxis := cameraPosition.Sub(target).Normalize()
	xAxis := up.Cross(zAxis)
	yAxis := zAxis.Cross(xAxis)
	return Matrix4{
		xAxis.X, xAxis.Y, xAxis.Z, 0,
		yAxis.X, yAxis.Y, yAxis.Z, 0,
		zAxis.X, zAxis.Y, zAxis.Z, 0,
		cameraPosition.X, cameraPosition.Y, cameraPosition.Z, 1,
	}
}

func TranslationMatrix(x, y, z float32) Matrix4 {
	return Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

func RotationXMatrix(radians float32) Matrix4 {
	c := float32(math.Cos(float64(radians)))
	s := float32(math.Sin(float64(radians)))
	return Matrix4{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

func RotationYMatrix(radians float32) Matrix4 {
	c := float32(math.Cos(float64(radians)))
	s := float32(math.Sin(float64(radians)))
	return Matrix4{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

func RotationZMatrix(radians float32) Matrix4 {
	c := float32(math.Cos(float64(radians)))
	s := float32(math.Sin(float64(radians)))
	return Matrix4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func ScaleMatrix(sx, sy, sz float32) Matrix4 {
	return Matrix4{
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	}
}

func QuaternionMatrix(q Quaternion) Matrix4 {
	xx := q.X * q.X
	xy := q.X * q.Y
	xz := q.X * q.Z
	xw := q.X * q.W

	yy := q.Y * q.Y
	yz := q.Y * q.Z
	yw := q.Y * q.W

	zw := q.Z * q.W
	zz := q.Z * q.Z

	return Matrix4{
		1 - 2*yy - 2*zz,
		2*xy - 2*zw,
		2*xz + 2*yw,
		0,

		2*xy + 2*zw,
		1 - 2*xx - 2*zz,
		2*yz - 2*xw,
		0,

		2*xz - 2*yw,
		2*yz + 2*xw,
		1 - 2*xx - 2*yy,
		0,

		0,
		0,
		0,
		1,
	}
}

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

func (m Matrix4) Inverse() Matrix4 {
	m00 := m[0*4+0]
	m01 := m[0*4+1]
	m02 := m[0*4+2]
	m03 := m[0*4+3]
	m10 := m[1*4+0]
	m11 := m[1*4+1]
	m12 := m[1*4+2]
	m13 := m[1*4+3]
	m20 := m[2*4+0]
	m21 := m[2*4+1]
	m22 := m[2*4+2]
	m23 := m[2*4+3]
	m30 := m[3*4+0]
	m31 := m[3*4+1]
	m32 := m[3*4+2]
	m33 := m[3*4+3]
	tmp0 := m22 * m33
	tmp1 := m32 * m23
	tmp2 := m12 * m33
	tmp3 := m32 * m13
	tmp4 := m12 * m23
	tmp5 := m22 * m13
	tmp6 := m02 * m33
	tmp7 := m32 * m03
	tmp8 := m02 * m23
	tmp9 := m22 * m03
	tmp10 := m02 * m13
	tmp11 := m12 * m03
	tmp12 := m20 * m31
	tmp13 := m30 * m21
	tmp14 := m10 * m31
	tmp15 := m30 * m11
	tmp16 := m10 * m21
	tmp17 := m20 * m11
	tmp18 := m00 * m31
	tmp19 := m30 * m01
	tmp20 := m00 * m21
	tmp21 := m20 * m01
	tmp22 := m00 * m11
	tmp23 := m10 * m01

	t0 := (tmp0*m11 + tmp3*m21 + tmp4*m31) - (tmp1*m11 + tmp2*m21 + tmp5*m31)
	t1 := (tmp1*m01 + tmp6*m21 + tmp9*m31) - (tmp0*m01 + tmp7*m21 + tmp8*m31)
	t2 := (tmp2*m01 + tmp7*m11 + tmp10*m31) - (tmp3*m01 + tmp6*m11 + tmp11*m31)
	t3 := (tmp5*m01 + tmp8*m11 + tmp11*m21) - (tmp4*m01 + tmp9*m11 + tmp10*m21)

	d := 1.0 / (m00*t0 + m10*t1 + m20*t2 + m30*t3)

	return Matrix4{
		d * t0,
		d * t1,
		d * t2,
		d * t3,
		d * ((tmp1*m10 + tmp2*m20 + tmp5*m30) - (tmp0*m10 + tmp3*m20 + tmp4*m30)),
		d * ((tmp0*m00 + tmp7*m20 + tmp8*m30) - (tmp1*m00 + tmp6*m20 + tmp9*m30)),
		d * ((tmp3*m00 + tmp6*m10 + tmp11*m30) - (tmp2*m00 + tmp7*m10 + tmp10*m30)),
		d * ((tmp4*m00 + tmp9*m10 + tmp10*m20) - (tmp5*m00 + tmp8*m10 + tmp11*m20)),
		d * ((tmp12*m13 + tmp15*m23 + tmp16*m33) - (tmp13*m13 + tmp14*m23 + tmp17*m33)),
		d * ((tmp13*m03 + tmp18*m23 + tmp21*m33) - (tmp12*m03 + tmp19*m23 + tmp20*m33)),
		d * ((tmp14*m03 + tmp19*m13 + tmp22*m33) - (tmp15*m03 + tmp18*m13 + tmp23*m33)),
		d * ((tmp17*m03 + tmp20*m13 + tmp23*m23) - (tmp16*m03 + tmp21*m13 + tmp22*m23)),
		d * ((tmp14*m22 + tmp17*m32 + tmp13*m12) - (tmp16*m32 + tmp12*m12 + tmp15*m22)),
		d * ((tmp20*m32 + tmp12*m02 + tmp19*m22) - (tmp18*m22 + tmp21*m32 + tmp13*m02)),
		d * ((tmp18*m12 + tmp23*m32 + tmp15*m02) - (tmp22*m32 + tmp14*m02 + tmp19*m12)),
		d * ((tmp22*m22 + tmp16*m02 + tmp21*m12) - (tmp20*m12 + tmp23*m22 + tmp17*m02)),
	}
}

func (m Matrix4) Transpose() Matrix4 {
	return Matrix4{
		m[0], m[4], m[8], m[12],
		m[1], m[5], m[9], m[13],
		m[2], m[6], m[10], m[14],
		m[3], m[7], m[11], m[15],
	}
}

func (m Matrix4) String() string {
	return fmt.Sprintf("%8.2f %8.2f %8.2f %8.2f\n%8.2f %8.2f %8.2f %8.2f\n%8.2f %8.2f %8.2f %8.2f\n%8.2f %8.2f %8.2f %8.2f", m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15])
}
