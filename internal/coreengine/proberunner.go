package coreengine

import (
	"bytes"
	"fmt"

	"github.com/citihub/probr/internal/summary"
	"github.com/cucumber/godog"
)

// ProbeRunner describes the interface that should be implemented to support the execution of tests.
type ProbeRunner interface {
	RunProbe(t *GodogProbe) error
}

// ProbeHandlerFunc describes a callback that should be implemented by test cases in order for ProbeRunner
// to be able to execute the test case.
type ProbeHandlerFunc func(t *GodogProbe) (int, *bytes.Buffer, error)

// GodogProbe encapsulates the specific data that GoDog feature based tests require in order to run.   This
// structure will be passed to the test handler callback.
type GodogProbe struct {
	ProbeDescriptor     *ProbeDescriptor
	ProbeInitializer    func(*godog.TestSuiteContext)
	ScenarioInitializer func(*godog.ScenarioContext)
	FeaturePath         *string
	Status              *ProbeStatus `json:"status,omitempty"`
	Results             *bytes.Buffer
}

// RunProbe runs the test case described by the supplied Test.  It looks in it's test register (the handlers global
// variable) for an entry with the same ProbeDescriptor as the supplied test.  If found, it uses the provided GodogProbe
func (ts *ProbeStore) RunProbe(test *GodogProbe) (int, error) {

	if test == nil {
		summary.State.GetProbeLog(test.ProbeDescriptor.Name).Result = "Internal Error - Test not found"
		return 2, fmt.Errorf("test is nil - cannot run test")
	}

	if test.ProbeDescriptor == nil {
		//update status
		*test.Status = Error
		summary.State.GetProbeLog(test.ProbeDescriptor.Name).Result = "Internal Error - Test descriptor not found"
		return 3, fmt.Errorf("test descriptor is nil - cannot run test")
	}

	s, o, err := GodogProbeHandler(test)

	if s == 0 {
		// success
		*test.Status = CompleteSuccess
	} else {
		// fail
		*test.Status = CompleteFail
	}

	test.Results = o // If in-mem output provided, store as Results
	return s, err
}
