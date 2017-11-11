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

func TestParseKeyValuePair(test *testing.T) {
	expectations := map[string][]string{
		"":        []string{"", ""},
		"500":     []string{"500", ""},
		"500=":    []string{"500", ""},
		"500=x":   []string{"500", "x"},
		"500=x=y": []string{"500", "x=y"},
		"=":       []string{"", ""},
	}

	for input, expect := range expectations {
		key, value := parseKeyValuePair(input)
		if key != expect[0] {
			test.Fatalf("Expected key of %s from %s, got %s", expect[0], input, key)
		}

		if value != expect[1] {
			test.Fatalf("Expected value of %s from %s, got %s", expect[1], input, value)
		}
	}
}

type headerParseExpect struct {
	headerValues []string
	expectedMap  map[string]string
}

func TestParseHeadersWithKeyValues(test *testing.T) {
	expectations := []headerParseExpect{
		headerParseExpect{[]string{"500"}, map[string]string{"500": ""}},
		headerParseExpect{[]string{"500=3,200=6"}, map[string]string{"500": "3", "200": "6"}},
		headerParseExpect{[]string{"500=3", "200=6"}, map[string]string{"500": "3", "200": "6"}},
		headerParseExpect{[]string{"500=3,200=6", "500=5"}, map[string]string{"500": "5", "200": "6"}},
	}

	for _, expectation := range expectations {
		actualMap := parseHeadersWithKeyValues(expectation.headerValues, ",")
		if len(actualMap) != len(expectation.expectedMap) {
			test.Fatalf("Expected map %v is a different length than actual map %v", expectation.expectedMap, actualMap)
		}

		for expectKey, expectValue := range expectation.expectedMap {
			actualValue, found := actualMap[expectKey]
			if !found {
				test.Fatalf("Did not find key %s in %v", expectKey, actualMap)
			}

			if actualValue != expectValue {
				test.Fatalf("Expected value %s for key %s but got %s", expectValue, expectKey, actualValue)
			}
		}
	}
}
