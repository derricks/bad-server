package badness

import (
	"testing"
)

func TestEqualFloats(test *testing.T) {
	if !float32sEqual(3.3, 3.3, .1) {
		test.Fatal("3.3 and 3.3 should be equal but are not")
	}
}

func TestUnequalFloats(test *testing.T) {
	if float32sEqual(3.3, 3.4, .1) {
		test.Fatal("3.3 and 3.4 should not be equal but are")
	}
}

func TestEqualFloat64s(test *testing.T) {
	if !float64sEqual(3.3, 3.3, .1) {
		test.Fatal("3.3 and 3.3 should be equal but are not")
	}
}

func TestUnequalFloat64ss(test *testing.T) {
	if float64sEqual(3.3, 3.4, .1) {
		test.Fatal("3.3 and 3.4 should not be equal but are")
	}
}
