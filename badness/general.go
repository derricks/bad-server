package badness

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type ResponseHandler func(response http.ResponseWriter) error

// GetResponsePipeline returns an appropriately ordered
// slice of badness functions (based on request headers) that can be applied to a ResponseWriter.
// functions take a ResponseWriter as an argument.
func GetResponsePipeline(request *http.Request) []ResponseHandler {
	pipeline := make([]ResponseHandler, 0)

	// proxies circumvent the normal header/body building portions of the pipeline because
	// it pre-empts other headers and follows a different path
	if requestHasHeader(request, ProxyRequest) {
		proxy := buildProxyResponse(request)
		pipeline = append(pipeline, proxy.buildProxyHeaderGenerator())

		affector, err := getResponseAffector(request, proxy.getProxyReader())
		if err != nil {
			pipeline = []ResponseHandler{generateBadResponseHandler(fmt.Sprintf("Could not get affector: %v", err))}
			return pipeline
		}
		pipeline = append(pipeline, buildBodyGenerator(affector))
		pipeline = append(pipeline, proxy.buildProxyCloser())

	} else {
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
			pipeline = []ResponseHandler{generateBadResponseHandler(fmt.Sprintf("Could not get affector: %v", err))}
		}
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
	} else if requestHasHeader(request, RandomJson) {
		// gather up all the values for the header into one string
		allHeaderValues := request.Header[RandomJson]
		templateInput, err := normalizeJsonTemplateParameters(allHeaderValues)

		var generator jsonElementGenerator
		if err != nil {
			generator = newErrorGenerator(fmt.Sprintf("Could not process input %v", err))
		} else {
			generator, err = createJsonTemplate(templateInput)
			if err != nil {
				generator = newErrorGenerator(fmt.Sprintf("Could not parse input for generator %v", err))
			}
		}

		reader, writer := io.Pipe()
		go func() {
			generator.generate(writer)
			writer.Close()
		}()
		return reader
	} else {
		return strings.NewReader("")
	}
}

const responseTemplateKey = "response_template="

// normalizeJsonTemplateParameters ensures that the X-Random-Json header values
// are laid out into the right format for being parsed by the json_template package
// it returns an error if there's not a "response_template" parameter
func normalizeJsonTemplateParameters(headerValues []string) (string, error) {
	rawTemplateString := strings.Join(headerValues, ";")
	// split them apart for sorting
	rawFields := strings.Split(rawTemplateString, ";")
	sort.Sort(templateDefinitions(rawFields))

	if !strings.HasPrefix(rawFields[0], responseTemplateKey) {
		return "", fmt.Errorf("No response_template key in header")
	}

	templateInput := strings.Join(rawFields, ";")
	// and then remove the response_template key to prevent it from affecting the parser
	templateInput = strings.Replace(templateInput, "response_template=", "", 1)
	return templateInput, nil
}

type templateDefinitions []string

func (defs templateDefinitions) Len() int {
	return len(defs)
}
func (defs templateDefinitions) Swap(a, b int) {
	defs[a], defs[b] = defs[b], defs[a]
}
func (defs templateDefinitions) Less(a, b int) bool {
	// if a starts with "response_template=", it is less
	return strings.HasPrefix(defs[a], "response_template=")
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
