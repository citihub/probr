// Package coreengine contains the types and functions responsible for managing tests and test execution.  This is the primary
// entry point to the core of the application and should be utilised by the probr library to create, execute and report
// on tests.
package coreengine

import (
	"errors"
	"log"
	"sync"

	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/summary"
)

// ProbeStatus type describes the status of the test, e.g. Pending, Running, CompleteSuccess, CompleteFail and Error
type ProbeStatus int

//ProbeStatus enumeration for the ProbeStatus type.
const (
	Pending ProbeStatus = iota
	Running
	CompleteSuccess
	CompleteFail
	Error
	Excluded
)

func (s ProbeStatus) String() string {
	return [...]string{"Pending", "Running", "CompleteSuccess", "CompleteFail", "Error", "Excluded"}[s]
}

// Group type describes the group to which the test belongs, e.g. kubernetes, clouddriver, coreengine, etc.
type Group int

// Group type enumeration
const (
	Kubernetes Group = iota
	CloudDriver
	CoreEngine
)

func (g Group) String() string {
	return [...]string{"kubernetes", "clouddriver", "coreengine"}[g]
}

// ProbeDescriptor describes the specific test case and includes name and group.
type ProbeDescriptor struct {
	Group Group  `json:"group,omitempty"`
	Name  string `json:"name,omitempty"`
}

// ProbeStore maintains a collection of tests to be run and their status.  FailedTests is an explicit
// collection of failed tests.
type ProbeStore struct {
	Tests       map[string]*GodogProbe
	FailedTests map[ProbeStatus]*GodogProbe
	Lock        sync.RWMutex
}

// GetAvailableTests return the collection of available tests.
func GetAvailableTests() *[]ProbeDescriptor {
	//TODO: to implement
	//get this from the ProbeRunner handler store - basically it's the collection of
	//tests that have registered a handler ..

	// return &p
	return nil
}

// NewProbeStore creates a new test manager, backed by ProbeStore
func NewProbeStore() *ProbeStore {
	return &ProbeStore{
		Tests: make(map[string]*GodogProbe),
	}
}

// AddProbe provided GodogProbe to the ProbeStore.
func (ts *ProbeStore) AddProbe(test *GodogProbe) string {
	ts.Lock.Lock()
	defer ts.Lock.Unlock()

	var status ProbeStatus
	if test.ProbeDescriptor.isExcluded() {
		status = Excluded
	} else {
		status = Pending
	}

	//add the test
	test.Status = &status
	ts.Tests[test.ProbeDescriptor.Name] = test

	summary.State.GetProbeLog(test.ProbeDescriptor.Name).Result = test.Status.String()
	summary.State.LogProbeMeta(test.ProbeDescriptor.Name, "group", test.ProbeDescriptor.Group.String())

	return test.ProbeDescriptor.Name
}

// GetProbe returns the test identified by the given name.
func (ts *ProbeStore) GetProbe(name string) (*GodogProbe, error) {
	ts.Lock.Lock()
	defer ts.Lock.Unlock()

	//get the test from the store
	t, exists := ts.Tests[name]

	if !exists {
		return nil, errors.New("test with name '" + name + "' not found")
	}
	return t, nil
}

// ExecProbe executes the test identified by the specified name.
func (ts *ProbeStore) ExecProbe(name string) (int, error) {
	t, err := ts.GetProbe(name)
	if err != nil {
		return 1, err // Failure
	}
	if t.Status.String() != Excluded.String() {
		return ts.RunProbe(t) // Return test results
	}
	return 0, nil // Succeed if test is excluded
}

// ExecAllProbes executes all tests that are present in the ProbeStore.
func (ts *ProbeStore) ExecAllProbes() (int, error) {
	status := 0
	var err error

	for name := range ts.Tests {
		st, err := ts.ExecProbe(name)
		summary.State.ProbeComplete(name)
		if err != nil {
			//log but continue with remaining tests
			log.Printf("[ERROR] error executing test: %v", err)
		}
		if st > status {
			status = st
		}
	}
	return status, err
}

func (td *ProbeDescriptor) isExcluded() bool {
	v := []string{td.Name, td.Group.String()} // iterable name & group strings
	for _, r := range v {
		if tagIsExcluded(r) {
			return true
		}
	}
	return false
}

func tagIsExcluded(tag string) bool {
	for _, exclusion := range config.Vars.TagExclusions {
		if tag == exclusion {
			return true
		}
	}
	return false
}
