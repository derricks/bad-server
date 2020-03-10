package badness

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
)

// expectations when there are not errors midstream
type jsonElementExpectation struct {
	generator     jsonElementGenerator
	expectedRegex string
}

func TestJsonElementGeneration(test *testing.T) {
	expects := []jsonElementExpectation{
		jsonElementExpectation{newFixedStringGenerator("test"), "^\"test\"$"},
		jsonElementExpectation{newArrayGenerator(0, newNoItemGenerator()), "\\[\\]"},
		jsonElementExpectation{newKeyValueGenerator("index", newFixedStringGenerator("value")), "\"index\":\"value\""},
		jsonElementExpectation{newArrayGenerator(1, newFixedStringGenerator("testing")), "\\[\"testing\"\\]"},
		jsonElementExpectation{newArrayGenerator(2, newFixedStringGenerator("hello")), "[\"hello\",\"hello\"]"},
		jsonElementExpectation{newBooleanGenerator(), "true|false"},
		jsonElementExpectation{newRandomStringGenerator(), "\"[a-zA-Z]{30}\""},
		jsonElementExpectation{newIntGenerator(10000), "[0-9]+"},
		jsonElementExpectation{newArrayGenerator(2, newIncrementGenerator(1)), "\\[1,2\\]"},
		jsonElementExpectation{newObjectGenerator([]keyValueGenerator{newKeyValueGenerator("key1", newIntGenerator(100)).(keyValueGenerator), newKeyValueGenerator("key2", newFixedStringGenerator("value")).(keyValueGenerator)}), "\\{\"key1\":[0-9]+,\"key2\":\"value\"}"},
		jsonElementExpectation{newErrorGenerator("no-error"), "\\{\"error\":\"no-error\"\\}"},
		jsonElementExpectation{newIntFromSetGenerator([]int{1, 2, 3}), "1|2|3"},
		jsonElementExpectation{newStringFromSetGenerator([]string{"a", "b", "c"}), "a|b|c"},
	}

	for index, expect := range expects {
		generatedString := generatedString(expect.generator)
		match, err := regexp.Match(expect.expectedRegex, []byte(generatedString))
		if err != nil {
			test.Fatalf("Regexp error: %v", err)
		}

		if !match {
			test.Errorf("Test %d: Expected %s but got %s", index, expect.expectedRegex, generatedString)
		}
	}
}

// a special writer that returns an error after the specified number
// of bytes has been reached
type interruptingWriter struct {
	interruptAt int
	totalBytes  int
	bytes.Buffer
}

func (writer *interruptingWriter) Write(bytes []byte) (int, error) {
	// we need to track the bytes for just this write,
	// as that's what we need to return to the client
	bytesThisWrite := 0
	for _, toWrite := range bytes {
		bytesWritten, _ := writer.Buffer.Write([]byte{toWrite})
		bytesThisWrite += bytesWritten
		writer.totalBytes += bytesWritten

		if writer.totalBytes == writer.interruptAt {
			return bytesThisWrite, fmt.Errorf("Stopping at %d bytes", writer.totalBytes)
		}
	}
	return bytesThisWrite, nil
}

type jsonWithInterruptExpectation struct {
	stopAt    int
	generator jsonElementGenerator
	// the regex of the string you'd expect from the interrupt
	expectedRegex string
}

