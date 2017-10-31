package badness

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseFromReader(test *testing.T) {

	testString := "testing string"

	bodyGeneratorFunction := buildBodyGenerator(strings.NewReader(testString))
	response := httptest.NewRecorder()
	bodyGeneratorFunction(response)
	responseBody := response.Body.String()

	if responseBody != testString {
		test.Fatalf("Response body should be %s was %s", testString, responseBody)
	}
}
