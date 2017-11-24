package badness

import (
	"testing"
)

type findBucketExpect struct {
	testNumber float64
	bucketMax  float64
}

func TestFindingHistogramBucket(test *testing.T) {
	hist := []histogramBucket{
		histogramBucket{.2},
		histogramBucket{.3},
		histogramBucket{.5},
	}

	expectations := []findBucketExpect{
		findBucketExpect{.1, .2},
		findBucketExpect{.25, .3},
		findBucketExpect{.7, .5},
	}

	for index, expect := range expectations {
		bucket := bucketForProbability(expect.testNumber, hist)
		if !float64sEqual(expect.bucketMax, hist[bucket].probability, .01) {
			test.Fatalf("Test: %d, Expected bucket max to be %f was %f", index, expect.bucketMax, hist[bucket].probability)
		}
	}

	bucket := bucketForProbability(2.0, hist)
	if bucket != -1 {
		test.Fatalf("Expected -1 for unavailable index, got %d", bucket)
	}
}
