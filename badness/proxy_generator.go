package badness

// ProxyRequest is a specialized response generator that passes along the request
// as-is to the host specified in the X-Proxy-To-Host header. That header is not passed along.
// The idea is that you can query existing web services but then add response affectors after the 
// fact.
import (
  "fmt"
  "net/http"
  "net/url"
)

const ProxyRequest = "X-Proxy-To-Host"

type proxyRequest struct {
  newHost string
  http.Request
}

// buidlProxyRequest uses the input request as a template to use for making
// a request to another service, and then returns a ResponseHandler that will
// write the response back to the client.
// The incoming request's ProxyRequest header is used to determine the host to talk to,
// but otherwise everything about the request is kept the same.
func buildProxyRequest(request *http.Request) ResponseHandler {
  newHost := getFirstHeaderValue(request, ProxyRequest)
  url, err := urlFromHostAndUrl(newHost, request.URL)
  if err != nil {
    return generateBadResponseHandler(fmt.Sprintf("Could not calculate URL: %v", err  ))
    // return a 400 and explanatory text (and a test)
  }

  // given the url, return a function that binds the url to make a request
  // and pipes the response from the server to the client of this program
  return func(_ http.ResponseWriter) error {return nil}
}

// urlFromHostAndUrl constructs a new URL by overlaying a parsed URL based on
// newHost onto oldUrl. For instance, if the newHost is https://abc.com,
// the new URL should have https as the scheme. But the newHost doesn't have a scheme,
// the Scheme in oldUrl will be used. Returns a new URL and an error
func urlFromHostAndUrl(newHost string, oldURL *url.URL) (*url.URL, error) {
  parsedUrl, err  := url.Parse(newHost)
  if err != nil {
    return nil, err
  }
  
  oldCopy := url.URL{
    Scheme: oldURL.Scheme,
    Opaque: oldURL.Opaque,
    User: oldURL.User,
    Host: oldURL.Host,
    Path: oldURL.Path,
    ForceQuery: oldURL.ForceQuery,
    RawQuery: oldURL.RawQuery,
    Fragment: oldURL.Fragment,
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