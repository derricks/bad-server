package badness

import (
	"math"
)

// float32sEqual compares float1 and float2 and returns if they're
// equal within tolerance
func float32sEqual(float1, float2, tolerance float32) bool {
	return math.Abs(float64(float2-float1)) < float64(tolerance)
}
