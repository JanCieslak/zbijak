package utils

import (
	"constraints"
	"math"
)

func MapValue(x, inMin, inMax, outMin, outMax float64) float64 {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}

func Lerp(start, end, p float64) float64 {
	return start + (end-start)*p
}

func Slerp(start, end, p float64) float64 {
	// Dupa fix
	if math.Abs(end-start) > 2 {
		return end
	}
	return start + (end-start)*p
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
