package badness

import (
	"fmt"
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

	histogramItems := buildHistogram(request.Header[CodeByHistogram])
	if len(histogramItems) == 0 {
		return func(response http.ResponseWriter) error {
			response.WriteHeader(http.StatusBadRequest)
			return nil
		}
	}

	return func(response http.ResponseWriter) error {
		buckets := statusCodeHistogramToHistogramBuckets(histogramItems)
		random := rand.Float64()
		bucket := bucketForProbability(random, buckets)
		if bucket == bucketNotFound {
			log.Printf("Could not find any bucket for %f. Do the probabilities add up to 1?", random)
			response.WriteHeader(http.StatusBadRequest)
		} else {
			response.WriteHeader(histogramItems[bucket].statusCode)
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

	var totalProbability float64

	parsedHeaders := parseHeadersWithKeyValues(headerValues, ",")
	for headerKey, headerValue := range parsedHeaders {
		// reconstruct the key=value string from the header since that's what
		// the helper function takes
		histogramValue := fmt.Sprintf("%s=%s", headerKey, headerValue)
		entry, err := parseHistogramHeader(histogramValue)

		if err != nil {
			log.Printf("Skipping bad histogram value %s: %v\n", histogramValue, err)
			continue
		}

		if float64sEqual(0.0, entry.probability, .1) {
			zeroProbs = append(zeroProbs, entry)
		} else {
			histogram = append(histogram, entry)
			totalProbability += entry.probability
		}
	}

	if len(zeroProbs) > 0 {
		// fill in entries with 0 probabilities
		perEmptyProbability := (float64(1.0) - totalProbability) / float64(len(zeroProbs))
		for _, entry := range zeroProbs {
			entry.probability = perEmptyProbability
			histogram = append(histogram, entry)
			totalProbability += perEmptyProbability
		}
	}

	if !float64sEqual(1.0, totalProbability, .1) {
		log.Printf("Probabilities do not add up to 1: %f\n", totalProbability)
		return []statusCodeHistogramEntry{}
	}

	sort.Sort(statusCodeHistogram(histogram))
	return histogram
}

// statusCodeHistogramToHistogramBuckets converts a slice
// of statusCodeHistogramEntry structs to a slice of
// histogramBucket structs. This is to get around go's
// type restrictions.
func statusCodeHistogramToHistogramBuckets(statusCodeEntries []statusCodeHistogramEntry) []histogramBucket {
	returnSlice := make([]histogramBucket, 0)
	for _, statusCodeEntry := range statusCodeEntries {
		returnSlice = append(returnSlice, statusCodeEntry.histogramBucket)
	}
	return returnSlice
}

type statusCodeHistogramEntry struct {
	statusCode int
	histogramBucket
}

// parseHistogramHeader converts a key=value pair to a histogram entry.
// The parser takes "key=value" and "key", in which case value is
// set to 0. If key cant be converted to an int, the function returns
// an error. If value can't be parsed as a float, the value will
// be set to 0
func parseHistogramHeader(keyValue string) (statusCodeHistogramEntry, error) {
	emptyEntry := statusCodeHistogramEntry{0, histogramBucket{0.0}}
	keyValueFields := strings.Split(keyValue, "=")

	statusCode, err := strconv.Atoi(keyValueFields[0])
	if err != nil {
		return emptyEntry, err
	}

	probability := 0.0
	if len(keyValueFields) > 1 && keyValueFields[1] != "" {
		probability, err = strconv.ParseFloat(keyValueFields[1], 64)
		if err != nil {
			return emptyEntry, err
		}
	}
	return statusCodeHistogramEntry{statusCode, histogramBucket{probability / 100.0}}, nil
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
