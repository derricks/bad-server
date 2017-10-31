package badness

import (
	"net/http"
	"strings"
)

type ResponseHandler func(response http.ResponseWriter) error

// GetResponsePipeline returns an appropriately ordered
// slice of badness functions (based on request headers) that can be applied to a ResponseWriter.
// functions take a ResponseWriter as an argument.
func GetResponsePipeline(request *http.Request) []ResponseHandler {
	pipeline := make([]ResponseHandler, 0)
	// generators that generate status codes go first
	if requestHasHeader(request, CodeByHistogram) {
		pipeline = append(pipeline, generateHistogramStatusCode(request))
	}

	var bodyGenerator ResponseHandler

	// build up body generators, for now with hard-coding
	if requestHasHeader(request, RequestBodyIsResponse) {
		bodyGenerator = buildBodyGenerator(request.Body)
	} else {
		bodyGenerator = buildBodyGenerator(strings.NewReader(""))
	}

	pipeline = append(pipeline, bodyGenerator)
	return pipeline
}

// requestHasHeader returns true if the given request has the handler, false if not
func requestHasHeader(request *http.Request, header string) bool {
	_, found := request.Header[header]
	return found
}
