package coreengine

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/citihub/probr/internal/summary"
	"github.com/cucumber/godog"
)

// TestRunner describes the interface that should be implemented to support the execution of tests.
type TestRunner interface {
	RunTest(t *Test) error
}

// TestHandlerFunc describes a callback that should be implemented by test cases in order for TestRunner
// to be able to execute the test case.
type TestHandlerFunc func(t *GodogTest) (int, *bytes.Buffer, error)

// GodogTest encapsulates the specific data that GoDog feature based tests require in order to run.   This
// structure will be passed to the test handler callback.
type GodogTest struct {
	TestDescriptor       *TestDescriptor
	TestSuiteInitializer func(*godog.TestSuiteContext)
	ScenarioInitializer  func(*godog.ScenarioContext)
	FeaturePath          *string
}

var (
	handlers    = make(map[string]*GodogTest)
	handlersMux sync.RWMutex
)

// AddTestHandler adds the TestHandlerFunc to the handler map, keyed on the TestDescriptor, and is effectively
// a register of the test cases.  This is the mechanism which links the test case handler to the TestRunner,
// therefore it is essential that the test case register itself with the TestRunner by calling this function
// supplying a description of the test and the GodogTest.  See pod_security_feature.init() for an example.
func AddTestHandler(td TestDescriptor, test *GodogTest) {
	handlersMux.Lock()
	defer handlersMux.Unlock()

	handlers[td.Name] = test
}

// RunTest runs the test case described by the supplied Test.  It looks in it's test register (the handlers global
// variable) for an entry with the same TestDescriptor as the supplied test.  If found, it uses the provided GodogTest
func (ts *TestStore) RunTest(t *Test) (int, error) {
	if t == nil {
		summary.State.GetProbeLog(t.TestDescriptor.Name).Result = "Internal Error - Test not found"
		return 2, fmt.Errorf("test is nil - cannot run test")
	}

	if t.TestDescriptor == nil {
		//update status
		*t.Status = Error
		summary.State.GetProbeLog(t.TestDescriptor.Name).Result = "Internal Error - Test descriptor not found"
		return 3, fmt.Errorf("test descriptor is nil - cannot run test")
	}

	// get the handler (based on the test supplied)
	test := getTest(t.TestDescriptor.Name)

	s, o, err := GodogTestHandler(test)
	if s == 0 {
		// success
		*t.Status = CompleteSuccess
	} else {
		// fail
		*t.Status = CompleteFail

		//TODO: this could be adjusted based on test strictness ...
	}

	t.Results = o // If in-mem output provided, store as Results
	return s, err
}

func getTest(testName string) *GodogTest {
	handlersMux.Lock()
	defer handlersMux.Unlock()
	return handlers[testName]
}
