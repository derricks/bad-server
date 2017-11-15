package badness

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"
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

// parseHeadersWithKeyValues takes all the key=value; pairs within
// a slice of headerValues and parses them into a map of key: value.
// duplicate keys will overwrite earlier instances
// Any values without an = or with an = followed by nothing
// will be set to an empty string.
func parseHeadersWithKeyValues(headerValues []string, parseWith string) map[string]string {
	keyValues := make(map[string]string)
	for _, valueString := range headerValues {
		pairs := strings.Split(valueString, parseWith)
		for _, pair := range pairs {
			key, value := parseKeyValuePair(pair)
			keyValues[key] = value
		}
	}
	return keyValues
}

// parseKeyValuePair takes a string of "key=value", "key", or "key="
// and returns the parsed key and value (or an empty string)
// only the first = is used for the split
func parseKeyValuePair(pair string) (key string, value string) {
	if pair == "" {
		return "", ""
	}

	pairParts := strings.SplitN(pair, "=", 2)
	if len(pairParts) < 2 {
		return pairParts[0], ""
	} else {
		return pairParts[0], pairParts[1]
	}
}

// stringToDuration will convert a string to a duration, including defaulting
// to milliseconds if it's only numeric
func stringToDuration(toParse string) (time.Duration, error) {
	// if it parses as a straight duration, use that
	duration, durationErr := time.ParseDuration(toParse)
	if durationErr == nil {
		return duration, nil
	}

	millis, intErr := strconv.Atoi(toParse)
	if intErr == nil {
		return time.Duration(millis) * time.Millisecond, nil
	}

	// use the duration error, since it's more useful to the client.
	return time.Duration(0), durationErr
}
