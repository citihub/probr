package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"

	"github.com/citihub/probr"
	"github.com/citihub/probr/cmd/cli_flags"
	"github.com/citihub/probr/internal/clouddriver/kubernetes"
	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/summary"
)

var (
	integrationTest = flag.Bool("integrationTest", false, "run integration tests")
)

//TODO: revise when interface this bit up ...
var kube = kubernetes.GetKubeInstance()

func main() {
	cli_flags.HandleFlags()
	config.LogConfigState()

	if config.Vars.LogLevel == "ERROR" {
		spin := spinner.New(spinner.CharSets[43], 100*time.Millisecond) // Build our new spinner
		spin.Start()                                                    // Start the spinner
	}

	//exec 'em all (for now!)
	s, ts, err := probr.RunAllProbes()
	if err != nil {
		log.Printf("[ERROR] Error executing tests %v", err)
		os.Exit(2) // Error code 1 is reserved for probe test failures, and should not fail in CI
	}
	log.Printf("[NOTICE] Overall test completion status: %v", s)
	summary.State.SetProbrStatus()

	if config.Vars.OutputType == "IO" {
		out, err := probr.GetAllProbeResults(ts)
		if err != nil {
			log.Printf("[ERROR] Experienced error getting test results: %v", s)
			os.Exit(2) // Error code 1 is reserved for probe test failures, and should not fail in CI
		}
		if out == nil || len(out) == 0 {
			log.Printf("[ERROR] Test results not written to file, possibly due to permissions on the specified output directory: %s", config.Vars.CucumberDir)
		}
	}
	summary.State.PrintSummary()

	if config.Vars.LogLevel == "ERROR" {
		spin.Stop()
	}
	os.Exit(s)
}
