package summary

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/citihub/probr/internal/config"
)

type Audit struct {
	EventName string
	Result    string
	Probes    map[string]*ProbeAudit
}

type ProbeAudit struct {
	Result      string
	Description string
	Steps       []StepAudit
}

type StepAudit struct {
	Name        string
	Result      string
	Description string
	Payload     string
}

func (a *Audit) Write() {
	if config.Vars.AuditEnabled == "true" {
		json, _ := json.MarshalIndent(a, "", "  ")
		data := []byte(json)
		err := ioutil.WriteFile(a.filepath(), data, 0644)
		if err != nil {
			log.Printf("[ERROR] Could not write to file: %s", a.filepath())
		}
	}
}

func (a *Audit) filepath() string {
	filename := a.EventName + "_audit.json"
	return filepath.Join(config.Vars.OutputDir, filename)
}
