package badness

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"

	"bad-server/badness/json_template"
)

const RandomJson = "X-Random-Json"

var stringCharacters = [52]byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}

type jsonElementGenerator interface {
	generate(writer io.Writer) (int, error)
}

// createJsonTemplate takes an input string and parses it with the json_template tools.
// The resulting DataDeclaration set is converted to a nested tree of json generators as needed.
// It is assumed that the first data definition in the input string is the root of the json.
// Calling methods will need to make sure of that state
func createJsonTemplate(input string) (jsonElementGenerator, error) {
	parser := json_template.NewParserWithString(input)
	template, err := parser.ParseTemplate()
	if err != nil {
		return nil, err
	}

	if len(template.Declarations) == 0 {
		return nil, fmt.Errorf("No json template definitions found")
	}

	return generatorFromDataDeclaration(template, template.Declarations[0])
}

// generatorFromDataDeclaration takes a json_template DataDeclaration and converts it
// to the appropriate json generator.
func generatorFromDataDeclaration(template *json_template.Template, declaration json_template.DataDeclaration) (jsonElementGenerator, error) {
	switch declaration.(type) {
	case json_template.PrimitiveDataType:
		return primitiveGeneratorFromDataType(declaration.(json_template.PrimitiveDataType))
	case json_template.ArrayDataType:
		return arrayGeneratorFromDataType(template, declaration.(json_template.ArrayDataType))
	case json_template.KeyValueDataType:
		return keyValueGeneratorFromDataType(template, declaration.(json_template.KeyValueDataType))
	case json_template.KeyNameDataType:
		// ensure parser has an object of the given type
		key := declaration.(json_template.KeyNameDataType).Literal
		object, found := template.CustomTypes[key]
		if !found {
			return nil, fmt.Errorf("Unknown data type: %s", key)
		}
		return objectGeneratorFromDataType(template, object.(json_template.ObjectDataType))
	case json_template.EnumStringDataType:
		return enumGeneratorFromStringEnumDataType(declaration.(json_template.EnumStringDataType))
	case json_template.EnumIntDataType:
		return enumGeneratorFromIntEnumDataType(declaration.(json_template.EnumIntDataType))
	case json_template.EnumFloatDataType:
		return enumGeneratorFromFloatEnumDataType(declaration.(json_template.EnumFloatDataType))
	default:
		return nil, fmt.Errorf("Unknown type: %v", declaration)
	}
}

func objectGeneratorFromDataType(template *json_template.Template, object json_template.ObjectDataType) (jsonElementGenerator, error) {
	keyValues := make([]keyValueGenerator, 0, len(object.Members))
	for _, keyValueData := range object.Members {
		generator, err := keyValueGeneratorFromDataType(template, keyValueData)
		if err != nil {
			return nil, err
		}
		keyValues = append(keyValues, generator.(keyValueGenerator))
	}
	return newObjectGenerator(keyValues), nil
}

func keyValueGeneratorFromDataType(template *json_template.Template, declaration json_template.KeyValueDataType) (jsonElementGenerator, error) {
	valueGenerator, err := generatorFromDataDeclaration(template, declaration.Value)
	if err != nil {
		return nil, err
	}
	return newKeyValueGenerator(declaration.Key, valueGenerator), nil
}

func arrayGeneratorFromDataType(template *json_template.Template, declaration json_template.ArrayDataType) (jsonElementGenerator, error) {
	nestedGenerator, err := generatorFromDataDeclaration(template, declaration.NestedType)
	if err != nil {
		return nil, err
	}
	return newArrayGenerator(declaration.Length, nestedGenerator), nil
}

func primitiveGeneratorFromDataType(declaration json_template.PrimitiveDataType) (jsonElementGenerator, error) {
	switch declaration.TokenLiteral() {
	case "string":
		return newRandomStringGenerator(), nil
	case "int":
		return newIntGenerator(10000), nil
	case "bool":
		return newBooleanGenerator(), nil
	case "increment":
		return newIncrementGenerator(1), nil
	case "float":
		return newFloatGenerator(1.0), nil
	default:
		return nil, fmt.Errorf("Unknown primitive type: %s", declaration.TokenLiteral())
	}
}

func enumGeneratorFromStringEnumDataType(declaration json_template.EnumStringDataType) (jsonElementGenerator, error) {
	return newStringFromSetGenerator(declaration.Values), nil
}

func enumGeneratorFromIntEnumDataType(declaration json_template.EnumIntDataType) (jsonElementGenerator, error) {
	return newIntFromSetGenerator(declaration.Values), nil
}

