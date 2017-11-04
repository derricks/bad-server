package badness

import (
  "log"
	"net/http"
  "strings"
)

// code for setting headers in a repsonse

// buildHeaderSetter returns a ResponseHandler that will
// set the given header to the given values
func buildHeaderSetter(header string, values []string) ResponseHandler {
  
  return func(response http.ResponseWriter) error {
    response.Header()[header] = values
    return nil
  }
}

const ForceHeader = "X-Return-Header"

// buildForcedHeader collates all the X-Return-Header values in request and returns 
// a slice of headerSetters that will set appropriate headers in the response
func buildForcedHeaders(request *http.Request) []ResponseHandler {
  headerSetters := make([]ResponseHandler,0)
  
  forcedHeaders, headerPresent := request.Header[ForceHeader]
  if !headerPresent {
    return headerSetters
  }
  headersToForce := collateForcedHeaders(forcedHeaders)

  for header, values := range headersToForce {
    headerSetters = append(headerSetters, buildHeaderSetter(header, values))
  }
  return headerSetters  
}

// collateForcedHeaders uses all the forcedHeader fields from the request
// to build a new map of forced header keys and values
// for instance: 
//    X-Return-Header: Test-Header-1: q
//    X-Return-Header: Test-Header-1: z
//    X-Return-Header: Test-Header-2: y
//  will create a map where Test-Header-1 has two values, and Test-Header-2 has one
func collateForcedHeaders(forcedHeaders []string) map[string][]string {
  
  collatedHeaders := make(map[string][]string)
  for _, forcedHeader := range forcedHeaders {
    forcedHeaderFields := strings.SplitN(forcedHeader, ":", 2)
    
    if len(forcedHeaderFields) < 2 {
      log.Printf("Invalid header field %s: http headers must be key: value", forcedHeader)
      continue
    }
    
    forcedHeaderKey := forcedHeaderFields[0]
    forcedHeaderValue := strings.TrimSpace(forcedHeaderFields[1])
		
    currentValues, wasSet := collatedHeaders[forcedHeaderKey]
    if wasSet {
      currentValues = append(currentValues, forcedHeaderValue)
      collatedHeaders[forcedHeaderKey] = currentValues
    } else {
      collatedHeaders[forcedHeaderKey] = []string{forcedHeaderValue}
    }
  }
  return collatedHeaders
}

