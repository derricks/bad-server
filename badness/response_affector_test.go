package badness

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type initialLatencyExpectation struct {
	waitDuration time.Duration
	errorExpectation
}

func buildLatencyExpectation(duration time.Duration, err error) initialLatencyExpectation {
	return initialLatencyExpectation{duration, errorExpectation{err}}
}

func TestInitialLatencyConstruction(test *testing.T) {
	expectations := map[string]initialLatencyExpectation{
		"":             buildLatencyExpectation(time.Duration(0)*time.Nanosecond, errors.New("expecting error about empty string")),
		"500":          buildLatencyExpectation(time.Duration(500)*time.Millisecond, nil),
		"30s":          buildLatencyExpectation(time.Duration(30)*time.Second, nil),
		"notParseable": buildLatencyExpectation(time.Duration(0)*time.Nanosecond, errors.New("expecting error about unparseable string")),
	}

	for waitString, expectation := range expectations {
		request := makeTestRequest()
		request.Header[PauseBeforeStart] = []string{waitString}

		tempReader, err := getInitialLatencyAffector(request, strings.NewReader(""))
		affector := tempReader.(initialLatency)
		checkErrorExpectation(waitString, expectation.errorExpectation, err, test)

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
	errorExpectation
}

func buildNoiseExpectation(frequency float64, err error) noiseExpectation {
	return noiseExpectation{frequency, errorExpectation{err}}
}

func TestNoiseAffectorConstruction(test *testing.T) {
	expectations := map[string]noiseExpectation{
		"":            buildNoiseExpectation(0.0, errors.New("blank string should yield error")),
		"17.2":        buildNoiseExpectation(.172, nil),
		"not a float": buildNoiseExpectation(0.0, errors.New("unparseable float should yield error")),
	}

	for headerValue, expectation := range expectations {
		request := makeTestRequest()
		request.Header[AddNoise] = []string{headerValue}

		tempReader, err := getNoiseAffector(request, strings.NewReader(""))
		affector := tempReader.(noiseAffector)
		checkErrorExpectation(fmt.Sprintf("Header %s", headerValue), expectation.errorExpectation, err, test)

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
