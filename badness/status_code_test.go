package badness

import (
	"errors"
	"testing"
)

type histogramParseExpect struct {
	statusCode  int
	probability float32
	err         error
}

func TestHistogramEntryParsing(test *testing.T) {
	expectations := map[string]histogramParseExpect{
		"500=33.2": histogramParseExpect{500, 33.2, nil},

		"": histogramParseExpect{0, 0.0, errors.New("empty string should yield error")},

		"500": histogramParseExpect{500, 0.0, nil},

		"500=": histogramParseExpect{500, 0.0, nil},

		"500=test": histogramParseExpect{0, 0.0, errors.New("500=test should yield an error")},

		"x": histogramParseExpect{0, 0.0, errors.New("x should yield an error")},

		"x=y": histogramParseExpect{0, 0.0, errors.New("x=y should yield parse errors")},
	}

	for parseString, expectation := range expectations {
		entry, err := parseHistogramHeader(parseString)

		if expectation.err != nil && err == nil {
			test.Fatalf("Header %s should not produce an error but did: %v", parseString, expectation.err)
		}

		if expectation.err == nil && err != nil {
			test.Fatalf("Header %s should produce an error but did not: %v", parseString, err)
		}

		if entry.statusCode != expectation.statusCode {
			test.Fatalf("Header %s should yield status code of %d, has %d", parseString, expectation.statusCode, entry.statusCode)
		}

		if !float32sEqual(entry.probability, expectation.probability, .01) {
			test.Fatalf("Header %s should yield probability of %f, has %f", parseString, expectation.probability, entry.probability)
		}

	}
}

type histogramExpect struct {
	histogramValues []string
	entries         []statusCodeHistogramEntry
}

func TestHistogramGeneration(test *testing.T) {
	expectations := []histogramExpect{
		histogramExpect{[]string{"500"}, []statusCodeHistogramEntry{statusCodeHistogramEntry{500, 100.0}}},
		histogramExpect{[]string{"500=50", "200=50"}, []statusCodeHistogramEntry{statusCodeHistogramEntry{500, 50.0}, statusCodeHistogramEntry{200, 50.0}}},
		histogramExpect{[]string{"500=50,200"}, []statusCodeHistogramEntry{statusCodeHistogramEntry{500, 50.0}, statusCodeHistogramEntry{200, 50.0}}},
		histogramExpect{[]string{""}, []statusCodeHistogramEntry{}},
	}

	for _, expectation := range expectations {
		histogram := buildHistogram(expectation.histogramValues)

		if len(histogram) != len(expectation.entries) {
			test.Fatalf("A different number of entries than expected, %d vs %d", len(histogram), len(expectation.entries))
		}

		for index, expectedEntry := range expectation.entries {
			actualEntry := histogram[index]

			if expectedEntry.statusCode != actualEntry.statusCode {
				test.Fatalf("Histogram item %d should have status code %d but is %d", index, expectedEntry.statusCode, actualEntry.statusCode)
			}

			if !float32sEqual(expectedEntry.probability, actualEntry.probability, .1) {
				test.Fatalf("Histogram item %d should have probability %f but is %f", index, expectedEntry.probability, actualEntry.probability)
			}

		}
	}

}
