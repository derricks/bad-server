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

		tempReader, err := getInitialLatencyAffector(request, strings.NewReader(""))
		affector := tempReader.(initialLatency)
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

type noiseExpectation struct {
	frequency float64
	err       error
}

func TestNoiseAffectorConstruction(test *testing.T) {
	expectations := map[string]noiseExpectation{
		"":            noiseExpectation{0.0, errors.New("blank string should yield error")},
		"17.2":        noiseExpectation{.172, nil},
		"not a float": noiseExpectation{0.0, errors.New("unparseable float should yield error")},
	}

	for headerValue, expectation := range expectations {
		request := makeTestRequest()
		request.Header[AddNoise] = []string{headerValue}

		tempReader, err := getNoiseAffector(request, strings.NewReader(""))
		affector := tempReader.(noiseAffector)
		if expectation.err == nil && err != nil {
			test.Fatalf("Header %s yielded unexpected error %v", headerValue, err)
		}

		if expectation.err != nil && err == nil {
			test.Fatalf("Header %s should have produced an error but did not: %v", headerValue, expectation.err)
		}

		if !float64sEqual(expectation.frequency, affector.noiseFrequency, .01) {
			test.Fatalf("Header %s should yield %f but actually yielded %f", headerValue, expectation.frequency, affector.noiseFrequency)
		}
	}
}

func TestNoiseAffector(test *testing.T) {
	testString := "testing"
	reader := strings.NewReader(testString)
	testBytes := []byte(testString)

	request := makeTestRequest()
	request.Header[AddNoise] = []string{"100.0"}

	tempReader, _ := getNoiseAffector(request, reader)
	affector := tempReader.(noiseAffector)

	output := make([]byte, 1024)
	bytesRead, _ := affector.Read(output)

	matches := 0
	// verify that each normal character has been replaced
	for index, normalCharacter := range testBytes {
		if output[index] == normalCharacter {
			matches++
		}
	}

	// while it's still possible for this to fail because two characters
	// are randomly set to themselves, the probability is 1:(256 * 256)
	if matches >= 2 {
		test.Fatalf("High number (%d) of mismatches between input %s and mutated version %s", matches, testString, string(output[0:bytesRead]))
	}
}
