package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

func TestReformatError(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf) // Intercept expected Stderr output
	defer func() {
		log.SetOutput(os.Stderr) // Return to normal Stderr handling after function
	}()

	longString := "Verify that this somewhat long string remains unchanged in the output after being handled"
	err := ReformatError(longString)
	errContainsString := strings.Contains(err.Error(), longString)
	if !errContainsString {
		t.Logf("Test string was not properly included in retured error")
		t.Fail()
	}
}

func TestFindString(t *testing.T) {

	var tests = []struct {
		slice         []string
		val           string
		expectedIndex int
		expectedFound bool
	}{
		{[]string{"the", "and", "for", "so", "go"}, "and", 1, true},
		{[]string{"the", "and", "for", "so", "go"}, "for", 2, true},
		{[]string{"the", "and", "for", "so", "go"}, "in", -1, false},
	}

	for _, c := range tests {

		testName := fmt.Sprintf("FindString(%q,%q) - Expected:%d,%v", c.slice, c.val, c.expectedIndex, c.expectedFound)

		t.Run(testName, func(t *testing.T) {
			actualPosition, actualFound := FindString(c.slice, c.val)

			if actualPosition != c.expectedIndex || actualFound != c.expectedFound {
				t.Errorf("\nCall: FindString(%q,%q)\nResult: %d,%v\nExpected: %d,%v", c.slice, c.val, actualPosition, actualFound, c.expectedIndex, c.expectedFound)
			}
		})
	}
}

func TestReplaceBytesValue(t *testing.T) {

	var tests = []struct {
		bytes          []byte
		oldValue       string
		newValue       string
		expectedResult []byte
	}{
		{[]byte("oldstringhere"), "old", "new", []byte("newstringhere")},                       //Replace a word with no spaces
		{[]byte("oink oink oink"), "k", "ky", []byte("oinky oinky oinky")},                     //Replace a character
		{[]byte("oink oink oink"), "oink", "moo", []byte("moo moo moo")},                       //Replace a word with spaces
		{[]byte("nothing to replace"), "www", "something", []byte("nothing to replace")},       //Nothing to replace due to no match
		{[]byte(""), "a", "b", []byte("")},                                                     //Empty string
		{[]byte("Unicode character: ㄾ"), "Unicode", "Unknown", []byte("Unknown character: ㄾ")}, //String with unicode character
		{[]byte("Unicode character: ㄾ"), "ㄾ", "none", []byte("Unicode character: none")},       //Replace unicode character
	}

	for _, c := range tests {

		testName := fmt.Sprintf("ReplaceBytesValue(%q,%q,%q) - Expected:%q", string(c.bytes), c.oldValue, c.newValue, string(c.expectedResult))

		t.Run(testName, func(t *testing.T) {
			actualResult := ReplaceBytesValue(c.bytes, c.oldValue, c.newValue)

			if string(actualResult) != string(c.expectedResult) {
				t.Errorf("\nCall: ReplaceBytesValue(%q,%q,%q)\nResult: %q\nExpected: %q", string(c.bytes), c.oldValue, c.newValue, string(actualResult), string(c.expectedResult))
			}
		})
	}
}

func TestCallerPath(t *testing.T) {
	type args struct {
		up int
	}
	tests := []struct {
		testName       string
		testArgs       args
		expectedResult string
	}{
		{"CallerPath(%v) - Expected: %q", args{up: 0}, "github.com/citihub/probr/internal/utils.TestCallerPath.func1"},
		{"CallerPath(%v) - Expected: %q", args{up: 1}, "testing.tRunner"},
	}

	for _, tt := range tests {
		tt.testName = fmt.Sprintf(tt.testName, tt.testArgs, tt.expectedResult)
		t.Run(tt.testName, func(t *testing.T) {
			if got := CallerPath(tt.testArgs.up); got != tt.expectedResult {
				t.Errorf("CallerPath(%v) = %v, Expected: %v", tt.testArgs.up, got, tt.expectedResult)
			}
		})
	}
}

func TestCallerName(t *testing.T) {
	type args struct {
		up int
	}
	tests := []struct {
		testName       string
		testArgs       args
		expectedResult string
	}{
		{"CallerName(%v) - Expected: %q", args{up: 0}, "func1"},
		{"CallerName(%v) - Expected: %q", args{up: 1}, "tRunner"},
		{"CallerName(%v) - Expected: %q", args{up: 2}, "goexit"},
	}
	for _, tt := range tests {
		tt.testName = fmt.Sprintf(tt.testName, tt.testArgs, tt.expectedResult)
		t.Run(tt.testName, func(t *testing.T) {
			if got := CallerName(tt.testArgs.up); got != tt.expectedResult {
				t.Errorf("CallerName(%v) = %v, Expected: %v", tt.testArgs.up, got, tt.expectedResult)
			}
		})
	}
}

func TestCallerFileLine(t *testing.T) {
	tests := []struct {
		testName        string
		expectedResult1 string
		expectedResult2 int
	}{
		{"CallerFileLine() - Expected: %q, %d", "c:/go/src/testing/testing.go", 0}, //TODO: Fix - Path for testing.go is local and may break in a diff environment. Get installation path for testing.go tool. Or remove this test if not required.
	}
	for _, tt := range tests {
		tt.testName = fmt.Sprintf(tt.testName, tt.expectedResult1, tt.expectedResult2)
		t.Run(tt.testName, func(t *testing.T) {
			got, _ := CallerFileLine()
			if got != tt.expectedResult1 {
				t.Errorf("CallerFileLine() got = %v, want %v", got, tt.expectedResult1)
			}
			// if got1 != tt.expectedResult2 {
			// 	t.Errorf("CallerFileLine() got1 = %v, want %v", got1, tt.expectedResult2)
			// }
		})
	}
}
