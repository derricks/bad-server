package badness

import (
  "net/url"
  "testing"
)

type urlFromUrlTest struct {
   baseUrl string
   overlayUrl string
   expectScheme string
   expectHost string
   expectPath string
   expectRawQuery string
   expectError bool
}
func TestUrlFromHostAndUrl(test *testing.T) {
  
  tests := []urlFromUrlTest{
    urlFromUrlTest{"https://www.google.com", "www.yahoo.com", "https", "www.abc.com", "", "", false},
    urlFromUrlTest{"12345", "www.yahoo.com", "https", "www.yahoo.com", "", "", true},
  }
  
  for index, testCase := range tests {
    parsedBase, err := url.Parse(testCase.baseUrl)
    if err != nil && !testCase.expectError {
      test.Fatalf("Test %d: Unexpected error: %v", index, err)
    }
    
    if err != nil && testCase.expectError {
      test.Fatalf("Test %d: Expected error but didn't get one", index)
    }
      
    newURL, _ := urlFromHostAndUrl(testCase.overlayUrl, parsedBase)
    compareFunc := func(expected, actual, fieldName string) {
      if expected != actual {
        test.Errorf("Test case %d: Expected %s of %s but got %s", index, fieldName, expected, actual)  
      }
    }
    compareFunc(testCase.expectScheme, newURL.Scheme, "scheme")
    compareFunc(testCase.expectHost, newURL.Host, "host")
    compareFunc(testCase.expectPath, newURL.Path, "path")
    compareFunc(testCase.expectRawQuery, newURL.RawQuery, "raw query")
  }
}