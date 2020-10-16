package summary

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/citihub/probr/internal/config"
)

type SummaryState struct {
	Status        string
	EventsPassed  int
	EventsFailed  int
	EventsSkipped int
	PodNames      []string
	Events        map[string]*Event
	EventTags     []config.Event // config.Event contains user-specified tagging options
}

var State SummaryState

func init() {
	State.Events = make(map[string]*Event)
}

// PrintSummary will print the current Events object state, formatted to JSON, if SummaryEnabled is not "false"
func (s *SummaryState) PrintSummary() {
	if config.Vars.SummaryEnabled == "false" {
		log.Printf("[NOTICE] Summary Log suppressed by configuration SummaryEnabled=false.")
	} else {
		summary, _ := json.MarshalIndent(s, "", "  ")
		fmt.Printf("%s", summary) // Summary output should not be handled by log levels
	}
}

// SetProbrStatus evaluates the current SummaryState state to set the Status
func (s *SummaryState) SetProbrStatus() {
	if s.EventsPassed > 0 && s.EventsFailed == 0 {
		s.Status = "Complete - All Events Completed Successfully"
	} else {
		s.Status = fmt.Sprintf("Complete - %v of %v Events Failed", s.EventsFailed, (len(s.Events) - s.EventsSkipped))
	}
	if config.Vars.Events != nil {
		s.EventTags = config.Vars.Events
	}
}

// LogEventMeta accepts a test name with a key and value to insert to the meta logs for that test. Overwrites key if already present.
func (s *SummaryState) LogEventMeta(name string, key string, value string) {
	e := s.GetEventLog(name)
	e.Meta[key] = value
	s.Events[name] = e
	s.Events[name].name = name // Event must be able to access its own name, but it is not publicly printed
}

// EventComplete takes an event name and status then updates the summary & event meta information
func (s *SummaryState) EventComplete(name string) {
	e := s.GetEventLog(name)
	s.completeEvent(e)
	e.audit.Write()
}

// GetEventLog initializes or returns existing log event for the provided test name
func (s *SummaryState) GetEventLog(n string) *Event {
	s.initEvent(n)
	return s.Events[n]
}

func (s *SummaryState) LogPodName(n string) {
	s.PodNames = append(s.PodNames, n)
}

func (s *SummaryState) initEvent(n string) {
	if s.Events[n] == nil {
		ap := filepath.Join(config.Vars.AuditDir, (n + ".json"))
		s.Events[n] = &Event{
			name:          n,
			Meta:          make(map[string]string),
			PodsCreated:   0,
			PodsDestroyed: 0,
			AuditPath:     ap,
			audit: &EventAudit{
				Name: n,
				path: ap,
			},
		}
	}
}

func (s *SummaryState) completeEvent(e *Event) {
	e.countFailures()
	if e.Meta["status"] == "Excluded" {
		e.AuditPath = ""
		s.EventsSkipped = s.EventsSkipped + 1
	} else if len(e.audit.Probes) < 1 {
		e.Meta["status"] = "No Probes Executed"
		e.AuditPath = ""
		s.EventsSkipped = s.EventsSkipped + 1
	} else if e.ProbesFailed < 1 {
		e.Meta["status"] = "Success"
		s.EventsPassed = s.EventsPassed + 1
	} else {
		e.Meta["status"] = "Failed"
		s.EventsFailed = s.EventsFailed + 1
	}
}
