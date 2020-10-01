package audit

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/citihub/probr/internal/config"
)

type Probe struct {
	Steps map[string]string
}

type Event struct {
	Meta          map[string]string
	PodsCreated   int
	PodsDestroyed int
	ProbesFailed  int
	Probes        map[string]*Probe
}

type AuditLogStruct struct {
	Status        string
	EventsPassed  int
	EventsFailed  int
	EventsSkipped int
	PodNames      []string
	Events        map[string]*Event
}

var AuditLog AuditLogStruct

// PrintAudit will print the current Events object state, formatted to JSON, if AuditEnabled is not "false"
func (a *AuditLogStruct) PrintAudit() {
	if config.Vars.AuditEnabled == "false" {
		log.Printf("[NOTICE] Audit Log suppressed by configuration AuditEnabled=false.")
	} else {
		audit, _ := json.MarshalIndent(a, "", "  ")
		fmt.Printf("%s", audit) // Audit output should not be handled by log levels
	}
}

// SetProbrStatus evaluates the current AuditLogStruct state to set the Status
func (a *AuditLogStruct) SetProbrStatus() {
	if a.EventsPassed > 0 && a.EventsFailed == 0 {
		a.Status = "Complete - All Events Completed Successfully"
	} else {
		a.Status = fmt.Sprintf("Complete - %v of %v Events Failed", a.EventsFailed, (len(a.Events) - a.EventsSkipped))
	}
}

// AuditMeta accepts a test name with a key and value to insert to the meta logs for that test. Overwrites key if already present.
func (a *AuditLogStruct) AuditMeta(name string, key string, value string) {
	e := a.GetEventLog(name)
	e.Meta[key] = value
	a.Events[name] = e
}

// AuditComplete takes an event name and status then updates the audit & event meta information
func (a *AuditLogStruct) AuditComplete(name string, status int) {
	e := a.GetEventLog(name)
	if len(e.Probes) < 1 {
		e.Meta["status"] = "Skipped"
		a.EventsPassed = a.EventsPassed + 1
	} else if status == 0 {
		e.Meta["status"] = "Event Passed"
		a.EventsFailed = a.EventsFailed + 1
	} else {
		e.Meta["status"] = "Event Failed"
		a.EventsSkipped = a.EventsSkipped + 1
	}
}

// GetEventLog initializes or returns existing log event for the provided test name
func (a *AuditLogStruct) GetEventLog(n string) *Event {
	a.logInit(n)
	return a.Events[n]
}

// CountPodCreated increments PodsCreated for event
func (e *Event) CountPodCreated() {
	e.PodsCreated = e.PodsCreated + 1
}

// CountPodDestroyed increments PodsDestroyed for event
func (e *Event) CountPodDestroyed() {
	e.PodsDestroyed = e.PodsDestroyed + 1
}

// AuditProbe
func (e *Event) AuditProbe(name string, key string, err error) {
	if e.Probes == nil {
		e.Probes = make(map[string]*Probe)
	}
	if e.Probes[name] == nil {
		e.Probes[name] = new(Probe)
		e.Probes[name].Steps = make(map[string]string)
	}
	if err == nil {
		e.Probes[name].Steps[key] = "Passed"
	} else {
		e.Probes[name].Steps[key] = "Failed"
		e.ProbesFailed = e.ProbesFailed + 1
	}
}

func (a *AuditLogStruct) AuditPodName(n string) {
	a.PodNames = append(a.PodNames, n)
}

// GetEventLog initializes log event if it doesn't already exist
func (a *AuditLogStruct) logInit(n string) {
	if a.Events == nil {
		a.Events = make(map[string]*Event)
		a.Status = "Running"
	}
	if a.Events[n] == nil {
		a.Events[n] = &Event{
			Meta:          make(map[string]string),
			PodsCreated:   0,
			PodsDestroyed: 0,
		}
	}
}
