package badness

// ProxyRequest is a specialized response generator that passes along the request
// as-is to the host specified in the X-Proxy-To-Host header. That header is not passed along.
// The idea is that you can query existing web services but then add response affectors after the
// fact.
import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const ProxyRequest = "X-Proxy-To-Host"

type proxiedResponse struct {
	response  *http.Response
	errorText string
}

// buildHeaderGenerator uses the cached response to generate a ResponseHandler
// function that can be used in the response pipeline
func (proxy proxiedResponse) buildProxyHeaderGenerator() ResponseHandler {
	return func(response http.ResponseWriter) error {
		if proxy.errorText != "" {
			response.WriteHeader(http.StatusBadRequest)
		} else {
			for header, values := range proxy.response.Header {
				response.Header()[header] = values
			}
			response.WriteHeader(proxy.response.StatusCode)
		}
		// nothing in this part returns an error
		return nil
	}
}

// getReader returns the Body of the response.
func (proxy proxiedResponse) getProxyReader() io.Reader {
	if proxy.errorText == "" {
		return proxy.response.Body
	} else {
		return strings.NewReader(proxy.errorText)
	}
}

// this ResponseHandler is put into the pipeline to ensure the response body
// is closed after all the response affectors come into play
func (proxy proxiedResponse) buildProxyCloser() ResponseHandler {
	return func(response http.ResponseWriter) error {
		if proxy.errorText == "" {
			proxy.response.Body.Close()
		}
		return nil
	}
}

// buildProxyResponse uses the input request as a template to use for making
// a request to another service, and then returns a proxiedResponse object that has ResponseHandlers
// for the various steps in the pipeline
func buildProxyResponse(request *http.Request) *proxiedResponse {
	newHost := getFirstHeaderValue(request, ProxyRequest)
	url, err := urlFromHostAndUrl(newHost, request.URL)
	if err != nil {
		return &proxiedResponse{&http.Response{}, fmt.Sprintf("Could not calculate URL: %v", err)}
	}

	newRequest, err := http.NewRequest(request.Method, url.String(), request.Body)
	if err != nil {
		return &proxiedResponse{&http.Response{}, fmt.Sprintf("%v", err)}
	}

	for header, values := range request.Header {
		newRequest.Header[header] = values
	}
	newRequest.ContentLength = request.ContentLength

	client := &http.Client{}
	response, err := client.Do(newRequest)
	if err != nil {
		return &proxiedResponse{&http.Response{}, fmt.Sprintf("%v", err)}
	}
	return &proxiedResponse{response, ""}
}

// urlFromHostAndUrl constructs a new URL by overlaying a parsed URL based on
// newHost onto oldUrl. For instance, if the newHost is https://abc.com,
// the new URL should have https as the scheme. But the newHost doesn't have a scheme,
// the Scheme in oldUrl will be used. Returns a new URL and an error
func urlFromHostAndUrl(newHost string, oldURL *url.URL) (*url.URL, error) {
	parsedUrl, err := url.Parse(newHost)
	if err != nil {
		return nil, err
	}

	oldCopy := url.URL{
		Scheme:     oldURL.Scheme,
		Opaque:     oldURL.Opaque,
		User:       oldURL.User,
		Host:       oldURL.Host,
		Path:       oldURL.Path,
		ForceQuery: oldURL.ForceQuery,
		RawQuery:   oldURL.RawQuery,
		Fragment:   oldURL.Fragment,
	}

	if parsedUrl.Scheme != oldCopy.Scheme {
		oldCopy.Scheme = parsedUrl.Scheme
	}

	if parsedUrl.Opaque != oldCopy.Opaque {
		oldCopy.Opaque = parsedUrl.Opaque
	}

	if !anyAreNil(parsedUrl.User, oldCopy.User) {
		if parsedUrl.User.String() != oldCopy.User.String() {
			oldCopy.User = parsedUrl.User
		}
	}

	if parsedUrl.Host != oldCopy.Host {
		oldCopy.Host = parsedUrl.Host
	}

	if parsedUrl.Path != oldCopy.Path {
		oldCopy.Path = parsedUrl.Path
	}

	if parsedUrl.ForceQuery != oldCopy.ForceQuery {
		oldCopy.ForceQuery = parsedUrl.ForceQuery
	}

	if parsedUrl.RawQuery != oldCopy.RawQuery {
		oldCopy.RawQuery = parsedUrl.RawQuery
	}

	if parsedUrl.Fragment != oldCopy.Fragment {
		oldCopy.Fragment = parsedUrl.Fragment
	}

	return &oldCopy, nil
}
