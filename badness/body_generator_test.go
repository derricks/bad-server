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

func TestRandomGeneratorLength(test *testing.T) {
	generator := newRandomBodyGenerator(700)
	var err error

	// make one shorter than buf to test multiple reads
	randomBuf1 := make([]byte, 500)
	bytesRead, _ := generator.Read(randomBuf1)
	if bytesRead != len(randomBuf1) {
		test.Fatalf("Should have read %d bytes from random generator but read %d", len(randomBuf1), bytesRead)
	}
	bytesRead, _ = generator.Read(randomBuf1)
	if bytesRead != 200 {
		test.Fatalf("Should have read %d bytes from random generator but read %d", 200, bytesRead)
	}
	bytesRead, err = generator.Read(randomBuf1)
	if err == nil {
		test.Fatalf("Did not get EOF error")
	}

	// a buf longer than randomBodyGenerator
	generator = newRandomBodyGenerator(700)
	randomBuf2 := make([]byte, 1000)
	bytesRead, _ = generator.Read(randomBuf2)
	if bytesRead != 700 {
		test.Fatalf("Expected 700 bytes from random body generator, got %d", bytesRead)
	}
	_, err = generator.Read(randomBuf2)
	if err == nil {
		test.Fatalf("Did not get expected error")
	}
}
