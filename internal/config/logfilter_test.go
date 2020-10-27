package config

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	defer func() {
		log.SetOutput(os.Stderr) // Return to normal Stderr handling after function
	}()

	var buf bytes.Buffer
	test_string := "[ERROR] This should log an error"

	log.SetOutput(&buf) // Intercept expected Stderr output
	log.Printf(test_string)
	if len(buf.String()) < len(test_string) {
		t.Logf("Test string was not written to logs as expected")
		t.Fail()
	} else if len(buf.String()) == len(test_string) {
		t.Logf("Logger did not append timestamp to test string as expected")
		t.Fail()
	}
}
