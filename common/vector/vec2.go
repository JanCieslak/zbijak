package vector

import "math"

type Vec2 struct {
	X, Y float64
}

func (v *Vec2) AddVec(other Vec2) {
	v.X += other.X
	v.Y += other.Y
}

func (v *Vec2) Add(x, y float64) {
	v.X += x
	v.Y += y
}

func (v *Vec2) Mul(val float64) {
	v.X *= val
	v.Y *= val
}

func (v Vec2) Len() float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2))
}

func (v *Vec2) Normalize() {
	length := v.Len()
	if length > 0 {
		v.X /= length
		v.Y /= length
	}
}
