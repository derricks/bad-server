package badness

// Functions for getting Readers that generate response bodies
import (
	"io"
	"math/rand"
	"net/http"
)

const RequestBodyIsResponse = "X-Request-Body-As-Response"

func buildBodyGenerator(generator io.Reader) ResponseHandler {

	return func(response http.ResponseWriter) error {
		buf := make([]byte, 1024)

		bytesRead, readErr := generator.Read(buf)
		for bytesRead > 0 {
			// per go docs:
			// Callers should always process the n > 0 bytes returned before considering the error err.
			// Doing so correctly handles I/O errors that happen after reading some bytes and also both of the allowed EOF behaviors.
			_, writeErr := response.Write(buf[0:bytesRead])
			if readErr != nil {
				return readErr
			}
			if writeErr != nil {
				return writeErr
			}

			bytesRead, readErr = generator.Read(buf)
		}
		return nil
	}
}

const GenerateRandomResponse = "X-Generate-Random"

type randomBodyGenerator struct {
	bytesToDeliver int
	bytesSoFar     int
}

func (generator *randomBodyGenerator) Read(buffer []byte) (int, error) {
	if generator.bytesSoFar < generator.bytesToDeliver {
		var bytesRead int
		if (generator.bytesToDeliver - generator.bytesSoFar) >= len(buffer) {
			// we have to deliver more than the buffer size, so
			// just fill it up
			bytesRead, _ = rand.Read(buffer)
		} else {
			// only fill up the number of bytes left
			bytesRead, _ = rand.Read(buffer[0 : generator.bytesToDeliver-generator.bytesSoFar])

		}
		generator.bytesSoFar = generator.bytesSoFar + bytesRead
		return bytesRead, nil
	}
	// no more bytes to read
	return 0, io.EOF
}

func newRandomBodyGenerator(bytesToDeliver int) *randomBodyGenerator {
	return &randomBodyGenerator{bytesToDeliver, 0}
}
