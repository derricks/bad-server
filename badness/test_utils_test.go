package badness

import (
	"testing"
)

// generic struct for expectations that include errors
// likely to mostly be used within another struct
type errorExpectation struct {
	err error
}

// a refactored method to handle the common task of checking returned errors against
// expectations within a unit test
func checkErrorExpectation(prefix string, expectation errorExpectation, actual error, test *testing.T) {
	if expectation.err != nil && actual == nil {
		test.Fatalf("%s did not cause expected error: %v", prefix, expectation.err)
	}

	if expectation.err == nil && actual != nil {
		test.Fatalf("%s produced unexpected error: %v", prefix, actual)
	}
}
