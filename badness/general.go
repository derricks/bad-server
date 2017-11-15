package badness

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ResponseHandler func(response http.ResponseWriter) error

// GetResponsePipeline returns an appropriately ordered
// slice of badness functions (based on request headers) that can be applied to a ResponseWriter.
// functions take a ResponseWriter as an argument.
func GetResponsePipeline(request *http.Request) []ResponseHandler {
	pipeline := make([]ResponseHandler, 0)

	pipeline = append(pipeline, getHeaderGenerators(request)...)
	// generators that generate status codes go first
	if requestHasHeader(request, CodeByHistogram) {
		pipeline = append(pipeline, generateHistogramStatusCode(request))
	}

	bodyGenerator := getBodyGenerator(request)
	affectedGenerator, err := getResponseAffector(request, bodyGenerator)
	if err == nil {
		pipeline = append(pipeline, buildBodyGenerator(affectedGenerator))
	} else {
		log.Printf("Could not create response affector: %v", err)
	}

	return pipeline
}

// getHeaderGenerators builds up a slice of ResponseHandlers based on headers
func getHeaderGenerators(request *http.Request) []ResponseHandler {
	responseHandlers := make([]ResponseHandler, 0)
	if requestHasHeader(request, ForceHeader) {
		forceHeaders := buildForcedHeaders(request)
		responseHandlers = append(responseHandlers, forceHeaders...)
	}
	return responseHandlers
}

// getBodyGenerator returns a Reader that will generate the body text
// based on settings in the request headers.
// currently we only support
func getBodyGenerator(request *http.Request) io.Reader {
	if requestHasHeader(request, RequestBodyIsResponse) {
		return request.Body
	} else if requestHasHeader(request, GenerateRandomResponse) {
		bodySizeField := getFirstHeaderValue(request, GenerateRandomResponse)
		bodySize, err := strconv.Atoi(bodySizeField)
		if err != nil {
			log.Printf("Could not convert body size for random data: %s", bodySizeField)
			return strings.NewReader("")
		}
		return newRandomBodyGenerator(bodySize)
	} else {
		return strings.NewReader("")
	}

}

type responseAffector func(request *http.Request, reader io.Reader) (io.Reader, error)

var headerToAffector = map[string]responseAffector{
	AddNoise:            getNoiseAffector,
	PauseBeforeStart:    getInitialLatencyAffector,
	RandomLaggyResponse: getRandomLagginessAffector,
}

// getResponseAffector uses the http request headers to decorate the given reader
// with appropriate affectors (things that affect the
// sending of the response regardless of the body)
func getResponseAffector(request *http.Request, reader io.Reader) (returnReader io.Reader, err error) {

	returnReader = reader

	for header, getter := range headerToAffector {
		if requestHasHeader(request, header) {
			returnReader, err = getter(request, returnReader)
			if err != nil {
				return nil, err
			}
		}
	}

	return returnReader, nil
}

// requestHasHeader returns true if the given request has the handler, false if not
func requestHasHeader(request *http.Request, header string) bool {
	_, found := request.Header[header]
	return found
}
