package badness

const bucketNotFound = -1

type histogramBucket struct {
	probability float64
}

// bucketForProbability takes a float64 and returns the first index of the bucket
// in the slice whose cumulative value is greater than float64
// in other words, given buckets of .1, .3, .6, a float64 value of .45 would
// return 2, because adding .1 and .3 is still less than the value.
// a value of .2 would return 1 because it's greater than .1 but less than
// the cumulative value of .1 and .3
//
// the goal is to find which bucket this falls into
// the function returns an index into the slice rather than the
// object itself, since the slice of histogramBucket has to be constructed
// from other structs that embed that.
// returns -1 if no value was found.
func bucketForProbability(probability float64, buckets []histogramBucket) int {
	cumulativeProbability := float64(0.0)

	for index, bucket := range buckets {
		cumulativeProbability += bucket.probability
		if cumulativeProbability > probability {
			return index
		} else {
			cumulativeProbability += bucket.probability
		}
	}
	return bucketNotFound
}
