package math2

import "math"

// Vector3 is a vector with x, y, and z.
type Vector3 struct {
	X float32
	Y float32
	Z float32
}

func (v Vector3) Sub(w Vector3) Vector3 {
	return Vector3{v.X - w.X, v.Y - w.Y, v.Z - w.Z}
}

func (v Vector3) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

func (v Vector3) Cross(w Vector3) Vector3 {
	return Vector3{
		v.Y*w.Z - v.Z*w.Y,
		v.Z*w.X - v.X*w.Z,
		v.X*w.Y - v.Y*w.X,
	}
}

func (v Vector3) Normalize() Vector3 {
	l := v.Length()
	if l > 0.00001 {
		return Vector3{v.X / l, v.Y / l, v.Z / l}
	}
	return Vector3{}
}
