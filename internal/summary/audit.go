package summary

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"runtime"
	"strings"

	"github.com/citihub/probr/internal/config"
	"github.com/cucumber/messages-go/v10"
)

type EventAudit struct {
	path   string
	Name   string
	Result string
	Probes map[string]*ProbeAudit
}

type ProbeAudit struct {
	Description string
	Result      string
	Steps       map[string]*StepAudit
	Tags        []*messages.Pickle_PickleTag
}

type StepAudit struct {
	Result      string
	Description string
	Payload     string
}

func (e *EventAudit) Write() {
	if config.Vars.AuditEnabled == "true" && e.probeRan() {
		json, _ := json.MarshalIndent(e, "", "  ")
		data := []byte(json)
		err := ioutil.WriteFile(e.path, data, 0755)

		if err != nil {
			log.Printf("[ERROR] Could not write to audit file: %s", e.path)
		}
	}
}

// logProbeStep sets pass/fail on probe based on err parameter
func (e *EventAudit) logProbeStep(name string, err error) {
	// Initialize any empty objects
	if e.Probes == nil {
		e.Probes = make(map[string]*ProbeAudit)
	}
	probe := e.Probes[name]
	if probe == nil {
		probe = &ProbeAudit{
			Steps: make(map[string]*StepAudit),
		}
	}

	// Now do the actual probe summary
	stepName := getCallerName()
	if err == nil {
		probe.Steps[stepName] = &StepAudit{Result: "Passed"}
	} else {
		probe.Steps[stepName] = &StepAudit{Result: "Failed"}
		probe.Result = "Failed"
	}
	e.Probes[name] = probe
}

// getCallerName retrieves the name of the function prior to the location it is called
func getCallerName() string {
	f := make([]uintptr, 1)
	runtime.Callers(3, f)                      // add full caller path to empty object
	step := runtime.FuncForPC(f[0] - 1).Name() // get full caller path in string form
	s := strings.Split(step, ".")              // split full caller path
	return s[len(s)-1]                         // select last element from caller path
}

func (e *EventAudit) probeRan() bool {
	if len(e.Probes) > 0 {
		return true
	}
	return false
}
