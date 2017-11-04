package badness

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const PauseBeforeStart = "X-Pause-Before-Response-Start"

// code that affects how body generators are affected before
// being sent in a response.

type initialLatency struct {
	reader      io.Reader
	initialWait time.Duration
	hasSlept    bool
}

func (affector initialLatency) Read(buffer []byte) (int, error) {
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
		return initialLatency{nil, time.Duration(0) * time.Nanosecond, false}, errors.New(fmt.Sprintf("No value defined for %s header. Pass an integer or a duration string", PauseBeforeStart))
	}

	// if field is an integer, use that
	millis, err := strconv.Atoi(waitString)
	if err == nil {
		return initialLatency{reader, time.Duration(millis) * time.Millisecond, false}, nil
	}

	// field was not an int. try and parse as duration
	duration, err := time.ParseDuration(waitString)
	if err == nil {
		return initialLatency{reader, duration, false}, nil
	} else {
		return initialLatency{nil, time.Duration(0) * time.Nanosecond, false}, err
	}
}

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