func enumGeneratorFromFloatEnumDataType(declaration json_template.EnumFloatDataType) (jsonElementGenerator, error) {
	return newFloatFromSetGenerator(declaration.Values), nil
}

type stringFromSetGenerator struct {
	values []string
}

func (generator stringFromSetGenerator) generate(writer io.Writer) (int, error) {
	index := rand.Intn(len(generator.values))
	return newFixedStringGenerator(generator.values[index]).generate(writer)
}

func newStringFromSetGenerator(values []string) jsonElementGenerator {
	return &stringFromSetGenerator{values}
}

type intFromSetGenerator struct {
	values []int
}

func (generator intFromSetGenerator) generate(writer io.Writer) (int, error) {
	index := rand.Intn(len(generator.values))
	return newFixedIntGenerator(generator.values[index]).generate(writer)
}

func newIntFromSetGenerator(values []int) jsonElementGenerator {
	return &intFromSetGenerator{values}
}

type floatFromSetGenerator struct {
	values []float64
}

func (generator floatFromSetGenerator) generate(writer io.Writer) (int, error) {
	index := rand.Intn(len(generator.values))
	return fixedFloatGenerator{generator.values[index]}.generate(writer)
}
func newFloatFromSetGenerator(values []float64) jsonElementGenerator {
	return floatFromSetGenerator{values}
}

// writeGeneratorsInList concatenates the output of the generators into the writer with the given character
func writeGeneratorsInList(generators []jsonElementGenerator, writer io.Writer, joinString string) (int, error) {
	bytesTotal := 0
	delimiter := []byte(joinString)
	for index, generator := range generators {
		bytes, err := generator.generate(writer)
		bytesTotal += bytes
		if err != nil {
			return bytesTotal, err
		}

		if index != len(generators)-1 {
			bytes, err = writer.Write(delimiter)
			bytesTotal += bytes
			if err != nil {
				return bytesTotal, err
			}
		}
	}
	return bytesTotal, nil
}

func newErrorGenerator(message string) jsonElementGenerator {
	keyValue := newKeyValueGenerator("error", newFixedStringGenerator(message))

	return newObjectGenerator([]keyValueGenerator{keyValue.(keyValueGenerator)})
}

// useful for generating empty arrays. Does nothing to the writer
type noItemGenerator struct{}

func (generator noItemGenerator) generate(writer io.Writer) (int, error) {
	return 0, nil
}
func newNoItemGenerator() jsonElementGenerator {
	return noItemGenerator{}
}

// ---   Generate random strings of a given length --
var quote = []byte{'"'}

type randomStringGenerator struct {
	length int
}

func (generator randomStringGenerator) generate(writer io.Writer) (int, error) {
	buffer := make([]byte, generator.length+2)
	buffer[0] = '"'

	for current := 0; current < generator.length; current++ {
		randomIndex := rand.Intn(len(stringCharacters))
		buffer[current+1] = stringCharacters[randomIndex]
	}
	// terminal quote is the full length - 1
	buffer[generator.length+2-1] = '"'

	bytesWritten, err := writer.Write(buffer[0 : generator.length+2])
	return bytesWritten, err
}

func newRandomStringGenerator() jsonElementGenerator {
	return randomStringGenerator{30}
}

// ---- Generate constant strings -----
type fixedStringGenerator struct {
	fixedString string
}

func (generator fixedStringGenerator) generate(writer io.Writer) (int, error) {
	bytesTotal, err := writer.Write(quote)

	if err != nil {
		return bytesTotal, err
	}
	bytes, err := writer.Write([]byte(generator.fixedString))
	bytesTotal += bytes
	if err != nil {
		return bytesTotal, err
	}
	bytes, err = writer.Write(quote)
	bytesTotal += bytes
	return bytesTotal, err
}

func newFixedStringGenerator(fixedString string) jsonElementGenerator {
	return fixedStringGenerator{fixedString}
}

// --------- bool generator
type booleanGenerator struct{}

func (generator booleanGenerator) generate(writer io.Writer) (int, error) {
	// generates true/false randomly
	boolean := "true"
	if rand.Float32() > 0.5 {
		boolean = "false"
	}
	return writer.Write([]byte(boolean))
}

func newBooleanGenerator() jsonElementGenerator {
	return &booleanGenerator{}
}

// --------- int generator. generates a random number up to the max
type intGenerator struct {
	maxNumber int
}

func (generator intGenerator) generate(writer io.Writer) (int, error) {
	intToWrite := rand.Intn(generator.maxNumber)
	nestedGenerator := fixedIntGenerator{intToWrite}
	return nestedGenerator.generate(writer)
}

