package math

import "math"

type Quaternion struct {
	X float32
	Y float32
	Z float32
	W float32
}

func NewAxisAngleQuaternion(axis Vector3, angleInRadians float32) Quaternion {
	axis = axis.Normalize()
	halfSin := float32(math.Sin(float64(angleInRadians * 0.5)))
	halfCos := float32(math.Cos(float64(angleInRadians * 0.5)))
	return Quaternion{
		axis.X * halfSin,
		axis.Y * halfSin,
		axis.Z * halfSin,
		halfCos,
	}
}

func (q Quaternion) Mult(r Quaternion) Quaternion {
	return Quaternion{
		q.W*r.X + q.X*r.W + q.Y*r.Z - q.Z*r.Y,
		q.W*r.Y + q.Y*r.W + q.Z*r.X - q.X*r.Z,
		q.W*r.Z + q.Z*r.W + q.X*r.Y - q.Y*r.X,
		q.W*r.W - q.X*r.X - q.Y*r.Y - q.Z*r.Z,
	}
}

func (q Quaternion) Conjugate() Quaternion {
	return Quaternion{-q.X, -q.Y, -q.Z, q.W}
}

func (q Quaternion) Normalize() Quaternion {
	l := float32(math.Sqrt(float64(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)))
	if l > 0.00001 {
		return Quaternion{q.X / l, q.Y / l, q.Z / l, q.W / l}
	}
	return Quaternion{}
}

func (q Quaternion) Rotate(v Vector3) Vector3 {
	r := q.Mult(Quaternion{v.X, v.Y, v.Z, 0}).Mult(q.Conjugate())
	return Vector3{r.X, r.Y, r.Z}
}
