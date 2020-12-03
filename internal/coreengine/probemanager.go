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

// ServicePack type describes the group to which the test belongs, e.g. kubernetes, clouddriver, coreengine, etc.
type ServicePack int

// ServicePack type enumeration
const (
	Kubernetes ServicePack = iota
	Storage
)

func (g ServicePack) String() string {
	return [...]string{"kubernetes", "storage"}[g]
}

// ProbeDescriptor describes the specific test case and includes name and group.
type ProbeDescriptor struct {
	ServicePack ServicePack `json:"group,omitempty"`
	Name        string      `json:"name,omitempty"`
}

// ProbeStore maintains a collection of probes to be run and their status.  FailedProbes is an explicit
// collection of failed probes.
type ProbeStore struct {
	Probes       map[string]*GodogProbe
	FailedProbes map[ProbeStatus]*GodogProbe
	Lock         sync.RWMutex
}

// NewProbeStore creates a new object to store GodogProbes
func NewProbeStore() *ProbeStore {
	return &ProbeStore{
		Probes: make(map[string]*GodogProbe),
	}
}

// AddProbe provided GodogProbe to the ProbeStore.
func (ps *ProbeStore) AddProbe(probe *GodogProbe) {
	ps.Lock.Lock()
	defer ps.Lock.Unlock()

	var status ProbeStatus
	if probe.ProbeDescriptor.isExcluded() {
		status = Excluded
	} else {
		status = Pending
	}

	//add the test
	probe.Status = &status
	ps.Probes[probe.ProbeDescriptor.Name] = probe

	summary.State.GetProbeLog(probe.ProbeDescriptor.Name).Result = probe.Status.String()
	summary.State.LogProbeMeta(probe.ProbeDescriptor.Name, "ServicePack", probe.ProbeDescriptor.ServicePack.String())
}

// GetProbe returns the test identified by the given name.
func (ps *ProbeStore) GetProbe(name string) (*GodogProbe, error) {
	ps.Lock.Lock()
	defer ps.Lock.Unlock()

	//get the test from the store
	p, exists := ps.Probes[name]

	if !exists {
		return nil, errors.New("test with name '" + name + "' not found")
	}
	return p, nil
}

// ExecProbe executes the test identified by the specified name.
func (ps *ProbeStore) ExecProbe(name string) (int, error) {
	p, err := ps.GetProbe(name)
	if err != nil {
		return 1, err // Failure
	}
	if p.Status.String() != Excluded.String() {
		return ps.RunProbe(p) // Return test results
	}
	return 0, nil // Succeed if test is excluded
}

// ExecAllProbes executes all tests that are present in the ProbeStore.
func (ps *ProbeStore) ExecAllProbes() (int, error) {
	status := 0
	var err error

	for name := range ps.Probes {
		st, err := ps.ExecProbe(name)
		summary.State.ProbeComplete(name)
		if err != nil {
			//log but continue with remaining probe
			log.Printf("[ERROR] error executing probe: %v", err)
		}
		if st > status {
			status = st
		}
	}
	return status, err
}

func (pd *ProbeDescriptor) isExcluded() bool {
	switch pd.ServicePack.String() {
	case "kubernetes":
		if config.Vars.ServicePacks.Kubernetes.Excluded {
			return true
		} else {
			return probeIsExcluded(pd.Name, config.Vars.ServicePacks.Kubernetes.ProbeExclusions)
		}
	case "storage":
		if config.Vars.ServicePacks.Storage.Excluded {
			return true
		} else {
			return probeIsExcluded(pd.Name, config.Vars.ServicePacks.Storage.ProbeExclusions)
		}
	default:
		log.Printf("[ERROR] unknown servoce pack %s", pd.ServicePack.String())
		return true
	}
}

func probeIsExcluded(name string, probeExclusions []config.ProbeExclusion) bool {
	for _, probeExclusion := range probeExclusions {
		if probeExclusion.Excluded && name == probeExclusion.Name {
			return true
		}
	}
	return false
}
