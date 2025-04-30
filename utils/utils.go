package utils

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2/terminal"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	StringType  = reflect.TypeOf("")
	IntType     = reflect.TypeOf(int(0))
	Int64Type   = reflect.TypeOf(int64(0))
	Int32Type   = reflect.TypeOf(int32(0))
	Float64Type = reflect.TypeOf(float64(0))
	Float32Type = reflect.TypeOf(float32(0))
	BoolType    = reflect.TypeOf(false)
	TimeType    = reflect.TypeOf(time.Time{})
)

// Keys converts a map into an array of keys
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MapToString converts a map[string]interface{} into a slice of formatted strings,
// where each entry is in the "key: value" format.
//
// Example:
//
//	input := map[string]interface{}{
//		"key1": "value1",
//		"key2": 123,
//	}
//
//	output := MapToString(input)
//	fmt.Println(output)
//
// Output:
//
//	[]string{"key1: value1", "key2: 123"}
func MapToString[K comparable, V any](m map[K]V) []string {
	lines := make([]string, 0, len(m))
	for k, v := range m {
		lines = append(lines, fmt.Sprintf("%s: %s", k, v))
	}
	return lines
}

// MapValues returns a slice containing all values from the given map.
//
// It accepts a generic map with keys of any comparable type K and values of any type V.
// The returned slice contains the values in an unspecified order.
//
// Example:
//
//	m := map[string]int{"a": 1, "b": 2, "c": 3}
//	values := MapValues(m) // []int{1, 2, 3}
//
// This function is useful when you need to extract values for iteration,
// transformation, or passing to other APIs that expect a slice.
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// ExtractKeyFromLine extracts the key from a "key: value" formatted string.
//
// Example:
//
//	input := "host: localhost"
//	key := ExtractKeyFromLine(input)
//	fmt.Println(key)
//
// Output: "host"
func ExtractKeyFromLine(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func GenerateJsonTag(label string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(label, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return fmt.Sprintf("`json:\"%s\"`", strings.ToLower(snake))
}

// HandleError processes an error with optional context and exits the program.
//
// If the error is a terminal interrupt (Ctrl+C), it prints a cancel message and exits with code 1.
// If a custom message is provided, it prints the message along with the error and exits with code 0.
// Otherwise, it prints a generic internal error message and exits with code 0.
//
// Example:
//
//	err := someOperation()
//	if err != nil {
//	    HandleError(err, "Failed to complete operation")
//	}
//
// Output (if err = errors.New("connection timeout")):
//
//	Failed to complete operation: connection timeout
//	exited wit status 0
//
// Output (if user presses Ctrl+C):
//
//	🛑 Operation cancelled by user (Ctrl+C).
//	exited wit status 0
func HandleError(err error, customMessage ...string) {
	if errors.Is(err, terminal.InterruptErr) {
		fmt.Println("🛑 Operation cancelled by user (Ctrl+C).")
		os.Exit(130)
	}
	if customMessage != nil {
		fmt.Println(wrapError(err, customMessage[0]))
		os.Exit(1)
	} else {
		fmt.Println("Internal error:", err.Error())
		os.Exit(1)
	}
}

func wrapError(err error, customMessage string) string {
	return fmt.Sprintf("%s: %s", customMessage, err.Error())
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func GetProjectName() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}
	return "", nil
}

// SnakeToCamel converts a string from snake case to camel case
func SnakeToCamel(snake string) string {
	parts := strings.Split(snake, "_")
	if len(parts) == 0 {
		return ""
	}

	// Start with the first word as is (lowercase)
	camel := parts[0]

	// Capitalize the first letter of the remaining parts
	for _, part := range parts[1:] {
		if part == "" {
			continue // skip empty parts, in case of double underscores
		}
		camel += strings.Title(part)
	}

	return camel
}

// IsSnakeCase checks if a string is in valid snake_case format
func IsSnakeCase(s string) bool {
	// Matches: lowercase letters, numbers, and underscores,
	// but not starting or ending with an underscore,
	// and no consecutive underscores.
	match, _ := regexp.MatchString(`^[a-z0-9]+(_[a-z0-9]+)*$`, s)
	return match
}

// SnakeToPascal converts snake_case to PascalCase
func SnakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part == "" {
			continue // Skip empty segments (e.g. from double underscores)
		}
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}

// PascalToSnake converts PascalCase to snake_case
func PascalToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			// Add underscore if not the first character
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// PascalToCamel converts a PascalCase string to camelCase.
//
// It lowercases the first character of the input string and returns
// the result. This is useful for generating field names, variables,
// or JSON tags that follow Go's camelCase naming convention.
//
// Examples:
//
//	PascalToCamel("UserRepo")   // "userRepo"
//	PascalToCamel("UUID")       // "uUID"
//	PascalToCamel("Car")        // "car"
//	PascalToCamel("")           // ""
func PascalToCamel(input string) string {
	if input == "" {
		return ""
	}
	return strings.ToLower(input[:1]) + input[1:]
}
