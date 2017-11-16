package badness

import (
	"errors"
	"fmt"
	"testing"
)

type histogramParseExpect struct {
	statusCode  int
	probability float32
	errorExpectation
}

func buildHistogramParseExpect(statusCode int, probability float32, err error) histogramParseExpect {
	return histogramParseExpect{statusCode, probability, errorExpectation{err}}
}

func TestHistogramEntryParsing(test *testing.T) {
	expectations := map[string]histogramParseExpect{
		"500=33.2": buildHistogramParseExpect(500, 33.2, nil),

		"": buildHistogramParseExpect(0, 0.0, errors.New("empty string should yield error")),

		"500": buildHistogramParseExpect(500, 0.0, nil),

		"500=": buildHistogramParseExpect(500, 0.0, nil),

		"500=test": buildHistogramParseExpect(0, 0.0, errors.New("500=test should yield an error")),

		"x": buildHistogramParseExpect(0, 0.0, errors.New("x should yield an error")),

		"x=y": buildHistogramParseExpect(0, 0.0, errors.New("x=y should yield parse errors")),
	}

	for parseString, expectation := range expectations {
		entry, err := parseHistogramHeader(parseString)

		checkErrorExpectation(fmt.Sprintf("Header %s", parseString), expectation.errorExpectation, err, test)

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
		histogramExpect{[]string{"500=60", "200=40"}, []statusCodeHistogramEntry{statusCodeHistogramEntry{200, 40.0}, statusCodeHistogramEntry{500, 60.0}}},
		histogramExpect{[]string{"500=60,200"}, []statusCodeHistogramEntry{statusCodeHistogramEntry{200, 40.0}, statusCodeHistogramEntry{500, 60.0}}},
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