// TestInterruptedJson verifies that jsonGenerators return appropriate bytesRead and errors
// in the face of underlying exceptions
func TestInterruptedJson(test *testing.T) {
	expects := []jsonWithInterruptExpectation{
		jsonWithInterruptExpectation{1, newFixedStringGenerator("test"), "^\"$"},
		jsonWithInterruptExpectation{3, newFixedStringGenerator("test"), "^\"te$"},
		jsonWithInterruptExpectation{6, newFixedStringGenerator("test"), "^\"test\"$"},
		jsonWithInterruptExpectation{1, newRandomStringGenerator(), "^\"$"},
		jsonWithInterruptExpectation{5, newRandomStringGenerator(), "^\"[a-zA-Z]{4}$"},
		jsonWithInterruptExpectation{32, newRandomStringGenerator(), "^\"[a-zA-Z]{30}\"$"},
		jsonWithInterruptExpectation{1, newBooleanGenerator(), "^(t|f)$"},
		jsonWithInterruptExpectation{4, newBooleanGenerator(), "^(true|fals)$"},
		jsonWithInterruptExpectation{1, newIntGenerator(1000000), "^[0-9]+$"},
		jsonWithInterruptExpectation{1, newArrayGenerator(2, newFixedStringGenerator("test")), "^\\[$"},
		jsonWithInterruptExpectation{8, newArrayGenerator(2, newFixedStringGenerator("test")), "^\\[\"test\","},
		jsonWithInterruptExpectation{10, newArrayGenerator(2, newFixedStringGenerator("test")), "^\\[\"test\",\"t$"},
		jsonWithInterruptExpectation{1, newKeyValueGenerator("key", newFixedStringGenerator("value")), "^\"$"},
		jsonWithInterruptExpectation{5, newKeyValueGenerator("key", newFixedStringGenerator("value")), "^\"key\"$"},
		jsonWithInterruptExpectation{6, newKeyValueGenerator("key", newFixedStringGenerator("value")), "^\"key\":$"},
		jsonWithInterruptExpectation{13, newKeyValueGenerator("key", newFixedStringGenerator("value")), "^\"key\":\"value\"$"},
		jsonWithInterruptExpectation{1, newObjectGenerator([]keyValueGenerator{newKeyValueGenerator("test", newFixedStringGenerator("value")).(keyValueGenerator)}), "\\{$"},
		jsonWithInterruptExpectation{7, newObjectGenerator([]keyValueGenerator{newKeyValueGenerator("test", newFixedStringGenerator("value")).(keyValueGenerator)}), "\\{\"test\"$"},
		jsonWithInterruptExpectation{8, newObjectGenerator([]keyValueGenerator{newKeyValueGenerator("test", newFixedStringGenerator("value")).(keyValueGenerator)}), "\\{\"test\":$"},
		jsonWithInterruptExpectation{15, newObjectGenerator([]keyValueGenerator{newKeyValueGenerator("test", newFixedStringGenerator("value")).(keyValueGenerator)}), "\\{\"test\":\"value\"$"},
		jsonWithInterruptExpectation{16, newObjectGenerator([]keyValueGenerator{newKeyValueGenerator("test", newFixedStringGenerator("value")).(keyValueGenerator)}), "\\{\"test\":\"value\"\\}$"},
	}

	for index, expect := range expects {
		writer := new(interruptingWriter)
		writer.interruptAt = expect.stopAt

		written, err := expect.generator.generate(writer)
		if err == nil {
			test.Errorf("Test case %d: Expected an error but got nil", index)
		}

		if written != expect.stopAt {
			test.Errorf("Test case %d: Expected %d bytes written but got %d", index, expect.stopAt, written)
		}

		generatedString := string(writer.Bytes()[0:written])
		match, err := regexp.Match(expect.expectedRegex, []byte(generatedString))
		if err != nil {
			test.Fatalf("Regexp error: %v", err)
		}

		if !match {
			test.Errorf("Test case %d: Expected %s but got %s", index, expect.expectedRegex, generatedString)
		}

	}
}

func TestTemplateDefinitionToJson(test *testing.T) {
	// template language strings to expected json regex
	tests := map[string]string{
		"string":                 "^\"[a-zA-Z]{30}\"$",
		"int":                    "^[0-9]+$",
		"bool":                   "^(true|false)$",
		"[string]:1":             "^\\[\"[a-zA-Z]{30}\"\\]",
		"test;test=title/string": "^{\"title\":\"[a-zA-Z]{30}\"}$",
	}

	for input, expected := range tests {
		generator, err := createJsonTemplate(input)
		if err != nil {
			test.Fatalf("Parsing input string failed: %v", err)
		}
		generated := generatedString(generator)
		regex := regexp.MustCompile(expected)

		if !regex.Match([]byte(generated)) {
			test.Errorf("Expected input %s to match regex %s, but was %s", input, expected, generated)
		}
	}
}

// Refactored method for generating a string from the generator's output
func generatedString(generator jsonElementGenerator) string {
	var buffer bytes.Buffer
	written, _ := generator.generate(&buffer)
	return string(buffer.Bytes()[0:written])
}
