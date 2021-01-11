// Package utils provides general utility methods.  The '*Ptr' functions were borrowed/inspired by the kubernetes go-client.
package utils

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gobuffalo/packr/v2"
)

var boxes map[string]*packr.Box

func init() {
	boxes = make(map[string]*packr.Box)
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to the passed string.
func StringPtr(s string) *string {
	return &s
}

// Int64Ptr returns a pointer to an int64
func Int64Ptr(i int64) *int64 {
	return &i
}

// FindString searches a []string for a specific value.
// If found, returns the index of first occurrence, and True. If not found, returns -1 and False.
func FindString(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// CallerName retrieves the name of the function prior to the location it is called
// If using CallerName(0), the current function's name will be returned
// If using CallerName(1), the current function's parent name will be returned
// If using CallerName(2), the current function's parent's parent name will be returned
func CallerName(up int) string {
	s := strings.Split(CallerPath(up+1), ".") // split full caller path
	return s[len(s)-1]                        // select last element from caller path
}

// CallerPath checks the goroutine's stack of function invocation and returns the following:
// For up=0, return full caller path for caller function
// For up=1, returns full caller path for caller of caller
func CallerPath(up int) string {
	f := make([]uintptr, 1)
	runtime.Callers(up+2, f)                  // add full caller path to empty object
	return runtime.FuncForPC(f[0] - 1).Name() // get full caller path in string form
}

// CallerFileLine returns file name and line of invoker
// Similar to CallerName(1), but with file and line returned
func CallerFileLine() (string, int) {
	_, file, line, _ := runtime.Caller(2)
	return file, line
}

// ReformatError prefixes the error string ready for logging and/or output
func ReformatError(e string, v ...interface{}) error {
	var b strings.Builder
	b.WriteString("[ERROR] ")
	b.WriteString(e)

	s := fmt.Sprintf(b.String(), v...)

	return fmt.Errorf(s)
}

func ReadStaticFile(path ...string) ([]byte, error) {
	filename := path[len(path)-1]
	dirpath := path[0:(len(path) - 1)]
	boxName := strings.Join(dirpath[:], ".")
	if boxes[boxName] == nil {
		boxes[boxName] = BoxStaticFile(boxName, dirpath...) // Name the box after the file being read
	}
	filepath := filepath.Join(boxes[boxName].ResolutionDir, filename)
	return ioutil.ReadFile(filepath)
}

func BoxStaticFile(boxName string, path ...string) *packr.Box {
	return packr.New(boxName, filepath.Join(path...)) // Establish static files for binary build
}

// ReplaceBytesValue replaces a substring with a new value for a given string in bytes
func ReplaceBytesValue(b []byte, old string, new string) []byte {
	newString := strings.Replace(string(b), old, new, -1)
	return []byte(newString)
}
