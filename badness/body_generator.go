package badness

// Functions for getting Readers that generate response bodies
import (
	"io"
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
