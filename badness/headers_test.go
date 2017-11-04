package badness

import (
  "net/http/httptest"
  "sort"
  "testing"
)

type forcedHeaderExpectation struct {
  forcedHeaders []string
  collatedHeaders map[string][]string
}
func buildForcedHeaderExpectation(headers []string, collation map[string][]string) forcedHeaderExpectation {
  return forcedHeaderExpectation{headers, collation}
}

func TestCollatedForcedHeaders(test *testing.T) {
  expectations := []forcedHeaderExpectation{
    // empty slice => empty map
    buildForcedHeaderExpectation([]string{}, map[string][]string{}),
    buildForcedHeaderExpectation([]string{"Test-1: x"}, map[string][]string{"Test-1": []string{"x"}}),
    buildForcedHeaderExpectation([]string{"Test-1: x", "Test-1: y", "Test-2: z"}, 
      map[string][]string{
        "Test-1": []string{"x","y"},
        "Test-2": []string{"z"},
      },
    ),
  }
  
  for _, expectation := range expectations {
    collation := collateForcedHeaders(expectation.forcedHeaders)
    if len(collation) != len(expectation.collatedHeaders) {
      test.Fatalf("Collated headers from %v is not the same length as expected", collation)
    }
    
    for expectedKey, expectedValues := range expectation.collatedHeaders {
      actualValues, valuesPresent := collation[expectedKey]
      if !valuesPresent {
        test.Fatalf("Expected key %s to have values.", expectedKey)
      }
      
      diffIndex, matched := compareStringSlices(expectedValues, actualValues)
      if !matched {
        test.Fatalf("Expected %v did not match %v at index %d", expectedValues, actualValues, diffIndex)
      }
      
    }    
  }
}

// compareStringSlices checks two slices for equality and returns an integer indicating
// where they differ and whether or not they matched up.
// note: this does an in-place sort on the slices, so it's destructive
func compareStringSlices(expected, actual []string) (index int, matched bool) {
  // compare the two slices
  if len(expected) != len(actual) {
    return 0, false
  }
  
  sort.Strings(actual)
  sort.Strings(expected)
  for index, _ := range expected {
    if expected[index] != actual[index] {
      return index, false
    }
  }
  return len(expected), true
}

func TestResponseGetsForcedHeaders(test *testing.T) {
  request := makeTestRequest()
  request.Header[ForceHeader] = []string{"Content-Type: application/json", "Content-Type: text/xml", "My-Header: 300"}
  headerSetters := buildForcedHeaders(request)
  response := httptest.NewRecorder()
  
  for _, setter := range headerSetters {
    setter(response)
  }
  
  contentTypeExpect := []string{"application/json", "text/xml"}
  myHeaderExpect := []string{"300"}
  
  diffIndex, match := compareStringSlices(response.Header()["Content-Type"], contentTypeExpect)
  if !match {
    test.Fatalf("Content-Type: %v differed from %v at %d", contentTypeExpect, response.Header()["Content-Type"], diffIndex)
  }
  
  diffIndex, match = compareStringSlices(response.Header()["My-Header"], myHeaderExpect)
  if !match {
    test.Fatalf("My-Header: %v differed from %v at %d", contentTypeExpect, response.Header()["Content-Type"], diffIndex)
  }
}