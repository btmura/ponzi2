package ponzi

import "math"

// vector3 is a vector with x, y, and z.
type vector3 struct {
	x float32
	y float32
	z float32
}

func (v vector3) sub(w vector3) vector3 {
	return vector3{v.x - w.x, v.y - w.y, v.z - w.z}
}

func (v vector3) length() float32 {
	return float32(math.Sqrt(float64(v.x*v.x + v.y*v.y + v.z*v.z)))
}

func (v vector3) cross(w vector3) vector3 {
	return vector3{
		v.y*w.z - v.z*w.y,
		v.z*w.x - v.x*w.z,
		v.x*w.y - v.y*w.x,
	}
}

func (v vector3) normalize() vector3 {
	l := v.length()
	if l > 0.00001 {
		return vector3{v.x / l, v.y / l, v.z / l}
	}
	return vector3{}
}
