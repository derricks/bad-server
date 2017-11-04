package badness

import (
	"math"
	"net/http"
	"net/http/httptest"
)

// float32sEqual compares float1 and float2 and returns if they're
// equal within tolerance
func float32sEqual(float1, float2, tolerance float32) bool {
	return math.Abs(float64(float2-float1)) < float64(tolerance)
}

// float64sEqual compares two float64s
func float64sEqual(float1, float2, tolerance float64) bool {
	return math.Abs(float2-float1) < tolerance
}

// getFirstHeaderValue returns the first value in the
// slice of strings tied to a header in request
func getFirstHeaderValue(request *http.Request, header string) string {
	if headerStrings, found := request.Header[header]; found {
		if len(headerStrings) > 0 {
			return headerStrings[0]
		} else {
			return ""
		}
	}
	return ""
}

// makeTestRequest constructs a mock http request
func makeTestRequest() *http.Request {
	return httptest.NewRequest("GET", "http://localhost", nil)
}
