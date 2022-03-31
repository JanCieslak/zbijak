package common

func MapValue(x, inMin, inMax, outMin, outMax float64) float64 {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}

func Lerp(start, end, p float64) float64 {
	return start + (end-start)*p
}
