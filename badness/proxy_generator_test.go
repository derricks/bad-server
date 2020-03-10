package badness

import (
	"net/url"
	"testing"
)

type urlFromUrlTest struct {
	baseUrl        string
	overlayUrl     string
	expectScheme   string
	expectHost     string
	expectPath     string
	expectRawQuery string
	expectError    bool
}

func TestUrlFromHostAndUrl(test *testing.T) {

	tests := []urlFromUrlTest{
		urlFromUrlTest{"https://www.google.com", "http://www.yahoo.com", "http", "www.yahoo.com", "", "", false},
	}

	for index, testCase := range tests {
		parsedBase, _ := url.Parse(testCase.baseUrl)
		newURL, err := urlFromHostAndUrl(testCase.overlayUrl, parsedBase)
		if err != nil && !testCase.expectError {
			test.Fatalf("Test %d: Unexpected error: %v", index, err)
		}

		if err == nil && testCase.expectError {
			test.Fatalf("Test %d: Expected error but didn't get one", index)
		}

		if err != nil { // no sense evaluating the rest of the test cases
			continue
		}

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
