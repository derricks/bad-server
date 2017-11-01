package badness

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type initialLatencyExpectation struct {
	waitDuration time.Duration
	err          error
}

func TestInitialLatencyConstruction(test *testing.T) {
	expectations := map[string]initialLatencyExpectation{
		"":             initialLatencyExpectation{time.Duration(0) * time.Nanosecond, errors.New("expecting error about empty string")},
		"500":          initialLatencyExpectation{time.Duration(500) * time.Millisecond, nil},
		"30s":          initialLatencyExpectation{time.Duration(30) * time.Second, nil},
		"notParseable": initialLatencyExpectation{time.Duration(0) * time.Nanosecond, errors.New("expecting error about unparseable string")},
	}

	for waitString, expectation := range expectations {
		request := makeTestRequest()
		request.Header[PauseBeforeStart] = []string{waitString}

		affector, err := getInitialLatencyAffector(request, strings.NewReader(""))
		if err == nil && expectation.err != nil {
			test.Fatalf("Expected error and got none for string %s : %v", waitString, expectation.err)
		}

		if err != nil && expectation.err == nil {
			test.Fatalf("Got unexpected error for string %s: %v", waitString, err)
		}

		if affector.initialWait.Nanoseconds() != expectation.waitDuration.Nanoseconds() {
			test.Fatalf("Expected %v for initial wait, got %v, string: %s", expectation.waitDuration, affector.initialWait, waitString)
		}

		if affector.hasSlept {
			test.Fatalf("Affector hasSlept is true, should be false")
		}
	}
}
