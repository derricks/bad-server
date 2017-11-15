package badness

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ----------------------- initial latency affector ---------------------------
const PauseBeforeStart = "X-Pause-Before-Response-Start"

// code that affects how body generators are affected before
// being sent in a response.

type initialLatency struct {
	reader      io.Reader
	initialWait time.Duration
	hasSlept    bool
}

func (affector *initialLatency) Read(buffer []byte) (int, error) {
	if !affector.hasSlept {
		time.Sleep(affector.initialWait)
		affector.hasSlept = true
	}

	return affector.reader.Read(buffer)
}

// getInitialLatencyAffector uses the request header to build an initialLatency affector around reader.
// the values for the header can be passed as a straight int, in which case
// they are interpreted as milliseconds, or a golang Duration string (e.g., 100ms).
// If multiple values of the header are defined, only the first is used
func getInitialLatencyAffector(request *http.Request, reader io.Reader) (io.Reader, error) {
	waitString := getFirstHeaderValue(request, PauseBeforeStart)
	if waitString == "" {
		return &initialLatency{nil, time.Duration(0) * time.Nanosecond, false}, errors.New(fmt.Sprintf("No value defined for %s header. Pass an integer or a duration string", PauseBeforeStart))
	}

	// if field is an integer, use that
	millis, err := strconv.Atoi(waitString)
	if err == nil {
		return &initialLatency{reader, time.Duration(millis) * time.Millisecond, false}, nil
	}

	// field was not an int. try and parse as duration
	duration, err := time.ParseDuration(waitString)
	if err == nil {
		return &initialLatency{reader, duration, false}, nil
	} else {
		return &initialLatency{nil, time.Duration(0) * time.Nanosecond, false}, err
	}
}

// -------------------------- random noise affector ----------------------------
const AddNoise = "X-Add-Noise"

// a response affector that randomly changes bytes.
// noiseFrequency sets the randomness level to use for modifying bytes
type noiseAffector struct {
	reader         io.Reader
	noiseFrequency float64
}

func (affector noiseAffector) Read(buffer []byte) (int, error) {
	// first populate the buffer. per docs, process bytes first
	bytesRead, err := affector.reader.Read(buffer)

	for index, _ := range buffer[0:bytesRead] {
		randomFloat := rand.Float64()
		if randomFloat < affector.noiseFrequency {
			buffer[index] = byte(rand.Intn(256))
		}
	}

	return bytesRead, err
}

// getNoiseAffector returns a noiseAffector using the percentage value in
// the header. If the header is defined multiple times, only the first is used.
func getNoiseAffector(request *http.Request, reader io.Reader) (io.Reader, error) {
	frequencyString := getFirstHeaderValue(request, AddNoise)
	if frequencyString == "" {
		return noiseAffector{}, errors.New(fmt.Sprintf("%s requires a float parameter", AddNoise))
	}

	percentage, err := strconv.ParseFloat(frequencyString, 64)
	if err != nil {
		return noiseAffector{}, err
	}
	return noiseAffector{reader: reader, noiseFrequency: percentage / 100.0}, nil
}

// ------------------------- random lagginess affector -------------------------
const RandomLaggyResponse = "X-Random-Delays"
const lagginessBufferSize = 50

type lagginessRandomizer struct {
	probability float64
	from        time.Duration
	upTo        time.Duration
}

type randomizerSet []lagginessRandomizer

func (randomizers randomizerSet) Len() int {
	return len(randomizers)
}
func (randomizers randomizerSet) Swap(left, right int) {
	randomizers[left], randomizers[right] = randomizers[right], randomizers[left]
}
func (randomizers randomizerSet) Less(left, right int) bool {
	return randomizers[left].probability < randomizers[right].probability
}

type randomLagginessAffector struct {
	reader    io.Reader
	histogram []lagginessRandomizer
}

