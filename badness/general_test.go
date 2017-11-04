package badness

import (
	"testing"
)

func TestHeaderPresent(test *testing.T) {
	request := makeTestRequest()
	request.Header[CodeByHistogram] = []string{""}

	if !requestHasHeader(request, CodeByHistogram) {
		test.Fatal("requestHasHeader should be true for X-Response-Code")
	}
}

func TestHeaderAbsent(test *testing.T) {
	request := makeTestRequest()

	if requestHasHeader(request, CodeByHistogram) {
		test.Fatal("requestHasHeader should be false for X-Response-Code")
	}
}
