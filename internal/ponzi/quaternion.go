package ponzi

import "math"

type quaternion struct {
	x float32
	y float32
	z float32
	w float32
}

func newAxisAngleQuaternion(axis vector3, angleInRadians float32) quaternion {
	axis = axis.normalize()
	halfSin := float32(math.Sin(float64(angleInRadians * 0.5)))
	halfCos := float32(math.Cos(float64(angleInRadians * 0.5)))
	return quaternion{
		axis.x * halfSin,
		axis.y * halfSin,
		axis.z * halfSin,
		halfCos,
	}
}

func (q quaternion) mult(r quaternion) quaternion {
	return quaternion{
		q.w*r.x + q.x*r.w + q.y*r.z - q.z*r.y,
		q.w*r.y + q.y*r.w + q.z*r.x - q.x*r.z,
		q.w*r.z + q.z*r.w + q.x*r.y - q.y*r.x,
		q.w*r.w - q.x*r.x - q.y*r.y - q.z*r.z,
	}
}

func (q quaternion) conjugate() quaternion {
	return quaternion{-q.x, -q.y, -q.z, q.w}
}

func (q quaternion) normalize() quaternion {
	l := float32(math.Sqrt(float64(q.x*q.x + q.y*q.y + q.z*q.z + q.w*q.w)))
	if l > 0.00001 {
		return quaternion{q.x / l, q.y / l, q.z / l, q.w / l}
	}
	return quaternion{}
}

func (q quaternion) rotate(v vector3) vector3 {
	r := q.mult(quaternion{v.x, v.y, v.z, 0}).mult(q.conjugate())
	return vector3{r.x, r.y, r.z}
}
