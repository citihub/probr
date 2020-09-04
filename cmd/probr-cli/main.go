package main

import (
	"flag"
	"log"
	"os"

	"github.com/citihub/probr"
	"github.com/citihub/probr/internal/clouddriver/kubernetes"

	"github.com/citihub/probr/internal/config" //needed for logging
	// _ "github.com/citihub/probr/test/features/clouddriver"
	// _ "github.com/citihub/probr/test/features/kubernetes/containerregistryaccess" //needed to run init on TestHandlers
	// _ "github.com/citihub/probr/test/features/kubernetes/internetaccess"          //needed to run init on TestHandlers
	// _ "github.com/citihub/probr/test/features/kubernetes/podsecuritypolicy"       //needed to run init on TestHandlers
)

var (
	integrationTest = flag.Bool("integrationTest", false, "run integration tests")
)

//TODO: revise when interface this bit up ...
var kube = kubernetes.GetKubeInstance()

func main() {
<<<<<<< HEAD
	var v string
	flag.StringVar(&v, "varsFile", "", "path to config file")
	i := flag.String("kubeConfig", "", "kube config file")
	o := flag.String("outputDir", "", "output directory")
=======
	var v, ot, t, i, o string
	
	flag.StringVar(&v, "varsFile", "", "path to config file")
	flag.StringVar(&ot, "outputType", "INMEM", "output defaults to write in memory, if 'IO' will write to specified output directory")
	flag.StringVar(&t, "tags", "", "test tags, e.g. -tags=\"@CIS-1.2.3, @CIS-4.5.6\".")
	flag.StringVar(&i, "kubeConfig", "", "kube config file")
	flag.StringVar(&o, "outputDir", "", "output directory")
>>>>>>> added tags option to config
	flag.Parse()

	// Will make config.Vars.XYZ available for the rest of the runtime
	err := config.Init(v)
	if err != nil {
		log.Fatalf("[ERROR] Could not create config from provided filepath: %v", err)
	}
	log.Printf("[NOTICE] Probr running with environment: ")
	log.Printf("[NOTICE] %+v", config.Vars)
	if len(i) > 0 {
		config.Vars.SetKubeConfigPath(i)
		log.Printf("[NOTICE] Kube Config has been overridden via command line to: " + i)
	}
	if len(o) > 0 {
		log.Printf("[NOTICE] Output Directory has been overridden via command line to: " + o)
	}
	if ot == "IO" {
		probr.SetIOPaths(i, o)
	}
	if len(t) > 0 {
		config.Vars.SetTags(t)
		log.Printf("[NOTICE] Tags have been added via command line to: " + t)
	}

	//exec 'em all (for now!)
	s, ts, err := probr.RunAllTests()

	if err != nil {
		log.Printf("[ERROR] Error executing tests %v", err)
		os.Exit(2) // Error code 1 is reserved for probe test failures, and should not fail in CI
	}
	log.Printf("[NOTICE] Overall test completion status: %v", s)

	out, err := probr.GetAllTestResults(ts)
	if err != nil {
		log.Printf("[ERROR] Experienced error getting test results: %v", s)
		os.Exit(2) // Error code 1 is reserved for probe test failures, and should not fail in CI
	}
	for k := range out {
		log.Printf("Test results in memory: %v", k)
	}
	os.Exit(s)
}
