package coreengine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/citihub/probr/internal/config"
)

func TestGetRootDir(t *testing.T) {
	// Make sure it doesn't catch one of the several fail conditions
	_, err := getRootDir()
	if err != nil {
		t.Fail()
	}
}

func TestGetOutputPath(t *testing.T) {
	var file *os.File
	d := "test_output_dir"
	f := "test_file"
	desiredFile := filepath.Join(d, f) + ".json"
	defer func() {
		// Cleanup test assets
		file.Close()
		err := os.RemoveAll(d)
		if err != nil {
			t.Logf("%s", err)
		}

		// Swallow any panics and print a verbose error message
		if err := recover(); err != nil {
			t.Logf("Panicked when trying to create directory or file: '%s'", desiredFile)
			t.Fail()
		}
	}()
	config.Vars.CucumberDir = d
	file, _ = getOutputPath(f)
	if desiredFile != file.Name() {
		t.Logf("Desired filepath '%s' does not match '%s'", desiredFile, file.Name())
		t.Fail()
	}
}

func TestScenarioString(t *testing.T) {
	gs := &godog.Scenario{Name: "test scenario"}

	// Start scenario
	s := scenarioString(true, gs)
	sContainsString := strings.Contains(s, "Start")
	if !sContainsString {
		t.Logf("Test string does not contain 'Start'")
		t.Fail()
	}

	// End scenario
	s = scenarioString(false, gs)
	sContainsString = strings.Contains(s, "End")
	if !sContainsString {
		t.Logf("Test string does not contain 'End'")
		t.Fail()
	}
}

func TestGetFeaturePath(t *testing.T) {
	type args struct {
		path []string
	}
	tests := []struct {
		testName       string
		testArgs       args
		expectedResult string
	}{
		{
			testName:       "GetFeaturePath_WithTwoSubfoldersAndFeatureName_ShouldReturnFeatureFilePath",
			testArgs:       args{path: []string{"service_packs", "kubernetes", "container_registry_access"}},
			expectedResult: filepath.Join("service_packs", "kubernetes", "container_registry_access", "container_registry_access.feature"), // Using filepath.join() instead of literal string in order to run test in Windows (\\) and Linux (/)
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := GetFeaturePath(tt.testArgs.path...); got != tt.expectedResult {
				t.Errorf("GetFeaturePath() = %v, Expected: %v", got, tt.expectedResult)
			}
		})
	}
}
