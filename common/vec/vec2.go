package vec

import "math"

type Vec2 struct {
	X, Y float64
}

var Right = Vec2{X: 1, Y: 0}

func NewIVec2(x, y int) Vec2 {
	return Vec2{
		float64(x),
		float64(y),
	}
}

func NewVec2(x, y float64) Vec2 {
	return Vec2{
		x,
		y,
	}
}

func (v *Vec2) Set(x, y float64) {
	v.X = x
	v.Y = y
}

func (v *Vec2) SetFrom(other Vec2) {
	v.X = other.X
	v.Y = other.Y
}

func (v *Vec2) Add(x, y float64) {
	v.X += x
	v.Y += y
}

func (v *Vec2) AddVec(other Vec2) {
	v.X += other.X
	v.Y += other.Y
}

func (v *Vec2) SubVec(other Vec2) {
	v.X -= other.X
	v.Y -= other.Y
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

func (v Vec2) Dot(v2 Vec2) float64 {
	return v.X*v2.X + v.Y*v2.Y
}
