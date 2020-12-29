package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"

	"github.com/citihub/probr"
	"github.com/citihub/probr/cmd/cli_flags"
	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/summary"
	"github.com/citihub/probr/service_packs/kubernetes"
)

//TODO: revise when interface this bit up ...
var kube = kubernetes.GetKubeInstance()

func main() {
	err := config.Init("") // Create default config
	if err != nil {
		log.Fatalf("[ERROR] error returned from config.Init: %v", err)
	}

	if len(os.Args[1:]) > 0 {
		cli_flags.HandleRequestForRequiredVars()
		cli_flags.HandleFlags()
	}

	config.LogConfigState()

	if showIndicator() {
		// At this loglevel, Probr is often silent for long periods. Add a visual runtime indicator.
		config.Spinner = spinner.New(spinner.CharSets[42], 500*time.Millisecond)
		config.Spinner.Start()
	}

	//exec 'em all (for now!)
	s, ts, err := probr.RunAllProbes()
	if err != nil {
		log.Printf("[ERROR] Error executing tests %v", err)
		exit(2) // Error code 1 is reserved for probe test failures, and should not fail in CI
	}
	log.Printf("[NOTICE] Overall test completion status: %v", s)
	summary.State.SetProbrStatus()

	if config.Vars.OutputType == "IO" {
		out := probr.GetAllProbeResults(ts)
		if out == nil || len(out) == 0 {
			summary.State.Meta["cucumber directory error"] = fmt.Sprintf(
				"Test results not written to file, possibly due to permissions on the specified output directory: %s",
				config.Vars.CucumberDir,
			)
		}
	}
	summary.State.PrintSummary()
	exit(s)
}

func showIndicator() bool {
	return config.Vars.LogLevel == "ERROR" && !config.Vars.Silent
}

func exit(status int) {
	if showIndicator() {
		config.Spinner.Stop()
	}
	os.Exit(status)
}
