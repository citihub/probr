package summary

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/citihub/probr/internal/config"
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
	Tags        []string
	Steps       map[string]*StepAudit
}

type StepAudit struct {
	Result      string
	Description string      // Long-form exlanation of anything happening in the step
	Payload     interface{} // Handles any values that are sent across the network
}

func (e *EventAudit) Write() {
	if config.Vars.AuditEnabled == "true" && e.probeRan() {
		_, err := os.Stat(e.path)
		if err == nil && config.Vars.OverwriteHistoricalAudits == "false" {
			// Historical audits are preserved by default
			log.Fatalf("[ERROR] AuditEnabled is set to true, but audit file already exists or Probr could not open: '%s'", e.path)
		}

		json, _ := json.MarshalIndent(e, "", "  ")
		data := []byte(json)
		err = ioutil.WriteFile(e.path, data, 0755)

		if err != nil {
			log.Fatalf("[ERROR] AuditEnabled is set to true, but Probr could not write audit to file: '%s'", e.path)
		}
	}
}

// auditProbeStep sets description, payload, and pass/fail based on err parameter
func (e *EventAudit) auditProbeStep(probeName string, description string, payload interface{}, err error) {
	// Initialize any empty objects
	probe := *e.Probes[probeName]
	// Now do the actual probe summary
	stepName := getCallerName()
	probe.Steps[stepName] = &StepAudit{
		Description: description,
		Payload:     payload,
	}
	if err == nil {
		probe.Steps[stepName].Result = "Passed"
	} else {
		probe.Steps[stepName].Result = "Failed"
		probe.Result = "Failed" // Track this in both summary and audit
	}
}

// getCallerName retrieves the name of the function prior to the location it is called
func getCallerName() string {
	f := make([]uintptr, 1)
	runtime.Callers(4, f)                      // add full caller path to empty object
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
