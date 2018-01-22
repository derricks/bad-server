package badness

import (
	"testing"
)

func TestHeaderPresent(test *testing.T) {
	request := makeTestRequest()
	request.Header[CodeByHistogram] = []string{""}

	if !requestHasHeader(request, CodeByHistogram) {
		test.Fatal("requestHasHeader should be true for X-Response-Code")
	}
}

func TestHeaderAbsent(test *testing.T) {
	request := makeTestRequest()

	if requestHasHeader(request, CodeByHistogram) {
		test.Fatal("requestHasHeader should be false for X-Response-Code")
	}
}

type normalizeJsonTemplateTest struct {
	inputs        []string
	output        string
	errorExpected bool
}

func TestNormalizeJsonTemplate(test *testing.T) {
	tests := []normalizeJsonTemplateTest{
		normalizeJsonTemplateTest{[]string{"response_template=bookshelf;bookshelf=books/[book]:10;;book=title/string,pages/int,isbn/string"}, "bookshelf;bookshelf=books/[book]:10;;book=title/string,pages/int,isbn/string", false},
		normalizeJsonTemplateTest{[]string{"bookshelf;bookshelf=books/[book]:10;;book=title/string,pages/int,isbn/string"}, "", true},
		normalizeJsonTemplateTest{[]string{"response_template=bookshelf;bookshelf=books/[book]:10", "book=title/string,pages/int,isbn/string"}, "bookshelf;bookshelf=books/[book]:10;book=title/string,pages/int,isbn/string", false},
		normalizeJsonTemplateTest{[]string{"bookshelf;bookshelf=books/[book]:10", "book=title/string,pages/int,isbn/string"}, "", true},
	}
	for _, templateTest := range tests {
		actual, err := normalizeJsonTemplateParameters(templateTest.inputs)

		if err != nil && !templateTest.errorExpected {
			test.Fatalf("Unexpected error parsing %v: %v", templateTest.inputs, err)
		}

		if err == nil && templateTest.errorExpected {
			test.Fatalf("Expected an error for %v but did not get one", templateTest.inputs)
		}

		if templateTest.output != actual {
			test.Errorf("Expected %s for %v, but got %s", templateTest.output, templateTest.inputs, actual)
		}
	}
}
