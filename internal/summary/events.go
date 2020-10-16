package summary

import (
	"github.com/cucumber/messages-go/v10"
)

type Event struct {
	name          string
	audit         *EventAudit
	Meta          map[string]string
	PodsCreated   int
	PodsDestroyed int
	ProbesFailed  int
	AuditPath     string
}

// CountPodCreated increments PodsCreated for event
func (e *Event) CountPodCreated() {
	e.PodsCreated = e.PodsCreated + 1
}

// CountPodDestroyed increments PodsDestroyed for event
func (e *Event) CountPodDestroyed() {
	e.PodsDestroyed = e.PodsDestroyed + 1
}

// countFailures stores the current total number of failures as e.ProbesFailed. Run at event end
func (e *Event) countFailures() {
	for _, v := range e.audit.Probes {
		if v.Result == "Failed" {
			e.ProbesFailed = e.ProbesFailed + 1
		}
	}
}

func (e *Event) LogProbeStep(name string, err error) {
	e.audit.logProbeStep(name, err)
}

func (e *Event) AuditProbe(name string, err error, tags []*messages.Pickle_PickleTag) {
	probe := e.audit.Probes[name]
	if probe != nil {
		probe.Tags = tags

		if err != nil {
			probe.Result = err.Error()
		} else {
			probe.Result = "Success"
		}
	}
}
