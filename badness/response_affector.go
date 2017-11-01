package badness

import (
	"errors"
	"fmt"
	"io"
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
func getInitialLatencyAffector(request *http.Request, reader io.Reader) (initialLatency, error) {
	waitString := request.Header[PauseBeforeStart][0]
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
