package badness

import (
	"testing"
)

func TestEqualFloats(test *testing.T) {
	if !float32sEqual(3.3, 3.3, .1) {
		test.Fatal("3.3 and 3.3 should be equal but are not")
	}
}

func TestUnequalFloats(test *testing.T) {
	if float32sEqual(3.3, 3.4, .1) {
		test.Fatal("3.3 and 3.4 should not be equal but are")
	}
}

func TestEqualFloat64s(test *testing.T) {
	if !float64sEqual(3.3, 3.3, .1) {
		test.Fatal("3.3 and 3.3 should be equal but are not")
	}
}

func TestUnequalFloat64ss(test *testing.T) {
	if float64sEqual(3.3, 3.4, .1) {
		test.Fatal("3.3 and 3.4 should not be equal but are")
	}
}

const fakeHeader = "X-Test-Header"

func TestFirstHeaderWhenNotSet(test *testing.T) {
	request := makeTestRequest()
	firstValue := getFirstHeaderValue(request, fakeHeader)
	if firstValue != "" {
		test.Fatalf("Expected blank string got %s", firstValue)
	}
}

type firstHeaderExpect struct {
	headerValues []string
	value        string
}

func TestFirstHeader(test *testing.T) {
	expectations := []firstHeaderExpect{
		firstHeaderExpect{[]string{}, ""},
		firstHeaderExpect{[]string{"value1"}, "value1"},
		firstHeaderExpect{[]string{"value2", "value1"}, "value2"},
	}

	for _, expectation := range expectations {
		request := makeTestRequest()
		request.Header[fakeHeader] = expectation.headerValues
		firstValue := getFirstHeaderValue(request, fakeHeader)
		if firstValue != expectation.value {
			test.Fatalf("Expected %s, got %s", expectation.value, firstValue)
		}
	}
}
