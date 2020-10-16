package summary

import (
	"runtime"
	"strings"
)

type Probe struct {
	Status string
	Steps  map[string]string
}

type Event struct {
	name          string
	Meta          map[string]string
	PodsCreated   int
	PodsDestroyed int
	ProbesFailed  int
	Probes        map[string]*Probe
	AuditLocation string
	audit         Audit
}

// prepare initializes any empty objects
func (e *Event) prepare(name string) {
	if e.Probes == nil {
		// This is the first time the event has run
		e.Probes = make(map[string]*Probe)
		e.audit.EventName = e.name
	}
	if e.Probes[name] == nil {
		// This is the first step entry for the specified probe
		e.Probes[name] = new(Probe)
		e.Probes[name].Steps = make(map[string]string)
	}
}

// CountPodCreated increments PodsCreated for event
func (e *Event) CountPodCreated() {
	e.PodsCreated = e.PodsCreated + 1
}

// CountPodDestroyed increments PodsDestroyed for event
func (e *Event) CountPodDestroyed() {
	e.PodsDestroyed = e.PodsDestroyed + 1
}

// LogProbeStep sets pass/fail on probe based on err parameter
func (e *Event) LogProbeStep(name string, err error) {
	e.prepare(name)

	stepName := getCallerName()
	if err == nil {
		e.Probes[name].Steps[stepName] = "Passed"
	} else {
		e.Probes[name].Steps[stepName] = "Failed"
		e.Probes[name].Status = "Failed"
	}
}

// countFailures stores the current total number of failures as e.ProbesFailed. Run at event end
func (e *Event) countFailures() {
	for _, v := range e.Probes {
		if v.Status == "Failed" {
			e.ProbesFailed = e.ProbesFailed + 1
		}
	}
}

// AuditProbe enters probe meta into audit log. Should be added to all scenarios/probes
func (e *Event) AuditProbe(name, result, description string) *ProbeAudit {
	probe := e.audit.Probes[name]
	probe.Result = result
	probe.Description = description
	return probe
}

// AuditProbe enters probe payload into audit log. Not required for all scenarios/probes
func (e *Event) AuditProbePayload(name string) {
	// PENDING IMPLEMENTATION
}

// getCallerName retrieves the name of the function prior to the location it is called
func getCallerName() string {
	f := make([]uintptr, 1)
	runtime.Callers(3, f)                      // add full caller path to empty object
	step := runtime.FuncForPC(f[0] - 1).Name() // get full caller path in string form
	s := strings.Split(step, ".")              // split full caller path
	return s[len(s)-1]                         // select last element from caller path
}