func (affector randomLagginessAffector) Read(buf []byte) (int, error) {
	// read up to lagginessBufferSize from reader
	// the reason to not just read directly into buf is we want to add lag throughout
	// the transmission; buf might be big enough normally to hold the entire message
	// so we wouldn't get to add lag

	// if buf is shorter than lagginessBufferSize, use that length. Otherwise, the
	// underlying reader will advance farther than we want
	var tempBuf []byte
	if lagginessBufferSize < len(buf) {
		tempBuf = make([]byte, lagginessBufferSize)
	} else {
		tempBuf = make([]byte, len(buf))
	}

	randomizer := affector.randomizerFromHistogram(rand.Float64())
	bytesRead, err := affector.reader.Read(tempBuf)
	time.Sleep(randomDurationBetween(randomizer.from, randomizer.upTo))
	// put into the read buffer
	for index, curByte := range tempBuf {
		buf[index] = curByte
	}
	return bytesRead, err
}

// randomizerFromHistogram will use the passed-in random number to find
// a lagginessRandomizer that meets the criteria
func (affector randomLagginessAffector) randomizerFromHistogram(random float64) lagginessRandomizer {
	accumulatedProbability := float64(0.0) // track how many histogram entries we've gone through

	for _, histogramEntry := range affector.histogram {
		accumulatedProbability += histogramEntry.probability
		if accumulatedProbability > random {
			return histogramEntry
		}
	}

	// if no match happened, return the last entry
	return affector.histogram[len(affector.histogram)-1]
}

// randomDurationBetween takes two durations and returns
// a random duration that happens between the two.
// durations can be passed in either order
func randomDurationBetween(from, upTo time.Duration) time.Duration {
	diff := int(math.Abs(float64(upTo.Nanoseconds() - from.Nanoseconds())))
	random := rand.Intn(diff)
	if upTo.Nanoseconds() > from.Nanoseconds() {
		return time.Duration(int(from.Nanoseconds())+random) * time.Nanosecond
	} else {
		return time.Duration(int(upTo.Nanoseconds())+random) * time.Nanosecond
	}
}

func getRandomLagginessAffector(request *http.Request, reader io.Reader) (io.Reader, error) {
	keyValuePairs := parseHeadersWithKeyValues(request.Header[RandomLaggyResponse], ",")
	histogram := make([]lagginessRandomizer, 0)
	// for entries without a probability. these will be emended later
	zeroProbs := make([]lagginessRandomizer, 0)
	totalProbability := 0.0

	for key, value := range keyValuePairs {
		lagginess := lagginessRandomizerFromKeyValue(key, value)
		if float64sEqual(lagginess.probability, 0.0, .01) {
			zeroProbs = append(zeroProbs, lagginess)
		} else {
			histogram = append(histogram, lagginess)
			totalProbability += lagginess.probability
		}
	}

	if len(zeroProbs) > 0 {
		// fill in entries with 0 probabilities
		perEmptyProbability := (float64(1.0) - totalProbability) / float64(len(zeroProbs))
		for _, entry := range zeroProbs {
			entry.probability = perEmptyProbability
			histogram = append(histogram, entry)
			totalProbability += perEmptyProbability
		}
	}

	if !float64sEqual(1.0, totalProbability, .1) {
		log.Printf("Probabilities do not add up to 100: %f\n", totalProbability)
	}

	// sort for testing predictability
	sort.Sort(randomizerSet(histogram))
	return randomLagginessAffector{reader, histogram}, nil
}

func lagginessRandomizerFromKeyValue(key, value string) lagginessRandomizer {
	var probability float64
	var err error
	var toDuration = time.Duration(0)
	var fromDuration = time.Duration(0)

	if value != "" {
		probability, err = strconv.ParseFloat(value, 64)
		if err != nil {
			log.Printf("Invalid probability value: %s. Setting to 0.0", value)
		}
	}

	durations := strings.Split(key, "-")
	if len(durations) == 1 {
		// only one duration, so that's the max time span
		toDuration, err = stringToDuration(durations[0])
		if err != nil {
			log.Printf("Invalid duration: using 0, %v", err)
		}
	} else if len(durations) > 1 {
		fromDuration, err = stringToDuration(durations[0])
		if err != nil {
			log.Printf("Invalid duration: using 0, %v", err)
		}

		toDuration, err = stringToDuration(durations[1])
		if err != nil {
			log.Printf("Invalid duration: using 0, %v", err)
		}
	}

	return lagginessRandomizer{probability / 100.0, fromDuration, toDuration}
}
