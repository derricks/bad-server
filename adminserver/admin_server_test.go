package adminserver

import (
	"net/http/httptest"
	"testing"
)

func TestUpdateHeaders(test *testing.T) {
	resetDefaultHeaders()

	request := httptest.NewRequest("POST", "/headers", nil)
	recorder := httptest.NewRecorder()

	setHeaders := map[string][]string{
		"X-Generate-Random": []string{"1000"},
		"X-Random-Json":     []string{"response_template=[returnObject]:100;returnObject=author/authorObject", "authorObject=id/int,name/string"},
	}

	request.Header = setHeaders
	updateDefaultHeaders(recorder, request)
	response := recorder.Result()

	for header, headerValue := range setHeaders {
		responseValue, found := response.Header[header]
		if !found {
			test.Fatalf("key %s should have been in response but was not", header)
		}

		if len(responseValue) != len(headerValue) {
			test.Fatalf("value for %s should be the same length as test value. expected %d was %d", header, len(headerValue), len(responseValue))
		}
	}
}

func TestGetDefaultHeaders(test *testing.T) {
	resetDefaultHeaders()
	request := httptest.NewRequest("POST", "/headers", nil)
	recorder := httptest.NewRecorder()
	setHeaders := map[string][]string{
		"X-Generate-Random": []string{"1000"},
		"X-Random-Json":     []string{"response_template=[returnObject]:100;returnObject=author/authorObjec;authorObject=id/int,name/string"},
	}
	request.Header = setHeaders
	updateDefaultHeaders(recorder, request)

	getHeaders := GetCurrentHeaders()
	for testHeader, expectedValue := range setHeaders {
		if actualValue, present := getHeaders[testHeader]; !present {
			test.Fatalf("Expected header %s in response. Was not present.", testHeader)

			if len(expectedValue) != len(actualValue) {
				test.Fatalf("Values for %s are different lengths. Expected %d got %d", testHeader, len(expectedValue), len(actualValue))
			}
		}
	}
}

func resetDefaultHeaders() {
	request := httptest.NewRequest("POST", "/headers", nil)
	response := httptest.NewRecorder()
	clearDefaultHeaders(response, request)
}
