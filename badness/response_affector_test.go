package badness

import (
	"errors"
	"fmt"
	"io"
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
		affector := tempReader.(*initialLatency)
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

func TestInitialLatencyEffect(test *testing.T) {
	request := makeTestRequest()
	request.Header[PauseBeforeStart] = []string{"300ms"}
	embeddedReader := strings.NewReader("longish string")
	buf := make([]byte, 10) // force two reads
	latencyAffector, err := getInitialLatencyAffector(request, embeddedReader)
	if err != nil {
		test.Fatalf("Unexpected error %v", err)
	}

	start := time.Now()
	// might as well check that bytes are being returned properly
	bytesRead, err := latencyAffector.Read(buf)
	if bytesRead != 10 || err != nil {
		test.Fatalf("Expected to read 10 bytes without error, got %d bytes and %v", bytesRead, err)
	}
	bytesRead, err = latencyAffector.Read(buf)
	if bytesRead != 4 && err != nil {
		test.Fatalf("Expected to read 4 bytes with no error, got %d and %v", bytesRead, err)
	}
	bytesRead, err = latencyAffector.Read(buf)
	if bytesRead != 0 && err != io.EOF {
		test.Fatalf("Expected 0 bytes and EOF, got %d bytes and %v", bytesRead, err)
	}
	timeTaken := time.Since(start)
	if timeTaken.Seconds() > 0.5 {
		test.Fatalf("Test should have taken ~300ms: took %f", timeTaken.Seconds())
	}
}

// test that durations can be extracted out of
// an affector's histogram
type durationFromHistogramExpect struct {
	probability float64
	duration    time.Duration
}

func TestFindRandomizerInRandomLatencyAffector(test *testing.T) {
	probabilities := []float64{.1, .5, .4}
	durations := []time.Duration{time.Duration(300) * time.Millisecond, time.Duration(1) * time.Second, time.Duration(1) * time.Minute}
	histogram := make([]lagginessRandomizer, 3)
	for index, _ := range probabilities {
		histogram = append(histogram, lagginessRandomizer{histogramBucket{probabilities[index]}, 0, durations[index]})
	}

	affector := randomLagginessAffector{strings.NewReader("not used"), histogram}

	expectations := []durationFromHistogramExpect{
		durationFromHistogramExpect{.05, time.Duration(300) * time.Millisecond},
		durationFromHistogramExpect{.3, time.Duration(1) * time.Second},
		durationFromHistogramExpect{.9, time.Duration(1) * time.Minute},
		durationFromHistogramExpect{2.0, time.Duration(1) * time.Minute},
	}

	for _, expect := range expectations {
		randomizer := affector.randomizerFromHistogram(expect.probability)
		if expect.duration.Nanoseconds() != randomizer.upTo.Nanoseconds() {
			test.Fatalf("Expected duration of %v for probability %f, got %v", expect.duration, expect.probability, randomizer.upTo)
		}
	}
}

func TestDurationBetween(test *testing.T) {
	lower := time.Duration(100) * time.Millisecond
	higher := time.Duration(300) * time.Millisecond

	between := randomDurationBetween(lower, higher)
	if between.Nanoseconds() < lower.Nanoseconds() || between.Nanoseconds() > higher.Nanoseconds() {
		test.Fatalf("Expected value between %v and %v but got %v", lower, higher, between)
	}

	between = randomDurationBetween(higher, lower)
	if between.Nanoseconds() < lower.Nanoseconds() || between.Nanoseconds() > higher.Nanoseconds() {
		test.Fatalf("Expected value between %v and %v but got %v", higher, lower, between)
	}

}

type lagginessRandomizerExpect struct {
	key, value  string
	probability float64
	from        time.Duration
	to          time.Duration
}

func TestLagginessRandomizerConstruction(test *testing.T) {
	expectations := []lagginessRandomizerExpect{
		lagginessRandomizerExpect{"100", "30.0", .3, time.Duration(0), time.Duration(100) * time.Millisecond},
		lagginessRandomizerExpect{"", "", 0.0, time.Duration(0), time.Duration(0)},
		lagginessRandomizerExpect{"100-300", "20.0", .2, time.Duration(100) * time.Millisecond, time.Duration(300) * time.Millisecond},
		lagginessRandomizerExpect{"100ms-1s", "10.0", .1, time.Duration(100) * time.Millisecond, time.Duration(1) * time.Second},
		lagginessRandomizerExpect{"3s", "", 0.0, time.Duration(0), time.Duration(3) * time.Second},
	}
	for index, expect := range expectations {
		actual := lagginessRandomizerFromKeyValue(expect.key, expect.value)
		if !float64sEqual(expect.probability, actual.probability, .01) {
			test.Fatalf("Test %d: expected %f for probability but got %f", index, expect.probability, actual.probability)
		}

		if expect.from.Nanoseconds() != actual.from.Nanoseconds() {
			test.Fatalf("Test %d: expected %v for from, got %v", index, expect.from, actual.from)
		}

		if expect.to.Nanoseconds() != actual.upTo.Nanoseconds() {
			test.Fatalf("Test %d: expected %v for to, got %v", index, expect.to, actual.upTo)
		}
	}
}

type randomLagginessExpect struct {
	header    string
	histogram []lagginessRandomizer
}

func TestGetRandomLagginesAffector(test *testing.T) {
	expectations := []randomLagginessExpect{
		randomLagginessExpect{"1s=20", []lagginessRandomizer{lagginessRandomizer{histogramBucket{.2}, time.Duration(0), time.Duration(1) * time.Second}}},
		randomLagginessExpect{"10ms=20,10ms-50ms=30,1s", []lagginessRandomizer{
			lagginessRandomizer{histogramBucket{.2}, time.Duration(0), time.Duration(10) * time.Millisecond},
			lagginessRandomizer{histogramBucket{.3}, time.Duration(10) * time.Millisecond, time.Duration(20) * time.Millisecond},
			lagginessRandomizer{histogramBucket{.5}, time.Duration(0), time.Duration(1) * time.Second},
		},
		},
	}

	for expectIndex, expectation := range expectations {
		request := makeTestRequest()
		request.Header[RandomLaggyResponse] = []string{expectation.header}

		affector, _ := getRandomLagginessAffector(request, strings.NewReader("not used"))

		if len(affector.(randomLagginessAffector).histogram) != len(expectation.histogram) {
			test.Fatalf("Test case %d: Affector histogram %v and expected histogram %v are not the same length", expectIndex, affector.(randomLagginessAffector).histogram, expectation.histogram)
		}

		// test all the lagginess randomizers
		for histogramIndex, actualRandomizer := range affector.(randomLagginessAffector).histogram {
			expectedRandomizer := expectation.histogram[histogramIndex]

			if !float64sEqual(expectedRandomizer.probability, actualRandomizer.probability, .01) {
				test.Fatalf("Test %d, histogram entry %d: expected %f got %f", expectIndex, histogramIndex, expectedRandomizer.probability, actualRandomizer.probability)
			}

		}
	}
}

type lagginessReadExpect struct {
	bufSize    int
	minElapsed time.Duration
	maxElapsed time.Duration
}

// verify that the minimum amount of predicted lag happens for given readers
func TestRandomLagginessRead(test *testing.T) {
	expectations := []lagginessReadExpect{
		// uses a smaller-than-lagginess-chunk-sized buffer
		lagginessReadExpect{35, time.Duration(150) * time.Millisecond, time.Duration(300) * time.Millisecond},
		// uses a larger buffer but reads should still be chunked up
		lagginessReadExpect{1000, time.Duration(150) * time.Millisecond, time.Duration(300) * time.Millisecond},
	}

	for _, expect := range expectations {
		start := time.Now()
		request := makeTestRequest()
		request.Header[RandomLaggyResponse] = []string{"50ms-100ms"}
		affector, _ := getRandomLagginessAffector(request, strings.NewReader("a fairly long string to test multiple reads with short buffers"))
		buf := make([]byte, expect.bufSize)

		_, err := affector.Read(buf)
		for err == nil {
			test.Log("Reading bytes")
			_, err = affector.Read(buf)
		}

		elapsed := time.Since(start)
		if elapsed.Nanoseconds() < expect.minElapsed.Nanoseconds() {
			test.Fatalf("Random lagginess didn't take long enough. Expected %v, got %v", expect.minElapsed, elapsed)
		}
		if elapsed.Nanoseconds() > expect.maxElapsed.Nanoseconds() {
			test.Fatalf("Random lagginess took too long. Expected %v, got %v", expect.maxElapsed, elapsed)
		}
	}
}
