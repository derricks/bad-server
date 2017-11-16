package badness

import (
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

const CodeByHistogram = "X-Response-Code-Histogram"

/* Functions for setting response codes */

// generate a ResponseHandler that handles the request's passed-in histogram
//
// If there's an error with the histogram, a ResponseHandler that generates
// bad request will be returned instead
func generateHistogramStatusCode(request *http.Request) ResponseHandler {

	histogram := buildHistogram(request.Header[CodeByHistogram])
	if len(histogram) == 0 {
		return func(response http.ResponseWriter) error {
			response.WriteHeader(http.StatusBadRequest)
			return nil
		}
	}

	return func(response http.ResponseWriter) error {
		probability := rand.Float32() * float32(100.0)
		entryCumulative := float32(0.0)

		// go through each entry. if the cumulative entry
		// is greater than the random probability
		// use that as the status code
		for _, entry := range histogram {
			entryCumulative += entry.probability
			if entryCumulative > probability {
				response.WriteHeader(entry.statusCode)
				return nil
			}
		}
		return nil
	}
}

// buildHistogram uses headerValues to build up a histogram
// that can be used to generate status codes.
// strings that don't yield errors are skipped (and logged); any entries
// with a probability of 0 will divide up any remaining probabilities
func buildHistogram(headerValues []string) []statusCodeHistogramEntry {
	histogram := make([]statusCodeHistogramEntry, 0)
	// their probabilities will be filled in afterwards
	zeroProbs := make([]statusCodeHistogramEntry, 0)

	var totalProbability float32

	for _, headerValue := range headerValues {
		for _, histogramValue := range strings.Split(headerValue, ",") {
			entry, err := parseHistogramHeader(histogramValue)

			if err != nil {
				log.Printf("Skipping bad histogram value %s: %v\n", histogramValue, err)
				continue
			}

			if float32sEqual(0.0, entry.probability, .1) {
				zeroProbs = append(zeroProbs, entry)
			} else {
				histogram = append(histogram, entry)
				totalProbability += entry.probability
			}
		}
	}

	if len(zeroProbs) > 0 {
		// fill in entries with 0 probabilities
		perEmptyProbability := (float32(100.0) - totalProbability) / float32(len(zeroProbs))
		for _, entry := range zeroProbs {
			entry.probability = perEmptyProbability
			histogram = append(histogram, entry)
			totalProbability += perEmptyProbability
		}
	}

	if !float32sEqual(100.0, totalProbability, .1) {
		log.Printf("Probabilities do not add up to 100: %f\n", totalProbability)
		return []statusCodeHistogramEntry{}
	}

  sort.Sort(statusCodeHistogram(histogram))
	return histogram
}

type statusCodeHistogramEntry struct {
	statusCode  int
	probability float32
}

// parseHistogramHeader converts a key=value pair to a histogram entry.
// The parser takes "key=value" and "key", in which case value is
// set to 0. If key cant be converted to an int, the function returns
// an error. If value can't be parsed as a float, the value will
// be set to 0
func parseHistogramHeader(keyValue string) (statusCodeHistogramEntry, error) {
	keyValueFields := strings.Split(keyValue, "=")

	statusCode, err := strconv.Atoi(keyValueFields[0])
	if err != nil {
		return statusCodeHistogramEntry{0, 0.0}, err
	}

	probability := 0.0
	if len(keyValueFields) > 1 && keyValueFields[1] != "" {
		probability, err = strconv.ParseFloat(keyValueFields[1], 32)
		if err != nil {
			return statusCodeHistogramEntry{0, 0.0}, err
		}
	}
	return statusCodeHistogramEntry{statusCode: statusCode, probability: float32(probability)}, nil
}

type statusCodeHistogram []statusCodeHistogramEntry
func (histogram statusCodeHistogram) Len() int {
	return len(histogram)
}
func (histogram statusCodeHistogram) Swap(left, right int) {
	histogram[left], histogram[right] = histogram[right], histogram[left]
}
func (histogram statusCodeHistogram) Less(left, right int) bool {
	return histogram[left].probability < histogram[right].probability
}