func newIntGenerator(maxNum int) jsonElementGenerator {
	return intGenerator{maxNum}
}

// ----------------- float generates a random float up to the max
type floatGenerator struct {
	maxNumber float64
}

func (generator floatGenerator) generate(writer io.Writer) (int, error) {
	return fixedFloatGenerator{rand.Float64() * generator.maxNumber}.generate(writer)
}

func newFloatGenerator(max float64) jsonElementGenerator {
	return floatGenerator{max}
}

type fixedFloatGenerator struct {
	value float64
}

func (generator fixedFloatGenerator) generate(writer io.Writer) (int, error) {
	floatAsString := fmt.Sprintf("%f", generator.value)
	return writer.Write([]byte(floatAsString))
}

// ---------- generates a fixed int
type fixedIntGenerator struct {
	number int
}

func (generator fixedIntGenerator) generate(writer io.Writer) (int, error) {
	intString := strconv.Itoa(generator.number)
	return writer.Write([]byte(intString))
}

func newFixedIntGenerator(number int) fixedIntGenerator {
	return fixedIntGenerator{number}
}

// ---------- increment generator. starting at a value, counts up
type incrementGenerator struct {
	current int
}

func (generator *incrementGenerator) generate(writer io.Writer) (int, error) {
	intString := strconv.Itoa(generator.current)
	bytesWritten, err := writer.Write([]byte(intString))
	// no matter what, increment the counter. Even if there's an error, it's more expected
	// that the number would go up versus staying the same.
	generator.current++
	return bytesWritten, err
}

func newIncrementGenerator(start int) jsonElementGenerator {
	return &incrementGenerator{start}
}

// ---- Array generator --------
var leftBracket = []byte{'['}
var rightBracket = []byte{']'}
var comma = []byte{','}

type arrayGenerator struct {
	length            int
	generatorToRepeat jsonElementGenerator
}

func (generator arrayGenerator) generate(writer io.Writer) (int, error) {
	bytesTotal, err := writer.Write(leftBracket)
	if err != nil {
		return bytesTotal, err
	}

	// make a list of generators to conform to writeGenerators spec
	generators := make([]jsonElementGenerator, 0)
	for index := 0; index < generator.length; index++ {
		generators = append(generators, generator.generatorToRepeat)
	}

	bytes, err := writeGeneratorsInList(generators, writer, ",")
	bytesTotal += bytes
	if err != nil {
		return bytesTotal, err
	}

	bytes, err = writer.Write(rightBracket)
	bytesTotal += bytes
	return bytesTotal, err
}

func newArrayGenerator(length int, generator jsonElementGenerator) jsonElementGenerator {
	return arrayGenerator{length, generator}
}

// -------------------- keyvalue generator -------------------
type keyValueGenerator struct {
	key   fixedStringGenerator
	value jsonElementGenerator
}

var colon = []byte{':'}

func (generator keyValueGenerator) generate(writer io.Writer) (int, error) {
	bytesTotal, err := generator.key.generate(writer)
	if err != nil {
		return bytesTotal, err
	}

	bytes, err := writer.Write(colon)
	bytesTotal += bytes
	if err != nil {
		return bytesTotal, err
	}

	bytes, err = generator.value.generate(writer)
	bytesTotal += bytes
	return bytesTotal, err
}

func newKeyValueGenerator(key string, value jsonElementGenerator) jsonElementGenerator {
	return keyValueGenerator{newFixedStringGenerator(key).(fixedStringGenerator), value}
}

// -------------- object generator
var leftBrace = []byte{'{'}
var rightBrace = []byte{'}'}

type objectGenerator struct {
	generators []keyValueGenerator
}

func (generator objectGenerator) generate(writer io.Writer) (int, error) {
	bytesTotal, err := writer.Write(leftBrace)
	if err != nil {
		return bytesTotal, err
	}

	// convert generator list  to jsonElementGenerator
	generators := make([]jsonElementGenerator, 0)
	for _, generator := range generator.generators {
		generators = append(generators, jsonElementGenerator(generator))
	}

	bytes, err := writeGeneratorsInList(generators, writer, ",")
	bytesTotal += bytes
	if err != nil {
		return bytesTotal, err
	}

	bytes, err = writer.Write(rightBrace)
	bytesTotal += bytes
	return bytesTotal, err
}

func newObjectGenerator(generators []keyValueGenerator) jsonElementGenerator {
	return objectGenerator{generators}
}
