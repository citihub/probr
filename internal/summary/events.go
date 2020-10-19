package summary

import (
	"github.com/cucumber/messages-go/v10"
)

type Event struct {
	name            string
	audit           *EventAudit
	Meta            map[string]interface{}
	PodsDestroyed   int
	ProbesFailed    int
	ProbesSucceeded int
	Status          string
}

// CountPodCreated increments pods_created for event
func (e *Event) CountPodCreated() {
	e.Meta["pods_created"] = e.Meta["pods_created"].(int) + 1
}

// CountPodDestroyed increments pods_destroyed for event
func (e *Event) CountPodDestroyed() {
	e.Meta["pods_destroyed"] = e.Meta["pods_destroyed"].(int) + 1
}

// countResults stores the current total number of failures as e.ProbesFailed. Run at event end
func (e *Event) countResults() {
	for _, v := range e.audit.Probes {
		if v.Result == "Failed" {
			e.ProbesFailed = e.ProbesFailed + 1
		} else {
			e.ProbesSucceeded = e.ProbesSucceeded + 1
		}
	}
}

func (e *Event) InitializeAuditor(name string, tags []*messages.Pickle_PickleTag) *ProbeAudit {
	if e.audit.Probes == nil {
		e.audit.Probes = make(map[int]*ProbeAudit)
	}
	probeCounter := len(e.audit.Probes) + 1
	var t []string
	for _, tag := range tags {
		t = append(t, tag.Name)
	}
	e.audit.Probes[probeCounter] = &ProbeAudit{
		Name:  name,
		Steps: make(map[int]*StepAudit),
		Tags:  t,
	}
	return e.audit.Probes[probeCounter]
}

// // validateProbeName adds a counter to the end if a probe is run twice under the same name
// func (e *Event) validateProbeName(name string) string {
// 	if e.audit.Probes[name] == nil {
// 		return name
// 	}
// 	var newName string
// 	lastChar := name[len(name)-1]
// 	count, err := strconv.Atoi(string(lastChar))
// 	if err == nil {
// 		newName = name[:len(name)-2] + " " + strconv.Itoa(count+1)
// 	} else {
// 		newName = name + " 2"
// 	}
// 	return e.validateProbeName(newName)
// }
